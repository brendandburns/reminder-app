package storage

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"reminder-app/internal/family"
	"reminder-app/internal/reminder"
)

type FileStorage struct {
	familyFile               string
	reminderFile             string
	completionEventFile      string
	familyIDCounter          int
	reminderIDCounter        int
	completionEventIDCounter int
	mu                       sync.Mutex
}

func NewFileStorage(familyFile, reminderFile, completionFile string) *FileStorage {
	fs := &FileStorage{
		familyFile:          familyFile,
		reminderFile:        reminderFile,
		completionEventFile: completionFile,
	}

	// Initialize counters based on existing data
	fs.recalculateCounters()

	return fs
}

// Helper function to extract numeric ID from string ID
func extractNumericID(id, prefix string) int {
	if strings.HasPrefix(id, prefix) {
		numStr := strings.TrimPrefix(id, prefix)
		if num, err := strconv.Atoi(numStr); err == nil {
			return num
		}
	}
	log.Printf("Invalid ID format: %s, expected prefix: %s", id, prefix)
	return 0
}

// Recalculate all counters based on existing data
func (fs *FileStorage) recalculateCounters() {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Recalculate family counter
	if families, err := fs.loadFamiliesUnsafe(); err == nil {
		maxID := 0
		for id := range families {
			if numID := extractNumericID(id, "fam"); numID > maxID {
				maxID = numID
			}
		}
		fs.familyIDCounter = maxID
	}

	// Recalculate reminder counter
	if reminders, err := fs.loadRemindersUnsafe(); err == nil {
		maxID := 0
		for id := range reminders {
			if numID := extractNumericID(id, "rem"); numID > maxID {
				maxID = numID
			}
		}
		fs.reminderIDCounter = maxID
	}

	// Recalculate completion event counter
	if events, err := fs.loadCompletionEventsUnsafe(); err == nil {
		maxID := 0
		for id := range events {
			if numID := extractNumericID(id, "cev"); numID > maxID {
				maxID = numID
			}
		}
		fs.completionEventIDCounter = maxID
	}
}

// Unsafe versions (without mutex) for internal use
func (fs *FileStorage) loadFamiliesUnsafe() (map[string]*family.Family, error) {
	families := make(map[string]*family.Family)
	if _, err := os.Stat(fs.familyFile); os.IsNotExist(err) {
		return families, nil
	}
	data, err := os.ReadFile(fs.familyFile) // updated
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return families, nil
	}
	if err := json.Unmarshal(data, &families); err != nil {
		return nil, err
	}
	return families, nil
}

func (fs *FileStorage) loadRemindersUnsafe() (map[string]*reminder.Reminder, error) {
	reminders := make(map[string]*reminder.Reminder)
	if _, err := os.Stat(fs.reminderFile); os.IsNotExist(err) {
		return reminders, nil
	}
	data, err := os.ReadFile(fs.reminderFile) // updated
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return reminders, nil
	}
	if err := json.Unmarshal(data, &reminders); err != nil {
		return nil, err
	}
	return reminders, nil
}

func (fs *FileStorage) loadCompletionEventsUnsafe() (map[string]*reminder.CompletionEvent, error) {
	events := make(map[string]*reminder.CompletionEvent)
	if _, err := os.Stat(fs.completionEventFile); os.IsNotExist(err) {
		return events, nil
	}
	data, err := os.ReadFile(fs.completionEventFile) // updated
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return events, nil
	}
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// Helper functions for file IO (thread-safe versions)
func (fs *FileStorage) loadFamilies() (map[string]*family.Family, error) {
	return fs.loadFamiliesUnsafe()
}

func (fs *FileStorage) saveFamilies(families map[string]*family.Family) error {
	data, err := json.MarshalIndent(families, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(fs.familyFile, data, 0644); err != nil { // updated
		return err
	}

	// Update counter after successful save
	maxID := 0
	for id := range families {
		if numID := extractNumericID(id, "fam"); numID > maxID {
			maxID = numID
		}
	}
	fs.familyIDCounter = maxID

	return nil
}

func (fs *FileStorage) loadReminders() (map[string]*reminder.Reminder, error) {
	return fs.loadRemindersUnsafe()
}

func (fs *FileStorage) saveReminders(reminders map[string]*reminder.Reminder) error {
	data, err := json.MarshalIndent(reminders, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(fs.reminderFile, data, 0644); err != nil { // updated
		return err
	}

	// Update counter after successful save
	maxID := 0
	for id := range reminders {
		if numID := extractNumericID(id, "rem"); numID > maxID {
			maxID = numID
		}
	}
	fs.reminderIDCounter = maxID

	return nil
}

func (fs *FileStorage) loadCompletionEvents() (map[string]*reminder.CompletionEvent, error) {
	return fs.loadCompletionEventsUnsafe()
}

func (fs *FileStorage) saveCompletionEvents(events map[string]*reminder.CompletionEvent) error {
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(fs.completionEventFile, data, 0644); err != nil { // updated
		return err
	}

	// Update counter after successful save
	maxID := 0
	for id := range events {
		if numID := extractNumericID(id, "cev"); numID > maxID {
			maxID = numID
		}
	}
	fs.completionEventIDCounter = maxID

	return nil
}

// Family operations
func (fs *FileStorage) CreateFamily(f *family.Family) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	families, err := fs.loadFamilies()
	if err != nil {
		return err
	}
	families[f.ID] = f

	// Update counter if this ID is greater than current
	if numID := extractNumericID(f.ID, "fam"); numID > fs.familyIDCounter {
		fs.familyIDCounter = numID
	}

	return fs.saveFamilies(families)
}

// Reminder operations
func (fs *FileStorage) CreateReminder(r *reminder.Reminder) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	reminders, err := fs.loadReminders()
	if err != nil {
		return err
	}
	reminders[r.ID] = r

	// Update counter if this ID is greater than current
	if numID := extractNumericID(r.ID, "rem"); numID > fs.reminderIDCounter {
		fs.reminderIDCounter = numID
	}

	return fs.saveReminders(reminders)
}

// CompletionEvent operations
func (fs *FileStorage) CreateCompletionEvent(e *reminder.CompletionEvent) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	events, err := fs.loadCompletionEvents()
	if err != nil {
		return err
	}
	events[e.ID] = e

	// Update counter if this ID is greater than current
	if numID := extractNumericID(e.ID, "cev"); numID > fs.completionEventIDCounter {
		fs.completionEventIDCounter = numID
	}

	return fs.saveCompletionEvents(events)
}

func (fs *FileStorage) GetFamily(id string) (*family.Family, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	families, err := fs.loadFamilies()
	if err != nil {
		return nil, err
	}
	f, ok := families[id]
	if !ok {
		return nil, errors.New("family not found")
	}
	return f, nil
}

func (fs *FileStorage) ListFamilies() ([]*family.Family, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	families, err := fs.loadFamilies()
	if err != nil {
		return nil, err
	}
	var list []*family.Family
	for _, f := range families {
		list = append(list, f)
	}
	return list, nil
}

func (fs *FileStorage) DeleteFamily(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	families, err := fs.loadFamilies()
	if err != nil {
		return err
	}
	delete(families, id)
	return fs.saveFamilies(families)
}

func (fs *FileStorage) GetReminder(id string) (*reminder.Reminder, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	reminders, err := fs.loadReminders()
	if err != nil {
		return nil, err
	}
	r, ok := reminders[id]
	if !ok {
		return nil, errors.New("reminder not found")
	}
	return r, nil
}

func (fs *FileStorage) ListReminders() ([]*reminder.Reminder, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	reminders, err := fs.loadReminders()
	if err != nil {
		return nil, err
	}
	var list []*reminder.Reminder
	for _, r := range reminders {
		list = append(list, r)
	}
	return list, nil
}

func (fs *FileStorage) DeleteReminder(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	reminders, err := fs.loadReminders()
	if err != nil {
		return err
	}
	delete(reminders, id)
	return fs.saveReminders(reminders)
}

func (fs *FileStorage) GetCompletionEvent(id string) (*reminder.CompletionEvent, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	events, err := fs.loadCompletionEvents()
	if err != nil {
		return nil, err
	}
	e, ok := events[id]
	if !ok {
		return nil, errors.New("completion event not found")
	}
	return e, nil
}

func (fs *FileStorage) ListCompletionEvents(reminderID string) ([]*reminder.CompletionEvent, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	events, err := fs.loadCompletionEvents()
	if err != nil {
		return nil, err
	}
	var list []*reminder.CompletionEvent
	for _, e := range events {
		if e.ReminderID == reminderID {
			list = append(list, e)
		}
	}
	return list, nil
}

func (fs *FileStorage) DeleteCompletionEvent(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	events, err := fs.loadCompletionEvents()
	if err != nil {
		return err
	}
	delete(events, id)
	return fs.saveCompletionEvents(events)
}

func (fs *FileStorage) GetCompletionEventIDCounter() int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.completionEventIDCounter
}

func (fs *FileStorage) GetFamilyIDCounter() int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.familyIDCounter
}

func (fs *FileStorage) GetReminderIDCounter() int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.reminderIDCounter
}

// Counter setter methods (useful for restoring state or testing)
func (fs *FileStorage) SetFamilyIDCounter(counter int) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.familyIDCounter = counter
	return nil
}

func (fs *FileStorage) SetReminderIDCounter(counter int) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.reminderIDCounter = counter
	return nil
}

func (fs *FileStorage) SetCompletionEventIDCounter(counter int) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.completionEventIDCounter = counter
	return nil
}
