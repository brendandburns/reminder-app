package storage

import (
	"os"
	"testing"
	"time"

	"reminder-app/internal/family"
	"reminder-app/internal/reminder"
)

func TestSQLiteStorage(t *testing.T) {
	// Create a temporary database file
	dbFile := "test_reminder.db"
	defer os.Remove(dbFile)

	// Initialize SQLite storage
	storage, err := NewSQLiteStorage(dbFile)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer storage.Close()

	// Use the shared test helper
	runStorageTests(t, storage)
}

func TestSQLiteStorageIDGeneration(t *testing.T) {
	// Create a temporary database file
	dbFile := "test_id_gen.db"
	defer os.Remove(dbFile)

	// Initialize SQLite storage
	storage, err := NewSQLiteStorage(dbFile)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer storage.Close()

	// Test ID generation functions
	familyID1 := GenerateFamilyID(storage)
	familyID2 := GenerateFamilyID(storage)

	if familyID1 == familyID2 {
		t.Error("Generated family IDs should be unique")
	}

	reminderID1 := GenerateReminderID(storage)
	reminderID2 := GenerateReminderID(storage)

	if reminderID1 == reminderID2 {
		t.Error("Generated reminder IDs should be unique")
	}

	eventID1 := GenerateCompletionEventID(storage)
	eventID2 := GenerateCompletionEventID(storage)

	if eventID1 == eventID2 {
		t.Error("Generated completion event IDs should be unique")
	}
}

func TestSQLiteStorageIDPersistence(t *testing.T) {
	// Create a temporary database file
	dbFile := "test_id_persistence.db"
	defer os.Remove(dbFile)

	// Initialize SQLite storage
	storage, err := NewSQLiteStorage(dbFile)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}

	// Generate and create a few families, reminders, and completion events
	fam1 := &family.Family{ID: GenerateFamilyID(storage), Name: "Fam1", Members: []string{"A"}}
	fam2 := &family.Family{ID: GenerateFamilyID(storage), Name: "Fam2", Members: []string{"B"}}
	if err := storage.CreateFamily(fam1); err != nil {
		t.Fatalf("CreateFamily fam1 failed: %v", err)
	}
	if err := storage.CreateFamily(fam2); err != nil {
		t.Fatalf("CreateFamily fam2 failed: %v", err)
	}

	due := time.Now().Add(24 * time.Hour)
	r1 := &reminder.Reminder{
		ID:           GenerateReminderID(storage),
		Title:        "R1",
		FamilyID:     fam1.ID,
		FamilyMember: "A",
		DueDate:      &due,
		Recurrence:   reminder.RecurrencePattern{Type: "once"},
	}
	r2 := &reminder.Reminder{
		ID:           GenerateReminderID(storage),
		Title:        "R2",
		FamilyID:     fam2.ID,
		FamilyMember: "B",
		DueDate:      &due,
		Recurrence:   reminder.RecurrencePattern{Type: "once"},
	}
	if err := storage.CreateReminder(r1); err != nil {
		t.Fatalf("CreateReminder r1 failed: %v", err)
	}
	if err := storage.CreateReminder(r2); err != nil {
		t.Fatalf("CreateReminder r2 failed: %v", err)
	}

	e1 := &reminder.CompletionEvent{ID: GenerateCompletionEventID(storage), ReminderID: r1.ID, CompletedBy: "A", CompletedAt: time.Now()}
	e2 := &reminder.CompletionEvent{ID: GenerateCompletionEventID(storage), ReminderID: r2.ID, CompletedBy: "B", CompletedAt: time.Now()}
	if err := storage.CreateCompletionEvent(e1); err != nil {
		t.Fatalf("CreateCompletionEvent e1 failed: %v", err)
	}
	if err := storage.CreateCompletionEvent(e2); err != nil {
		t.Fatalf("CreateCompletionEvent e2 failed: %v", err)
	}

	// Check counters after creation
	if got, want := storage.GetFamilyIDCounter(), 2; got != want {
		t.Errorf("FamilyIDCounter after create: got %d, want %d", got, want)
	}
	if got, want := storage.GetReminderIDCounter(), 2; got != want {
		t.Errorf("ReminderIDCounter after create: got %d, want %d", got, want)
	}
	if got, want := storage.GetCompletionEventIDCounter(), 2; got != want {
		t.Errorf("CompletionEventIDCounter after create: got %d, want %d", got, want)
	}

	// Close the storage
	storage.Close()

	// Reload storage and check counters are restored
	storage2, err := NewSQLiteStorage(dbFile)
	if err != nil {
		t.Fatalf("Failed to reload SQLite storage: %v", err)
	}
	defer storage2.Close()

	if got, want := storage2.GetFamilyIDCounter(), 2; got != want {
		t.Errorf("FamilyIDCounter after reload: got %d, want %d", got, want)
	}
	if got, want := storage2.GetReminderIDCounter(), 2; got != want {
		t.Errorf("ReminderIDCounter after reload: got %d, want %d", got, want)
	}
	if got, want := storage2.GetCompletionEventIDCounter(), 2; got != want {
		t.Errorf("CompletionEventIDCounter after reload: got %d, want %d", got, want)
	}

	// Generate new IDs and check they increment
	newFamID := GenerateFamilyID(storage2)
	if newFamID != "fam3" {
		t.Errorf("Next family ID after reload: got %s, want fam3", newFamID)
	}
	newRemID := GenerateReminderID(storage2)
	if newRemID != "rem3" {
		t.Errorf("Next reminder ID after reload: got %s, want rem3", newRemID)
	}
	newCevID := GenerateCompletionEventID(storage2)
	if newCevID != "cev3" {
		t.Errorf("Next completion event ID after reload: got %s, want cev3", newCevID)
	}
}

func TestSQLiteStorageCreateTablesError(t *testing.T) {
	// Test with invalid database path to trigger error
	_, err := NewSQLiteStorage("/invalid/path/test.db")
	if err == nil {
		t.Error("Expected error when creating SQLite storage with invalid path")
	}
}

func TestSQLiteStorageTimeHandling(t *testing.T) {
	// Create a temporary database file
	dbFile := "test_time.db"
	defer os.Remove(dbFile)

	// Initialize SQLite storage
	storage, err := NewSQLiteStorage(dbFile)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer storage.Close()

	// Create a family first
	family1 := &family.Family{
		ID:      "fam1",
		Name:    "Test Family",
		Members: []string{"Alice"},
	}
	err = storage.CreateFamily(family1)
	if err != nil {
		t.Fatalf("Failed to create family: %v", err)
	}

	// Test reminder with completed time
	dueDate := time.Now().Add(24 * time.Hour)
	completedAt := time.Now()
	reminder1 := &reminder.Reminder{
		ID:          "rem1",
		Title:       "Completed Reminder",
		Description: "Test Description",
		DueDate:     &dueDate,
		Recurrence: reminder.RecurrencePattern{
			Type: "once",
		},
		Completed:    true,
		CompletedAt:  &completedAt,
		FamilyID:     "fam1",
		FamilyMember: "Alice",
	}

	err = storage.CreateReminder(reminder1)
	if err != nil {
		t.Fatalf("Failed to create reminder with completed time: %v", err)
	}

	// Retrieve and verify the time handling
	retrievedReminder, err := storage.GetReminder("rem1")
	if err != nil {
		t.Fatalf("Failed to get reminder: %v", err)
	}

	if !retrievedReminder.Completed {
		t.Error("Expected reminder to be completed")
	}

	if retrievedReminder.CompletedAt == nil {
		t.Error("Expected completed at time to be set")
	} else {
		// Allow for some time difference due to precision loss in database storage
		timeDiff := retrievedReminder.CompletedAt.Sub(completedAt)
		if timeDiff > time.Second || timeDiff < -time.Second {
			t.Errorf("Time difference too large: %v", timeDiff)
		}
	}

	// Clean up
	storage.DeleteReminder("rem1")
	storage.DeleteFamily("fam1")
}

func TestSQLiteStorageRecurrenceEndDateHandling(t *testing.T) {
	// Create a temporary database file
	dbFile := "test_end_date.db"
	defer os.Remove(dbFile)

	// Initialize SQLite storage
	storage, err := NewSQLiteStorage(dbFile)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer storage.Close()

	// Create a family first
	family1 := &family.Family{
		ID:      "fam1",
		Name:    "Test Family",
		Members: []string{"Alice"},
	}
	err = storage.CreateFamily(family1)
	if err != nil {
		t.Fatalf("Failed to create family: %v", err)
	}

	// Test reminder with empty end date (should be converted to far future)
	dueDate := time.Now().Add(24 * time.Hour)
	reminder1 := &reminder.Reminder{
		ID:          "rem1",
		Title:       "No End Date Reminder",
		Description: "Test Description",
		DueDate:     &dueDate,
		Recurrence: reminder.RecurrencePattern{
			Type:    "weekly",
			Days:    []string{"monday"},
			EndDate: "", // Empty end date
		},
		Completed:    false,
		FamilyID:     "fam1",
		FamilyMember: "Alice",
	}

	err = storage.CreateReminder(reminder1)
	if err != nil {
		t.Fatalf("Failed to create reminder with empty end date: %v", err)
	}

	// Retrieve and verify the end date handling
	retrievedReminder, err := storage.GetReminder("rem1")
	if err != nil {
		t.Fatalf("Failed to get reminder: %v", err)
	}

	// End date should be converted back to empty string
	if retrievedReminder.Recurrence.EndDate != "" {
		t.Errorf("Expected empty end date, got: %s", retrievedReminder.Recurrence.EndDate)
	}

	// Test reminder with actual end date
	reminder2 := &reminder.Reminder{
		ID:          "rem2",
		Title:       "With End Date Reminder",
		Description: "Test Description",
		DueDate:     &dueDate,
		Recurrence: reminder.RecurrencePattern{
			Type:    "weekly",
			Days:    []string{"friday"},
			EndDate: "2025-12-31T23:59:59Z",
		},
		Completed:    false,
		FamilyID:     "fam1",
		FamilyMember: "Alice",
	}

	err = storage.CreateReminder(reminder2)
	if err != nil {
		t.Fatalf("Failed to create reminder with end date: %v", err)
	}

	// Retrieve and verify the end date is preserved
	retrievedReminder2, err := storage.GetReminder("rem2")
	if err != nil {
		t.Fatalf("Failed to get reminder: %v", err)
	}

	if retrievedReminder2.Recurrence.EndDate != "2025-12-31T23:59:59Z" {
		t.Errorf("Expected end date to be preserved, got: %s", retrievedReminder2.Recurrence.EndDate)
	}

	// Clean up
	storage.DeleteReminder("rem1")
	storage.DeleteReminder("rem2")
	storage.DeleteFamily("fam1")
}
