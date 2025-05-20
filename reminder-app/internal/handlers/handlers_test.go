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
	Reminders = make(map[string]*reminder.Reminder) // reset state
	router := setupRouter()
	body := []byte(`{"id":"rem1","title":"Test","description":"desc","due_date":"2025-05-21T10:00:00Z"}`)
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
	if r.ID != "rem1" || r.Title != "Test" {
		t.Errorf("unexpected reminder: %+v", r)
	}
}

func TestGetReminderHandler(t *testing.T) {
	due, _ := time.Parse(time.RFC3339, "2025-05-21T10:00:00Z")
	Reminders = map[string]*reminder.Reminder{"rem2": {ID: "rem2", Title: "T2", Description: "D2", DueDate: due}}
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
	if r.ID != "rem2" {
		t.Errorf("unexpected reminder: %+v", r)
	}
}
