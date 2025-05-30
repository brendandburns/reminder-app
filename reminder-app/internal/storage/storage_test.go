package storage

import (
	"os"
	"reflect"
	"reminder-app/internal/family"
	"reminder-app/internal/reminder"
	"testing"
	"time"
)

func testFamily() *family.Family {
	return &family.Family{
		ID:      "fam1",
		Name:    "Test Family",
		Members: []string{"Alice", "Bob"},
	}
}

func testReminder() *reminder.Reminder {
	due, _ := time.Parse(time.RFC3339, "2025-05-21T10:00:00Z")
	return &reminder.Reminder{
		ID:           "rem1",
		Title:        "Test Reminder",
		Description:  "Test Desc",
		DueDate:      &due,
		FamilyID:     "fam1",
		FamilyMember: "Alice",
		// Explicitly set recurrence to indicate non-recurring
		Recurrence: reminder.RecurrencePattern{
			Type: "once",
		},
	}
}

func testReminderWithNullDueDate() *reminder.Reminder {
	return &reminder.Reminder{
		ID:           "rem2",
		Title:        "Test Reminder No Due Date",
		Description:  "Test Desc No Due Date",
		DueDate:      nil, // Null due date
		FamilyID:     "fam1",
		FamilyMember: "Bob",
		// Explicitly set recurrence to indicate non-recurring
		Recurrence: reminder.RecurrencePattern{
			Type: "once",
		},
	}
}

func runStorageTests(t *testing.T, store Storage) {
	// Family CRUD
	f := testFamily()
	if err := store.CreateFamily(f); err != nil {
		t.Fatalf("CreateFamily failed: %v", err)
	}
	gotFam, err := store.GetFamily(f.ID)
	if err != nil {
		t.Fatalf("GetFamily failed: %v", err)
	}
	if !reflect.DeepEqual(gotFam, f) {
		t.Errorf("GetFamily: got %+v, want %+v", gotFam, f)
	}
	fams, err := store.ListFamilies()
	if err != nil || len(fams) != 1 {
		t.Errorf("ListFamilies: got %d, want 1", len(fams))
	}
	if err := store.DeleteFamily(f.ID); err != nil {
		t.Errorf("DeleteFamily failed: %v", err)
	}
	_, err = store.GetFamily(f.ID)
	if err == nil {
		t.Errorf("expected error after DeleteFamily, got nil")
	}

	// Reminder CRUD and Update
	f = testFamily()
	_ = store.CreateFamily(f)
	r := testReminder()
	if err := store.CreateReminder(r); err != nil {
		t.Fatalf("CreateReminder failed: %v", err)
	}
	gotRem, err := store.GetReminder(r.ID)
	if err != nil {
		t.Fatalf("GetReminder failed: %v", err)
	}
	if gotRem.ID != r.ID || gotRem.Title != r.Title {
		t.Errorf("GetReminder: got %+v, want %+v", gotRem, r)
	}

	// Test updating an existing reminder (upsert functionality)
	originalTitle := r.Title
	r.Title = "Updated Test Reminder"
	r.Description = "Updated Test Description"
	r.Completed = true
	completedTime := time.Now()
	r.CompletedAt = &completedTime
	r.Recurrence.Type = "weekly"
	r.Recurrence.Days = []string{"monday", "wednesday"}

	if err := store.CreateReminder(r); err != nil {
		t.Fatalf("UpdateReminder (via CreateReminder) failed: %v", err)
	}

	// Verify the reminder was updated
	updatedRem, err := store.GetReminder(r.ID)
	if err != nil {
		t.Fatalf("GetReminder after update failed: %v", err)
	}

	if updatedRem.Title != "Updated Test Reminder" {
		t.Errorf("Update failed - Title: got %s, want 'Updated Test Reminder'", updatedRem.Title)
	}
	if updatedRem.Description != "Updated Test Description" {
		t.Errorf("Update failed - Description: got %s, want 'Updated Test Description'", updatedRem.Description)
	}
	if !updatedRem.Completed {
		t.Error("Update failed - Completed should be true")
	}
	if updatedRem.CompletedAt == nil {
		t.Error("Update failed - CompletedAt should not be nil")
	}
	if updatedRem.Recurrence.Type != "weekly" {
		t.Errorf("Update failed - Recurrence type: got %s, want 'weekly'", updatedRem.Recurrence.Type)
	}
	if len(updatedRem.Recurrence.Days) != 2 || updatedRem.Recurrence.Days[0] != "monday" || updatedRem.Recurrence.Days[1] != "wednesday" {
		t.Errorf("Update failed - Recurrence days: got %v, want ['monday', 'wednesday']", updatedRem.Recurrence.Days)
	}

	// Verify we still have only one reminder (not a duplicate)
	rems, err := store.ListReminders()
	if err != nil || len(rems) != 1 {
		t.Errorf("ListReminders after update: got %d, want 1", len(rems))
	}

	if err := store.DeleteReminder(r.ID); err != nil {
		t.Errorf("DeleteReminder failed: %v", err)
	}
	_, err = store.GetReminder(r.ID)
	if err == nil {
		t.Errorf("expected error after DeleteReminder, got nil")
	}

	// CompletionEvent CRUD and Update
	// First recreate the reminder for completion event testing
	r.Title = originalTitle // Reset to original title
	r.Completed = false
	r.CompletedAt = nil
	r.Recurrence.Type = "once"
	r.Recurrence.Days = nil
	if err := store.CreateReminder(r); err != nil {
		t.Fatalf("Recreate reminder for completion event test failed: %v", err)
	}

	e := &reminder.CompletionEvent{
		ID:          "cev1",
		ReminderID:  r.ID,
		CompletedAt: time.Now(),
		CompletedBy: "Alice",
	}
	if err := store.CreateCompletionEvent(e); err != nil {
		t.Fatalf("CreateCompletionEvent failed: %v", err)
	}
	gotEv, err := store.GetCompletionEvent(e.ID)
	if err != nil {
		t.Fatalf("GetCompletionEvent failed: %v", err)
	}
	if gotEv.ID != e.ID || gotEv.ReminderID != r.ID {
		t.Errorf("GetCompletionEvent: got %+v, want %+v", gotEv, e)
	}

	// Test updating an existing completion event (upsert functionality)
	e.CompletedBy = "Bob"
	newCompletedTime := time.Now().Add(time.Hour)
	e.CompletedAt = newCompletedTime

	if err := store.CreateCompletionEvent(e); err != nil {
		t.Fatalf("UpdateCompletionEvent (via CreateCompletionEvent) failed: %v", err)
	}

	// Verify the completion event was updated
	updatedEv, err := store.GetCompletionEvent(e.ID)
	if err != nil {
		t.Fatalf("GetCompletionEvent after update failed: %v", err)
	}

	if updatedEv.CompletedBy != "Bob" {
		t.Errorf("Update failed - CompletedBy: got %s, want 'Bob'", updatedEv.CompletedBy)
	}

	// Allow for some time difference due to precision
	timeDiff := updatedEv.CompletedAt.Sub(newCompletedTime)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("Update failed - CompletedAt time difference too large: %v", timeDiff)
	}

	// Verify we still have only one completion event (not a duplicate)
	evs, err := store.ListCompletionEvents(r.ID)
	if err != nil || len(evs) != 1 {
		t.Errorf("ListCompletionEvents after update: got %d, want 1", len(evs))
	}

	if err := store.DeleteCompletionEvent(e.ID); err != nil {
		t.Errorf("DeleteCompletionEvent failed: %v", err)
	}
	_, err = store.GetCompletionEvent(e.ID)
	if err == nil {
		t.Errorf("expected error after DeleteCompletionEvent, got nil")
	}

	// Test reminder with null due date
	nullDueReminder := testReminderWithNullDueDate()
	if err := store.CreateReminder(nullDueReminder); err != nil {
		t.Fatalf("CreateReminder with null due date failed: %v", err)
	}

	gotNullDueReminder, err := store.GetReminder(nullDueReminder.ID)
	if err != nil {
		t.Fatalf("GetReminder with null due date failed: %v", err)
	}

	if gotNullDueReminder.DueDate != nil {
		t.Errorf("Expected null due date, got %v", gotNullDueReminder.DueDate)
	}

	if gotNullDueReminder.Title != nullDueReminder.Title {
		t.Errorf("Null due date reminder title: got %s, want %s", gotNullDueReminder.Title, nullDueReminder.Title)
	}

	// Verify it appears in the list
	allReminders, err := store.ListReminders()
	if err != nil {
		t.Fatalf("ListReminders failed: %v", err)
	}

	var foundNullDueReminder bool
	for _, rem := range allReminders {
		if rem.ID == nullDueReminder.ID {
			foundNullDueReminder = true
			if rem.DueDate != nil {
				t.Errorf("Listed reminder should have null due date, got %v", rem.DueDate)
			}
			break
		}
	}

	if !foundNullDueReminder {
		t.Error("Null due date reminder not found in list")
	}

	// Clean up null due date reminder
	if err := store.DeleteReminder(nullDueReminder.ID); err != nil {
		t.Errorf("DeleteReminder for null due date failed: %v", err)
	}

	// Clean up the reminder we recreated
	store.DeleteReminder(r.ID)
	store.DeleteFamily(f.ID)
}

func TestMemoryStorage(t *testing.T) {
	store := NewMemoryStorage()
	runStorageTests(t, store)
}

func TestFileStorage(t *testing.T) {
	famFile := "test_families.json"
	remFile := "test_reminders.json"
	completeFile := "test_completion_events.json"
	// Clean up files before and after
	os.Remove(famFile)
	os.Remove(remFile)
	os.Remove(completeFile)
	defer os.Remove(famFile)
	defer os.Remove(remFile)
	defer os.Remove(completeFile)

	store := NewFileStorage(famFile, remFile, completeFile)
	runStorageTests(t, store)
}

func TestFileStorageIDPersistence(t *testing.T) {
	famFile := "test_families_id.json"
	remFile := "test_reminders_id.json"
	completeFile := "test_completion_events_id.json"
	os.Remove(famFile)
	os.Remove(remFile)
	os.Remove(completeFile)
	defer os.Remove(famFile)
	defer os.Remove(remFile)
	defer os.Remove(completeFile)

	store := NewFileStorage(famFile, remFile, completeFile)

	// Generate and create a few families, reminders, and completion events
	fam1 := &family.Family{ID: GenerateFamilyID(store), Name: "Fam1", Members: []string{"A"}}
	fam2 := &family.Family{ID: GenerateFamilyID(store), Name: "Fam2", Members: []string{"B"}}
	if err := store.CreateFamily(fam1); err != nil {
		t.Fatalf("CreateFamily fam1 failed: %v", err)
	}
	if err := store.CreateFamily(fam2); err != nil {
		t.Fatalf("CreateFamily fam2 failed: %v", err)
	}

	due := time.Now().Add(24 * time.Hour)
	r1 := &reminder.Reminder{ID: GenerateReminderID(store), Title: "R1", FamilyID: fam1.ID, FamilyMember: "A", DueDate: &due}
	r2 := &reminder.Reminder{ID: GenerateReminderID(store), Title: "R2", FamilyID: fam2.ID, FamilyMember: "B", DueDate: &due}
	if err := store.CreateReminder(r1); err != nil {
		t.Fatalf("CreateReminder r1 failed: %v", err)
	}
	if err := store.CreateReminder(r2); err != nil {
		t.Fatalf("CreateReminder r2 failed: %v", err)
	}

	e1 := &reminder.CompletionEvent{ID: GenerateCompletionEventID(store), ReminderID: r1.ID, CompletedBy: "A", CompletedAt: time.Now()}
	e2 := &reminder.CompletionEvent{ID: GenerateCompletionEventID(store), ReminderID: r2.ID, CompletedBy: "B", CompletedAt: time.Now()}
	if err := store.CreateCompletionEvent(e1); err != nil {
		t.Fatalf("CreateCompletionEvent e1 failed: %v", err)
	}
	if err := store.CreateCompletionEvent(e2); err != nil {
		t.Fatalf("CreateCompletionEvent e2 failed: %v", err)
	}

	// Check counters after creation
	if got, want := store.GetFamilyIDCounter(), 2; got != want {
		t.Errorf("FamilyIDCounter after create: got %d, want %d", got, want)
	}
	if got, want := store.GetReminderIDCounter(), 2; got != want {
		t.Errorf("ReminderIDCounter after create: got %d, want %d", got, want)
	}
	if got, want := store.GetCompletionEventIDCounter(), 2; got != want {
		t.Errorf("CompletionEventIDCounter after create: got %d, want %d", got, want)
	}

	// Reload storage and check counters are restored
	store2 := NewFileStorage(famFile, remFile, completeFile)
	if got, want := store2.GetFamilyIDCounter(), 2; got != want {
		t.Errorf("FamilyIDCounter after reload: got %d, want %d", got, want)
	}
	if got, want := store2.GetReminderIDCounter(), 2; got != want {
		t.Errorf("ReminderIDCounter after reload: got %d, want %d", got, want)
	}
	if got, want := store2.GetCompletionEventIDCounter(), 2; got != want {
		t.Errorf("CompletionEventIDCounter after reload: got %d, want %d", got, want)
	}

	// Generate new IDs and check they increment
	newFamID := GenerateFamilyID(store2)
	if newFamID != "fam3" {
		t.Errorf("Next family ID after reload: got %s, want fam3", newFamID)
	}
	newRemID := GenerateReminderID(store2)
	if newRemID != "rem3" {
		t.Errorf("Next reminder ID after reload: got %s, want rem3", newRemID)
	}
	newCevID := GenerateCompletionEventID(store2)
	if newCevID != "cev3" {
		t.Errorf("Next completion event ID after reload: got %s, want cev3", newCevID)
	}
}
