package storage

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"reminder-app/internal/family"
	"reminder-app/internal/reminder"
)

// MongoStorage implements the Storage interface using MongoDB
type MongoStorage struct {
	client                    *mongo.Client
	database                  *mongo.Database
	familyCollection          *mongo.Collection
	reminderCollection        *mongo.Collection
	completionEventCollection *mongo.Collection
	counterCollection         *mongo.Collection
	mu                        sync.Mutex
}

// Counter document structure for ID generation
type Counter struct {
	ID    string `bson:"_id"`
	Value int    `bson:"value"`
}

// NewMongoStorage creates a new MongoDB storage instance
func NewMongoStorage(connectionString, databaseName string) (*MongoStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(databaseName)

	ms := &MongoStorage{
		client:                    client,
		database:                  database,
		familyCollection:          database.Collection("families"),
		reminderCollection:        database.Collection("reminders"),
		completionEventCollection: database.Collection("completion_events"),
		counterCollection:         database.Collection("counters"),
	}

	// Initialize counters if they don't exist
	err = ms.initializeCounters()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize counters: %w", err)
	}

	return ms, nil
}

// Close closes the MongoDB connection
func (ms *MongoStorage) Close(ctx context.Context) error {
	return ms.client.Disconnect(ctx)
}

// initializeCounters initializes the counter documents if they don't exist
func (ms *MongoStorage) initializeCounters() error {
	ctx := context.Background()

	counterTypes := []string{"family", "reminder", "completion_event"}

	for _, counterType := range counterTypes {
		filter := bson.M{"_id": counterType}
		update := bson.M{
			"$setOnInsert": bson.M{
				"_id":   counterType,
				"value": 0,
			},
		}
		opts := options.Update().SetUpsert(true)

		_, err := ms.counterCollection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return fmt.Errorf("failed to initialize %s counter: %w", counterType, err)
		}
	}

	return nil
}

// getNextCounter atomically increments and returns the next counter value
func (ms *MongoStorage) getNextCounter(counterType string) (int, error) {
	ctx := context.Background()

	filter := bson.M{"_id": counterType}
	update := bson.M{"$inc": bson.M{"value": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var counter Counter
	err := ms.counterCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if err != nil {
		return 0, fmt.Errorf("failed to get next counter for %s: %w", counterType, err)
	}

	return counter.Value, nil
}

// setCounter sets the counter value
func (ms *MongoStorage) setCounter(counterType string, value int) error {
	ctx := context.Background()

	filter := bson.M{"_id": counterType}
	update := bson.M{"$set": bson.M{"value": value}}
	opts := options.Update().SetUpsert(true)

	_, err := ms.counterCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to set counter for %s: %w", counterType, err)
	}

	return nil
}

// getCounter gets the current counter value
func (ms *MongoStorage) getCounter(counterType string) (int, error) {
	ctx := context.Background()

	filter := bson.M{"_id": counterType}

	var counter Counter
	err := ms.counterCollection.FindOne(ctx, filter).Decode(&counter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get counter for %s: %w", counterType, err)
	}

	return counter.Value, nil
}

// Family operations

func (ms *MongoStorage) CreateFamily(f *family.Family) error {
	ctx := context.Background()

	_, err := ms.familyCollection.InsertOne(ctx, f)
	if err != nil {
		return fmt.Errorf("failed to create family: %w", err)
	}

	return nil
}

func (ms *MongoStorage) GetFamily(id string) (*family.Family, error) {
	ctx := context.Background()

	filter := bson.M{"id": id}

	var f family.Family
	err := ms.familyCollection.FindOne(ctx, filter).Decode(&f)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("family not found")
		}
		return nil, fmt.Errorf("failed to get family: %w", err)
	}

	return &f, nil
}

func (ms *MongoStorage) ListFamilies() ([]*family.Family, error) {
	ctx := context.Background()

	cursor, err := ms.familyCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to list families: %w", err)
	}
	defer cursor.Close(ctx)

	var families []*family.Family
	for cursor.Next(ctx) {
		var f family.Family
		if err := cursor.Decode(&f); err != nil {
			return nil, fmt.Errorf("failed to decode family: %w", err)
		}
		families = append(families, &f)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return families, nil
}

func (ms *MongoStorage) DeleteFamily(id string) error {
	ctx := context.Background()

	filter := bson.M{"id": id}

	result, err := ms.familyCollection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete family: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("family not found")
	}

	return nil
}

// Reminder operations

func (ms *MongoStorage) CreateReminder(r *reminder.Reminder) error {
	ctx := context.Background()

	_, err := ms.reminderCollection.InsertOne(ctx, r)
	if err != nil {
		return fmt.Errorf("failed to create reminder: %w", err)
	}

	return nil
}

func (ms *MongoStorage) GetReminder(id string) (*reminder.Reminder, error) {
	ctx := context.Background()

	filter := bson.M{"id": id}

	var r reminder.Reminder
	err := ms.reminderCollection.FindOne(ctx, filter).Decode(&r)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("reminder not found")
		}
		return nil, fmt.Errorf("failed to get reminder: %w", err)
	}

	return &r, nil
}

func (ms *MongoStorage) ListReminders() ([]*reminder.Reminder, error) {
	ctx := context.Background()

	cursor, err := ms.reminderCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to list reminders: %w", err)
	}
	defer cursor.Close(ctx)

	var reminders []*reminder.Reminder
	for cursor.Next(ctx) {
		var r reminder.Reminder
		if err := cursor.Decode(&r); err != nil {
			return nil, fmt.Errorf("failed to decode reminder: %w", err)
		}
		reminders = append(reminders, &r)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return reminders, nil
}

func (ms *MongoStorage) DeleteReminder(id string) error {
	ctx := context.Background()

	filter := bson.M{"id": id}

	result, err := ms.reminderCollection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("reminder not found")
	}

	return nil
}

// CompletionEvent operations

func (ms *MongoStorage) CreateCompletionEvent(e *reminder.CompletionEvent) error {
	ctx := context.Background()

	_, err := ms.completionEventCollection.InsertOne(ctx, e)
	if err != nil {
		return fmt.Errorf("failed to create completion event: %w", err)
	}

	return nil
}

func (ms *MongoStorage) GetCompletionEvent(id string) (*reminder.CompletionEvent, error) {
	ctx := context.Background()

	filter := bson.M{"id": id}

	var e reminder.CompletionEvent
	err := ms.completionEventCollection.FindOne(ctx, filter).Decode(&e)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("completion event not found")
		}
		return nil, fmt.Errorf("failed to get completion event: %w", err)
	}

	return &e, nil
}

func (ms *MongoStorage) ListCompletionEvents(reminderID string) ([]*reminder.CompletionEvent, error) {
	ctx := context.Background()

	filter := bson.M{"reminderid": reminderID}

	cursor, err := ms.completionEventCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list completion events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []*reminder.CompletionEvent
	for cursor.Next(ctx) {
		var e reminder.CompletionEvent
		if err := cursor.Decode(&e); err != nil {
			return nil, fmt.Errorf("failed to decode completion event: %w", err)
		}
		events = append(events, &e)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return events, nil
}

func (ms *MongoStorage) DeleteCompletionEvent(id string) error {
	ctx := context.Background()

	filter := bson.M{"id": id}

	result, err := ms.completionEventCollection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete completion event: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("completion event not found")
	}

	return nil
}

// ID counter operations

func (ms *MongoStorage) GetFamilyIDCounter() int {
	counter, err := ms.getCounter("family")
	if err != nil {
		return 0
	}
	return counter
}

func (ms *MongoStorage) SetFamilyIDCounter(counter int) error {
	return ms.setCounter("family", counter)
}

func (ms *MongoStorage) GetReminderIDCounter() int {
	counter, err := ms.getCounter("reminder")
	if err != nil {
		return 0
	}
	return counter
}

func (ms *MongoStorage) SetReminderIDCounter(counter int) error {
	return ms.setCounter("reminder", counter)
}

func (ms *MongoStorage) GetCompletionEventIDCounter() int {
	counter, err := ms.getCounter("completion_event")
	if err != nil {
		return 0
	}
	return counter
}

func (ms *MongoStorage) SetCompletionEventIDCounter(counter int) error {
	return ms.setCounter("completion_event", counter)
}

// Helper functions for MongoDB integration

// GenerateMongoFamilyID generates a new family ID using MongoDB counter
func GenerateMongoFamilyID(ms *MongoStorage) (string, error) {
	counter, err := ms.getNextCounter("family")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("fam%d", counter), nil
}

// GenerateMongoReminderID generates a new reminder ID using MongoDB counter
func GenerateMongoReminderID(ms *MongoStorage) (string, error) {
	counter, err := ms.getNextCounter("reminder")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("rem%d", counter), nil
}

// GenerateMongoCompletionEventID generates a new completion event ID using MongoDB counter
func GenerateMongoCompletionEventID(ms *MongoStorage) (string, error) {
	counter, err := ms.getNextCounter("completion_event")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("cev%d", counter), nil
}

// RecalculateCountersFromData recalculates counters based on existing data in MongoDB
func (ms *MongoStorage) RecalculateCountersFromData() error {
	ctx := context.Background()

	// Recalculate family counter
	familyCount, err := ms.getMaxIDFromCollection(ctx, ms.familyCollection, "id", "fam")
	if err != nil {
		return fmt.Errorf("failed to recalculate family counter: %w", err)
	}
	err = ms.setCounter("family", familyCount)
	if err != nil {
		return fmt.Errorf("failed to set family counter: %w", err)
	}

	// Recalculate reminder counter
	reminderCount, err := ms.getMaxIDFromCollection(ctx, ms.reminderCollection, "id", "rem")
	if err != nil {
		return fmt.Errorf("failed to recalculate reminder counter: %w", err)
	}
	err = ms.setCounter("reminder", reminderCount)
	if err != nil {
		return fmt.Errorf("failed to set reminder counter: %w", err)
	}

	// Recalculate completion event counter
	eventCount, err := ms.getMaxIDFromCollection(ctx, ms.completionEventCollection, "id", "cev")
	if err != nil {
		return fmt.Errorf("failed to recalculate completion event counter: %w", err)
	}
	err = ms.setCounter("completion_event", eventCount)
	if err != nil {
		return fmt.Errorf("failed to set completion event counter: %w", err)
	}

	return nil
}

// getMaxIDFromCollection finds the maximum numeric ID in a collection
func (ms *MongoStorage) getMaxIDFromCollection(ctx context.Context, collection *mongo.Collection, idField, prefix string) (int, error) {
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	maxID := 0
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		if idVal, ok := doc[idField].(string); ok {
			if strings.HasPrefix(idVal, prefix) {
				numStr := strings.TrimPrefix(idVal, prefix)
				if num, err := strconv.Atoi(numStr); err == nil && num > maxID {
					maxID = num
				}
			}
		}
	}

	return maxID, cursor.Err()
}
