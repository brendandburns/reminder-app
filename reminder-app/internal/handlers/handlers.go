package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	fam "reminder-app/internal/family"
	rem "reminder-app/internal/reminder"

	"github.com/gorilla/mux"
)

var (
	Families          = make(map[string]*fam.Family)
	Reminders         = make(map[string]*rem.Reminder)
	Mu                sync.Mutex
	familyIDCounter   int
	reminderIDCounter int
)

// Family Handlers
func CreateFamilyHandler(w http.ResponseWriter, r *http.Request) {
	var f fam.Family
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		log.Printf("%s %s %s %d - Bad Request: failed to read body: %v", r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, err)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset body for further reading

	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("%s %s %s %d - Bad Request: %v, Body: %s", r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, err, string(body))
		return
	}
	Mu.Lock()
	familyIDCounter++
	f.ID = "fam" + itoa(familyIDCounter)
	Families[f.ID] = &f
	Mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(f)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusCreated)
}

func GetFamilyHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	f, ok := Families[id]
	Mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusOK)
}

func ListFamiliesHandler(w http.ResponseWriter, r *http.Request) {
	Mu.Lock()
	defer Mu.Unlock()
	var list []*fam.Family
	for _, f := range Families {
		list = append(list, f)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusOK)
}

func DeleteFamilyHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	delete(Families, id)
	Mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusNoContent)
}

// Reminder Handlers
func CreateReminderHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		DueDate      string `json:"due_date"`
		FamilyID     string `json:"family_id"`
		FamilyMember string `json:"family_member"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		log.Printf("%s %s %s %d - Bad Request: failed to read body: %v", r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, err)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset body for further reading

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("%s %s %s %d - Bad Request: %v, Body: %s", r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, err, string(body))
		return
	}
	due, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		http.Error(w, "invalid due_date format", http.StatusBadRequest)
		log.Printf("%s %s %s %d - Bad Request: invalid due_date format: %v", r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, err)
		return
	}

	if req.FamilyID == "" || req.FamilyMember == "" {
		http.Error(w, "family_id and family_member are required", http.StatusBadRequest)
		log.Printf("%s %s %s %d - Bad Request: family_id and family_member are required", r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest)
		return
	}

	Mu.Lock()
	family, ok := Families[req.FamilyID]
	if !ok {
		Mu.Unlock()
		http.Error(w, "family not found", http.StatusBadRequest)
		log.Printf("%s %s %s %d - Bad Request: family not found: %s", r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, req.FamilyID)
		return
	}

	memberExists := false
	for _, member := range family.Members {
		if member == req.FamilyMember {
			memberExists = true
			break
		}
	}
	if !memberExists {
		Mu.Unlock()
		http.Error(w, "family member not found", http.StatusBadRequest)
		log.Printf("%s %s %s %d - Bad Request: family member not found: %s", r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, req.FamilyMember)
		return
	}
	reminderIDCounter++
	id := "rem" + itoa(reminderIDCounter)
	reminder := rem.NewReminder(id, req.Title, req.Description, due, req.FamilyID, req.FamilyMember)
	Reminders[reminder.ID] = reminder
	Mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reminder)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusCreated)
}

func GetReminderHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	reminder, ok := Reminders[id]
	Mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		log.Printf("%s %s %s %d - Not Found: reminder id '%s' does not exist", r.Method, r.URL.Path, r.UserAgent(), http.StatusNotFound, id)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reminder)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusOK)
}

func ListRemindersHandler(w http.ResponseWriter, r *http.Request) {
	Mu.Lock()
	defer Mu.Unlock()
	var list []*rem.Reminder
	for _, r := range Reminders {
		list = append(list, r)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusOK)
}

func DeleteReminderHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	_, existed := Reminders[id]
	delete(Reminders, id)
	Mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
	if existed {
		log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusNoContent)
	} else {
		log.Printf("%s %s %s %d - Not Found: reminder id '%s' does not exist (delete)", r.Method, r.URL.Path, r.UserAgent(), http.StatusNoContent, id)
	}
}

// Helper function for int to string
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
