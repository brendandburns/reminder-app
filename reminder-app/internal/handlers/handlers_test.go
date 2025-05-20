package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reminder-app/internal/family"
	"reminder-app/internal/reminder"
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
	return r
}

func TestCreateFamilyHandler(t *testing.T) {
	Families = make(map[string]*family.Family) // reset state
	router := setupRouter()
	body := []byte(`{"id":"fam1","name":"Doe","members":["Alice","Bob"]}`)
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
	if f.ID != "fam1" || f.Name != "Doe" || len(f.Members) != 2 {
		t.Errorf("unexpected family.Family: %+v", f)
	}
}

func TestGetFamilyHandler(t *testing.T) {
	Families = map[string]*family.Family{"fam2": {ID: "fam2", Name: "Smith", Members: []string{"Tom"}}}
	router := setupRouter()
	req := httptest.NewRequest("GET", "/families/fam2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	var f family.Family
	if err := json.NewDecoder(resp.Body).Decode(&f); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if f.ID != "fam2" {
		t.Errorf("unexpected family.Family: %+v", f)
	}
}

func TestCreateReminderHandler(t *testing.T) {
	// Reset state
	Reminders = make(map[string]*reminder.Reminder)
	Families = make(map[string]*family.Family)
	reminderIDCounter = 0

	// Create a test family first
	testFamily := &family.Family{
		ID:      "fam1",
		Name:    "Test Family",
		Members: []string{"Alice", "Bob"},
	}
	Families[testFamily.ID] = testFamily

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
		if r.ID != "rem1" || r.Title != "Test" || r.FamilyID != "fam1" || r.FamilyMember != "Alice" {
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
	due, _ := time.Parse(time.RFC3339, "2025-05-21T10:00:00Z")
	Reminders = map[string]*reminder.Reminder{
		"rem2": {
			ID:           "rem2",
			Title:        "T2",
			Description:  "D2",
			DueDate:      due,
			FamilyID:     "fam1",
			FamilyMember: "Alice",
		},
	}
	router := setupRouter()
	req := httptest.NewRequest("GET", "/reminders/rem2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	var r reminder.Reminder
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if r.ID != "rem2" || r.FamilyID != "fam1" || r.FamilyMember != "Alice" {
		t.Errorf("unexpected reminder: %+v", r)
	}
}
