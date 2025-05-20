package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	fam "reminder-app/internal/family"
	rem "reminder-app/internal/reminder"

	"github.com/gorilla/mux"
)

var (
	Families  = make(map[string]*fam.Family)
	Reminders = make(map[string]*rem.Reminder)
	Mu        sync.Mutex
)

// Family Handlers
func CreateFamilyHandler(w http.ResponseWriter, r *http.Request) {
	var f fam.Family
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	Mu.Lock()
	Families[f.ID] = &f
	Mu.Unlock()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(f)
}

func GetFamilyHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	f, ok := Families[id]
	Mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(f)
}

func ListFamiliesHandler(w http.ResponseWriter, r *http.Request) {
	Mu.Lock()
	defer Mu.Unlock()
	var list []*fam.Family
	for _, f := range Families {
		list = append(list, f)
	}
	json.NewEncoder(w).Encode(list)
}

func DeleteFamilyHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	delete(Families, id)
	Mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// Reminder Handlers
func CreateReminderHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		DueDate     string `json:"due_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	due, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		http.Error(w, "invalid due_date format", http.StatusBadRequest)
		return
	}
	reminder := rem.NewReminder(req.ID, req.Title, req.Description, due)
	Mu.Lock()
	Reminders[reminder.ID] = reminder
	Mu.Unlock()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reminder)
}

func GetReminderHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	reminder, ok := Reminders[id]
	Mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(reminder)
}

func ListRemindersHandler(w http.ResponseWriter, r *http.Request) {
	Mu.Lock()
	defer Mu.Unlock()
	var list []*rem.Reminder
	for _, r := range Reminders {
		list = append(list, r)
	}
	json.NewEncoder(w).Encode(list)
}

func DeleteReminderHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	delete(Reminders, id)
	Mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}
