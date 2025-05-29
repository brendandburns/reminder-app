package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"reminder-app/internal/family"
	"reminder-app/internal/reminder"

	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

// skipIfNoDocker skips the test if Docker is not available
func skipIfNoDocker(t *testing.T) {
	// Check if we can run Docker commands
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping Docker-based tests in CI environment")
	}
}

// setupMongoTestContainer sets up a MongoDB test container and returns the storage instance and cleanup function
func setupMongoTestContainer(t *testing.T) (*MongoStorage, func()) {
	skipIfNoDocker(t)

	ctx := context.Background()

	mongoContainer, err := mongodb.RunContainer(ctx)
	if err != nil {
		t.Skipf("Failed to start MongoDB container (Docker may not be available): %v", err)
	}

	connectionString, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		mongoContainer.Terminate(ctx)
		t.Skipf("Failed to get MongoDB connection string: %v", err)
	}

	mongoStorage, err := NewMongoStorage(connectionString, "test_reminder_app")
	if err != nil {
		mongoContainer.Terminate(ctx)
		t.Skipf("Failed to create MongoDB storage: %v", err)
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		mongoStorage.Close(ctx)
		mongoContainer.Terminate(ctx)
	}

	return mongoStorage, cleanup
}

func TestMongoStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MongoDB integration test in short mode")
	}

	mongoStorage, cleanup := setupMongoTestContainer(t)
	defer cleanup()

	// Run the common storage tests
	runStorageTests(t, mongoStorage)
}

func TestMongoStorageIDGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MongoDB integration test in short mode")
	}

	mongoStorage, cleanup := setupMongoTestContainer(t)
	defer cleanup()

	// Test MongoDB-specific ID generation functions
	t.Run("GenerateMongoFamilyID", func(t *testing.T) {
		id1, err := GenerateMongoFamilyID(mongoStorage)
		if err != nil {
			t.Fatalf("GenerateMongoFamilyID failed: %v", err)
		}
		if id1 != "fam1" {
			t.Errorf("Expected first family ID to be 'fam1', got '%s'", id1)
		}

		id2, err := GenerateMongoFamilyID(mongoStorage)
		if err != nil {
			t.Fatalf("GenerateMongoFamilyID failed: %v", err)
		}
		if id2 != "fam2" {
			t.Errorf("Expected second family ID to be 'fam2', got '%s'", id2)
		}
	})

	t.Run("GenerateMongoReminderID", func(t *testing.T) {
		id1, err := GenerateMongoReminderID(mongoStorage)
		if err != nil {
			t.Fatalf("GenerateMongoReminderID failed: %v", err)
		}
		if id1 != "rem1" {
			t.Errorf("Expected first reminder ID to be 'rem1', got '%s'", id1)
		}

		id2, err := GenerateMongoReminderID(mongoStorage)
		if err != nil {
			t.Fatalf("GenerateMongoReminderID failed: %v", err)
		}
		if id2 != "rem2" {
			t.Errorf("Expected second reminder ID to be 'rem2', got '%s'", id2)
		}
	})

	t.Run("GenerateMongoCompletionEventID", func(t *testing.T) {
		id1, err := GenerateMongoCompletionEventID(mongoStorage)
		if err != nil {
			t.Fatalf("GenerateMongoCompletionEventID failed: %v", err)
		}
		if id1 != "cev1" {
			t.Errorf("Expected first completion event ID to be 'cev1', got '%s'", id1)
		}

		id2, err := GenerateMongoCompletionEventID(mongoStorage)
		if err != nil {
			t.Fatalf("GenerateMongoCompletionEventID failed: %v", err)
		}
		if id2 != "cev2" {
			t.Errorf("Expected second completion event ID to be 'cev2', got '%s'", id2)
		}
	})
}

func TestMongoStorageCounterOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MongoDB integration test in short mode")
	}

	mongoStorage, cleanup := setupMongoTestContainer(t)
	defer cleanup()

	t.Run("FamilyIDCounter", func(t *testing.T) {
		// Initial counter should be 0
		counter := mongoStorage.GetFamilyIDCounter()
		if counter != 0 {
			t.Errorf("Expected initial family counter to be 0, got %d", counter)
		}

		// Set counter and verify
		err := mongoStorage.SetFamilyIDCounter(5)
		if err != nil {
			t.Fatalf("SetFamilyIDCounter failed: %v", err)
		}

		counter = mongoStorage.GetFamilyIDCounter()
		if counter != 5 {
			t.Errorf("Expected family counter to be 5, got %d", counter)
		}
	})

	t.Run("ReminderIDCounter", func(t *testing.T) {
		// Initial counter should be 0
		counter := mongoStorage.GetReminderIDCounter()
		if counter != 0 {
			t.Errorf("Expected initial reminder counter to be 0, got %d", counter)
		}

		// Set counter and verify
		err := mongoStorage.SetReminderIDCounter(3)
		if err != nil {
			t.Fatalf("SetReminderIDCounter failed: %v", err)
		}

		counter = mongoStorage.GetReminderIDCounter()
		if counter != 3 {
			t.Errorf("Expected reminder counter to be 3, got %d", counter)
		}
	})

	t.Run("CompletionEventIDCounter", func(t *testing.T) {
		// Initial counter should be 0
		counter := mongoStorage.GetCompletionEventIDCounter()
		if counter != 0 {
			t.Errorf("Expected initial completion event counter to be 0, got %d", counter)
		}

		// Set counter and verify
		err := mongoStorage.SetCompletionEventIDCounter(7)
		if err != nil {
			t.Fatalf("SetCompletionEventIDCounter failed: %v", err)
		}

		counter = mongoStorage.GetCompletionEventIDCounter()
		if counter != 7 {
			t.Errorf("Expected completion event counter to be 7, got %d", counter)
		}
	})
}

func TestMongoStorageRecalculateCounters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MongoDB integration test in short mode")
	}

	mongoStorage, cleanup := setupMongoTestContainer(t)
	defer cleanup()

	// Create some test data
	fam1 := &family.Family{ID: "fam5", Name: "Test Family 1", Members: []string{"Alice"}}
	fam2 := &family.Family{ID: "fam3", Name: "Test Family 2", Members: []string{"Bob"}}
	fam3 := &family.Family{ID: "fam10", Name: "Test Family 3", Members: []string{"Charlie"}}

	err := mongoStorage.CreateFamily(fam1)
	if err != nil {
		t.Fatalf("CreateFamily failed: %v", err)
	}
	err = mongoStorage.CreateFamily(fam2)
	if err != nil {
		t.Fatalf("CreateFamily failed: %v", err)
	}
	err = mongoStorage.CreateFamily(fam3)
	if err != nil {
		t.Fatalf("CreateFamily failed: %v", err)
	}

	due := time.Now().Add(24 * time.Hour)
	rem1 := &reminder.Reminder{ID: "rem2", Title: "Test Reminder 1", FamilyID: fam1.ID, FamilyMember: "Alice", DueDate: due}
	rem2 := &reminder.Reminder{ID: "rem8", Title: "Test Reminder 2", FamilyID: fam2.ID, FamilyMember: "Bob", DueDate: due}

	err = mongoStorage.CreateReminder(rem1)
	if err != nil {
		t.Fatalf("CreateReminder failed: %v", err)
	}
	err = mongoStorage.CreateReminder(rem2)
	if err != nil {
		t.Fatalf("CreateReminder failed: %v", err)
	}

	cev1 := &reminder.CompletionEvent{ID: "cev4", ReminderID: rem1.ID, CompletedBy: "Alice", CompletedAt: time.Now()}
	cev2 := &reminder.CompletionEvent{ID: "cev1", ReminderID: rem2.ID, CompletedBy: "Bob", CompletedAt: time.Now()}
	cev3 := &reminder.CompletionEvent{ID: "cev12", ReminderID: rem1.ID, CompletedBy: "Alice", CompletedAt: time.Now()}

	err = mongoStorage.CreateCompletionEvent(cev1)
	if err != nil {
		t.Fatalf("CreateCompletionEvent failed: %v", err)
	}
	err = mongoStorage.CreateCompletionEvent(cev2)
	if err != nil {
		t.Fatalf("CreateCompletionEvent failed: %v", err)
	}
	err = mongoStorage.CreateCompletionEvent(cev3)
	if err != nil {
		t.Fatalf("CreateCompletionEvent failed: %v", err)
	}

	// Recalculate counters based on existing data
	err = mongoStorage.RecalculateCountersFromData()
	if err != nil {
		t.Fatalf("RecalculateCountersFromData failed: %v", err)
	}

	// Check that counters are set to the maximum ID values
	familyCounter := mongoStorage.GetFamilyIDCounter()
	if familyCounter != 10 {
		t.Errorf("Expected family counter to be 10, got %d", familyCounter)
	}

	reminderCounter := mongoStorage.GetReminderIDCounter()
	if reminderCounter != 8 {
		t.Errorf("Expected reminder counter to be 8, got %d", reminderCounter)
	}

	completionEventCounter := mongoStorage.GetCompletionEventIDCounter()
	if completionEventCounter != 12 {
		t.Errorf("Expected completion event counter to be 12, got %d", completionEventCounter)
	}
}

func TestMongoStorageQueryOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MongoDB integration test in short mode")
	}

	mongoStorage, cleanup := setupMongoTestContainer(t)
	defer cleanup()

	// Test CompletionEvent queries by reminder ID
	t.Run("ListCompletionEventsByReminderID", func(t *testing.T) {
		// Create test data
		due := time.Now().Add(24 * time.Hour)
		rem1 := &reminder.Reminder{ID: "rem1", Title: "Test Reminder 1", FamilyID: "fam1", FamilyMember: "Alice", DueDate: due}
		rem2 := &reminder.Reminder{ID: "rem2", Title: "Test Reminder 2", FamilyID: "fam1", FamilyMember: "Bob", DueDate: due}

		err := mongoStorage.CreateReminder(rem1)
		if err != nil {
			t.Fatalf("CreateReminder failed: %v", err)
		}
		err = mongoStorage.CreateReminder(rem2)
		if err != nil {
			t.Fatalf("CreateReminder failed: %v", err)
		}

		// Create completion events for rem1
		cev1 := &reminder.CompletionEvent{ID: "cev1", ReminderID: rem1.ID, CompletedBy: "Alice", CompletedAt: time.Now()}
		cev2 := &reminder.CompletionEvent{ID: "cev2", ReminderID: rem1.ID, CompletedBy: "Alice", CompletedAt: time.Now().Add(1 * time.Hour)}
		cev3 := &reminder.CompletionEvent{ID: "cev3", ReminderID: rem2.ID, CompletedBy: "Bob", CompletedAt: time.Now()}

		err = mongoStorage.CreateCompletionEvent(cev1)
		if err != nil {
			t.Fatalf("CreateCompletionEvent failed: %v", err)
		}
		err = mongoStorage.CreateCompletionEvent(cev2)
		if err != nil {
			t.Fatalf("CreateCompletionEvent failed: %v", err)
		}
		err = mongoStorage.CreateCompletionEvent(cev3)
		if err != nil {
			t.Fatalf("CreateCompletionEvent failed: %v", err)
		}

		// Query completion events for rem1 - should return 2 events
		events, err := mongoStorage.ListCompletionEvents(rem1.ID)
		if err != nil {
			t.Fatalf("ListCompletionEvents failed: %v", err)
		}

		if len(events) != 2 {
			t.Errorf("Expected 2 completion events for rem1, got %d", len(events))
		}

		// Query completion events for rem2 - should return 1 event
		events, err = mongoStorage.ListCompletionEvents(rem2.ID)
		if err != nil {
			t.Fatalf("ListCompletionEvents failed: %v", err)
		}

		if len(events) != 1 {
			t.Errorf("Expected 1 completion event for rem2, got %d", len(events))
		}

		// Query completion events for non-existent reminder - should return empty slice
		events, err = mongoStorage.ListCompletionEvents("rem999")
		if err != nil {
			t.Fatalf("ListCompletionEvents failed: %v", err)
		}

		if len(events) != 0 {
			t.Errorf("Expected 0 completion events for rem999, got %d", len(events))
		}
	})
}

func TestMongoStorageErrorCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MongoDB integration test in short mode")
	}

	mongoStorage, cleanup := setupMongoTestContainer(t)
	defer cleanup()

	t.Run("GetNonExistentFamily", func(t *testing.T) {
		_, err := mongoStorage.GetFamily("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent family, got nil")
		}
	})

	t.Run("GetNonExistentReminder", func(t *testing.T) {
		_, err := mongoStorage.GetReminder("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent reminder, got nil")
		}
	})

	t.Run("GetNonExistentCompletionEvent", func(t *testing.T) {
		_, err := mongoStorage.GetCompletionEvent("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent completion event, got nil")
		}
	})

	t.Run("DeleteNonExistentFamily", func(t *testing.T) {
		err := mongoStorage.DeleteFamily("nonexistent")
		if err == nil {
			t.Error("Expected error when deleting non-existent family, got nil")
		}
	})

	t.Run("DeleteNonExistentReminder", func(t *testing.T) {
		err := mongoStorage.DeleteReminder("nonexistent")
		if err == nil {
			t.Error("Expected error when deleting non-existent reminder, got nil")
		}
	})

	t.Run("DeleteNonExistentCompletionEvent", func(t *testing.T) {
		err := mongoStorage.DeleteCompletionEvent("nonexistent")
		if err == nil {
			t.Error("Expected error when deleting non-existent completion event, got nil")
		}
	})
}

// TestMongoStorageConnectionError tests behavior when MongoDB is not available
func TestMongoStorageConnectionError(t *testing.T) {
	_, err := NewMongoStorage("mongodb://nonexistent:27017", "test_db")
	if err == nil {
		t.Error("Expected error when connecting to non-existent MongoDB, got nil")
	}
}
