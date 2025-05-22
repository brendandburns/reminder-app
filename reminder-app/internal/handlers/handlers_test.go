package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reminder-app/internal/family"
	"reminder-app/internal/reminder"
	"reminder-app/internal/storage"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func setupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/families", CreateFamilyHandler).Methods("POST")
	r.HandleFunc("/families", ListFamiliesHandler).Methods("GET")
	r.HandleFunc("/families/{id}", GetFamilyHandler).Methods("GET")
	r.HandleFunc("/families/{id}", DeleteFamilyHandler).Methods("DELETE")
	r.HandleFunc("/reminders", CreateReminderHandler).Methods("POST")
	r.HandleFunc("/reminders", ListRemindersHandler).Methods("GET")
	r.HandleFunc("/reminders/{id}", GetReminderHandler).Methods("GET")
	r.HandleFunc("/reminders/{id}", DeleteReminderHandler).Methods("DELETE")
	r.HandleFunc("/reminders/{id}", UpdateReminderHandler).Methods("PATCH") // Add PATCH route for testing

	// Add new completion event routes
	r.HandleFunc("/completion-events", CreateCompletionEventHandler).Methods("POST")
	r.HandleFunc("/completion-events/{id}", GetCompletionEventHandler).Methods("GET")
	r.HandleFunc("/completion-events/{id}", DeleteCompletionEventHandler).Methods("DELETE")
	r.HandleFunc("/reminders/{id}/completion-events", ListCompletionEventsHandler).Methods("GET")

	return r
}

func setupTestStorage() {
	Store = storage.NewMemoryStorage()
	familyIDCounter = 0
	reminderIDCounter = 0
}

func TestCreateFamilyHandler(t *testing.T) {
	setupTestStorage()
	router := setupRouter()
	body := []byte(`{"name":"Doe","members":["Alice","Bob"]}`)
	req := httptest.NewRequest("POST", "/families", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	var f family.Family
	if err := json.NewDecoder(resp.Body).Decode(&f); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if f.Name != "Doe" || len(f.Members) != 2 {
		t.Errorf("unexpected family.Family: %+v", f)
	}
}

func TestGetFamilyHandler(t *testing.T) {
	setupTestStorage()
	// Create a family in storage
	f := &family.Family{ID: "fam2", Name: "Smith", Members: []string{"Tom"}}
	_ = Store.CreateFamily(f)
	router := setupRouter()
	req := httptest.NewRequest("GET", "/families/fam2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	var got family.Family
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got.ID != "fam2" {
		t.Errorf("unexpected family.Family: %+v", got)
	}
}

func TestCreateReminderHandler(t *testing.T) {
	setupTestStorage()
	// Create a test family first
	testFamily := &family.Family{
		ID:      "fam1",
		Name:    "Test Family",
		Members: []string{"Alice", "Bob"},
	}
	_ = Store.CreateFamily(testFamily)

	router := setupRouter()

	t.Run("Successful reminder creation", func(t *testing.T) {
		body := []byte(`{
			"title": "Test",
			"description": "Test reminder",
			"due_date": "2024-01-01T10:00:00Z",
			"family_id": "fam1",
			"family_member": "Alice"
		}`)
		req := httptest.NewRequest("POST", "/reminders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", resp.StatusCode)
		}
		var r reminder.Reminder
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if r.Title != "Test" || r.FamilyID != "fam1" || r.FamilyMember != "Alice" {
			t.Errorf("unexpected reminder: %+v", r)
		}
	})

	t.Run("Invalid family ID", func(t *testing.T) {
		body := []byte(`{
			"title": "Test",
			"description": "Test reminder",
			"due_date": "2024-01-01T10:00:00Z",
			"family_id": "invalid",
			"family_member": "Alice"
		}`)
		req := httptest.NewRequest("POST", "/reminders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("Invalid family member", func(t *testing.T) {
		body := []byte(`{
			"title": "Test",
			"description": "Test reminder",
			"due_date": "2024-01-01T10:00:00Z",
			"family_id": "fam1",
			"family_member": "Charlie"
		}`)
		req := httptest.NewRequest("POST", "/reminders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", resp.StatusCode)
		}
	})
}

func TestGetReminderHandler(t *testing.T) {
	setupTestStorage()
	due, _ := time.Parse(time.RFC3339, "2025-05-21T10:00:00Z")
	// Create a family and reminder in storage
	f := &family.Family{ID: "fam1", Name: "Smith", Members: []string{"Alice"}}
	_ = Store.CreateFamily(f)
	r := &reminder.Reminder{
		ID:           "rem2",
		Title:        "T2",
		Description:  "D2",
		DueDate:      due,
		FamilyID:     "fam1",
		FamilyMember: "Alice",
	}
	_ = Store.CreateReminder(r)
	router := setupRouter()
	req := httptest.NewRequest("GET", "/reminders/rem2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	var got reminder.Reminder
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got.ID != "rem2" || got.FamilyID != "fam1" || got.FamilyMember != "Alice" {
		t.Errorf("unexpected reminder: %+v", got)
	}
}

func TestUpdateReminderHandler(t *testing.T) {
	setupTestStorage()
	// Create a family and reminder in storage
	f := &family.Family{ID: "fam1", Name: "Smith", Members: []string{"Alice"}}
	_ = Store.CreateFamily(f)
	due, _ := time.Parse(time.RFC3339, "2025-05-21T10:00:00Z")
	r := &reminder.Reminder{
		ID:           "rem1",
		Title:        "Old Title",
		Description:  "Old Desc",
		DueDate:      due,
		FamilyID:     "fam1",
		FamilyMember: "Alice",
		Recurrence: reminder.RecurrencePattern{
			Type: "once",
		},
	}
	_ = Store.CreateReminder(r)
	router := setupRouter()

	t.Run("Patch title and mark completed", func(t *testing.T) {
		patch := map[string]interface{}{
			"title":     "New Title",
			"completed": true,
		}
		body, _ := json.Marshal(patch)
		req := httptest.NewRequest("PATCH", "/reminders/rem1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
		var updated reminder.Reminder
		if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if updated.Title != "New Title" {
			t.Errorf("expected title 'New Title', got '%s'", updated.Title)
		}
		if !updated.Completed || updated.CompletedAt == nil {
			t.Errorf("expected reminder to be completed with completion time, got %+v", updated)
		}
	})

	t.Run("Patch due_date", func(t *testing.T) {
		newDue := "2026-01-01T12:00:00Z"
		patch := map[string]interface{}{
			"due_date": newDue,
		}
		body, _ := json.Marshal(patch)
		req := httptest.NewRequest("PATCH", "/reminders/rem1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
		var updated reminder.Reminder
		if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if updated.DueDate.Format(time.RFC3339) != newDue {
			t.Errorf("expected due_date '%s', got '%s'", newDue, updated.DueDate.Format(time.RFC3339))
		}
	})

	t.Run("Complete recurring reminder", func(t *testing.T) {
		// Create a recurring reminder first
		recurringReminder := &reminder.Reminder{
			ID:           "rem2",
			Title:        "Recurring Task",
			Description:  "Daily task",
			DueDate:      due,
			FamilyID:     "fam1",
			FamilyMember: "Alice",
			Recurrence: reminder.RecurrencePattern{
				Type: "daily",
				Days: []string{"monday", "tuesday"},
			},
		}
		_ = Store.CreateReminder(recurringReminder)

		// Mark it as completed
		patch := map[string]interface{}{
			"completed": true,
		}
		body, _ := json.Marshal(patch)
		req := httptest.NewRequest("PATCH", "/reminders/rem2", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var updated reminder.Reminder
		if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
			t.Fatalf("decode error: %v", err)
		}

		// For recurring reminders:
		// - CompletedAt should be set to current time
		// - Completed flag should remain false
		if updated.Completed {
			t.Error("recurring reminder should not be marked as completed")
		}
		if updated.CompletedAt == nil {
			t.Error("CompletedAt should be set for recurring reminder")
		}
	})
}

func TestCompletionEventHandlers(t *testing.T) {
	setupTestStorage()
	// Create required test data first
	f := &family.Family{ID: "fam1", Name: "Smith", Members: []string{"Alice"}}
	_ = Store.CreateFamily(f)
	due, _ := time.Parse(time.RFC3339, "2025-05-21T10:00:00Z")
	r := &reminder.Reminder{
		ID:           "rem1",
		Title:        "Test Reminder",
		DueDate:      due,
		FamilyID:     "fam1",
		FamilyMember: "Alice",
	}
	_ = Store.CreateReminder(r)
	router := setupRouter()

	t.Run("Create completion event", func(t *testing.T) {
		completion := map[string]interface{}{
			"reminder_id":  "rem1",
			"completed_by": "Alice",
		}
		body, _ := json.Marshal(completion)
		req := httptest.NewRequest("POST", "/completion-events", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", resp.StatusCode)
		}

		var created reminder.CompletionEvent
		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if created.ReminderID != "rem1" || created.CompletedBy != "Alice" {
			t.Errorf("unexpected completion event: %+v", created)
		}
		if created.CompletedAt.IsZero() {
			t.Error("CompletedAt should be set")
		}

		// Get the event from storage to verify that it was created
		storedEvent, err := Store.GetCompletionEvent(created.ID)
		if err != nil {
			t.Fatalf("failed to get completion event from storage: %v", err)
		}
		if storedEvent.ID != created.ID || storedEvent.ReminderID != "rem1" {
			t.Errorf("unexpected stored completion event: %+v", storedEvent)
		}
		if storedEvent.CompletedBy != "Alice" {
			t.Errorf("expected completed_by 'Alice', got '%s'", storedEvent.CompletedBy)
		}
		if storedEvent.CompletedAt.IsZero() {
			t.Error("CompletedAt should be set in stored event")
		}
	})	

	t.Run("Get completion event", func(t *testing.T) {
		// First create an event
		event := &reminder.CompletionEvent{
			ID:          "cev1",
			ReminderID:  "rem1",
			CompletedBy: "Alice",
			CompletedAt: time.Now(),
		}
		_ = Store.CreateCompletionEvent(event)

		req := httptest.NewRequest("GET", "/completion-events/cev1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var got reminder.CompletionEvent
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if got.ID != "cev1" || got.ReminderID != "rem1" {
			t.Errorf("unexpected completion event: %+v", got)
		}
	})

	t.Run("List completion events for reminder", func(t *testing.T) {
		// Create a reminder first
		testReminder := &reminder.Reminder{
			ID:           "rem3",
			Title:        "Test Reminder",
			Description:  "Test Description",
			DueDate:      due,
			FamilyID:     "fam1",
			FamilyMember: "Alice",
		}
		_ = Store.CreateReminder(testReminder)

		// Create a couple of completion events for this reminder
		event1 := &reminder.CompletionEvent{
			ID:          "cev2",
			ReminderID:  "rem3",
			CompletedBy: "Alice",
			CompletedAt: time.Now(),
		}
		event2 := &reminder.CompletionEvent{
			ID:          "cev3",
			ReminderID:  "rem3",
			CompletedBy: "Alice",
			CompletedAt: time.Now().Add(-24 * time.Hour), // completed yesterday
		}
		_ = Store.CreateCompletionEvent(event1)
		_ = Store.CreateCompletionEvent(event2)

		req := httptest.NewRequest("GET", "/reminders/rem3/completion-events", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var events []reminder.CompletionEvent
		if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("expected 2 completion events, got %d", len(events))
		}

		// Verify the events are for the correct reminder
		for _, e := range events {
			if e.ReminderID != "rem3" {
				t.Errorf("got event for reminder %s, want rem3", e.ReminderID)
			}
		}
	})

	t.Run("Delete completion event", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/completion-events/cev1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("expected status 204, got %d", resp.StatusCode)
		}

		// Verify deletion
		req = httptest.NewRequest("GET", "/completion-events/cev1", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Result().StatusCode != http.StatusNotFound {
			t.Error("completion event should not exist after deletion")
		}
	})
}
