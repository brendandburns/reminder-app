package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	fam "reminder-app/internal/family"
	"reminder-app/internal/reminder"
	"reminder-app/internal/storage"

	"github.com/gorilla/mux"
)

var (
	// Remove old maps, use storage instead
	Store storage.Storage
	Mu    sync.Mutex
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
	f.ID = storage.GenerateFamilyID(Store)
	err = Store.CreateFamily(&f)
	Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(f)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusCreated)
}

func GetFamilyHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	f, err := Store.GetFamily(id)
	Mu.Unlock()
	if err != nil {
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
	list, err := Store.ListFamilies()
	Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusOK)
}

func DeleteFamilyHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	err := Store.DeleteFamily(id)
	Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusNoContent)
}

// Reminder Handlers
func CreateReminderHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title        string                     `json:"title"`
		Description  string                     `json:"description"`
		DueDate      string                     `json:"due_date"`
		FamilyID     string                     `json:"family_id"`
		FamilyMember string                     `json:"family_member"`
		Recurrence   reminder.RecurrencePattern `json:"recurrence"`
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
	family, err := Store.GetFamily(req.FamilyID)
	if err != nil {
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
	if req.Recurrence.Type == "" {
		req.Recurrence.Type = "once"
	}

	// Validate recurrence pattern
	switch req.Recurrence.Type {
	case "once":
		// No additional validation needed
	case "weekly":
		if len(req.Recurrence.Days) == 0 {
			Mu.Unlock()
			http.Error(w, "weekly recurrence requires at least one day", http.StatusBadRequest)
			return
		}
		for _, day := range req.Recurrence.Days {
			if !isValidWeekday(day) {
				Mu.Unlock()
				http.Error(w, "invalid weekday in recurrence pattern", http.StatusBadRequest)
				return
			}
		}
	case "monthly":
		if req.Recurrence.Date < 1 || req.Recurrence.Date > 31 {
			Mu.Unlock()
			http.Error(w, "monthly recurrence requires a date between 1 and 31", http.StatusBadRequest)
			return
		}
	default:
		Mu.Unlock()
		http.Error(w, "invalid recurrence type", http.StatusBadRequest)
		return
	}

	if req.Recurrence.EndDate != "" {
		_, err = time.Parse(time.RFC3339, req.Recurrence.EndDate)
		if err != nil {
			Mu.Unlock()
			http.Error(w, "invalid end_date format", http.StatusBadRequest)
			return
		}
	}

	id := storage.GenerateReminderID(Store)
	re := reminder.NewReminder(id, req.Title, req.Description, due, req.FamilyID, req.FamilyMember, req.Recurrence)
	err = Store.CreateReminder(re)
	Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(re)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusCreated)
}

func GetReminderHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	reminder, err := Store.GetReminder(id)
	Mu.Unlock()
	if err != nil {
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
	list, err := Store.ListReminders()
	Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusOK)
}

func DeleteReminderHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	err := Store.DeleteReminder(id)
	Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusNoContent)
}

func UpdateReminderHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	Mu.Lock()
	r, err := Store.GetReminder(id)
	if err != nil {
		Mu.Unlock()
		http.NotFound(w, req)
		log.Printf("%s %s %s %d - Not Found: reminder id '%s' does not exist (update)", req.Method, req.URL.Path, req.UserAgent(), http.StatusNotFound, id)
		return
	}
	// Read and decode partial update
	var patch map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&patch); err != nil {
		Mu.Unlock()
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	updated := false
	for k, v := range patch {
		switch k {
		case "title":
			if s, ok := v.(string); ok {
				r.Title = s
				updated = true
			}
		case "description":
			if s, ok := v.(string); ok {
				r.Description = s
				updated = true
			}
		case "due_date":
			if s, ok := v.(string); ok {
				if t, err := time.Parse(time.RFC3339, s); err == nil {
					r.DueDate = t
					updated = true
				}
			}
		case "completed":
			if b, ok := v.(bool); ok {
				now := time.Now()
				if r.IsRecurring() {
					// For recurring reminders, never set Completed=true, just set CompletedAt
					if b {
						r.CompletedAt = &now
					} else {
						r.CompletedAt = nil
					}
					r.Completed = false
					updated = true
				} else {
					if b && !r.Completed {
						r.MarkCompleted()
						updated = true
					} else if !b && r.Completed {
						r.Completed = false
						r.CompletedAt = nil
						updated = true
					}
				}
				// Create a completion event
				completionEvent := &reminder.CompletionEvent{
					ID:          fmt.Sprintf("cev%d", Store.GetCompletionEventIDCounter()+1),
					ReminderID:  r.ID,
					CompletedBy: r.FamilyMember, // Assuming the assigned member completed it
					CompletedAt: now,
				}

				if err := Store.CreateCompletionEvent(completionEvent); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		case "recurrence":
			if rec, ok := v.(map[string]interface{}); ok {
				var rp reminder.RecurrencePattern
				b, _ := json.Marshal(rec)
				if err := json.Unmarshal(b, &rp); err == nil {
					r.Recurrence = rp
					updated = true
				}
			}
		case "family_member":
			if s, ok := v.(string); ok {
				r.FamilyMember = s
				updated = true
			}
		}
	}

	if updated {
		err = Store.CreateReminder(r) // Overwrite existing
	}
	Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(r)
	log.Printf("%s %s %s %d - PATCH reminder %s", req.Method, req.URL.Path, req.UserAgent(), http.StatusOK, id)
}

// --- CompletionEvent Handlers ---
func CreateCompletionEventHandler(w http.ResponseWriter, r *http.Request) {
	var e reminder.CompletionEvent
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if e.ID == "" {
		Mu.Lock()
		e.ID = storage.GenerateCompletionEventID(Store)
		Mu.Unlock()
	}
	if e.ReminderID == "" || e.CompletedBy == "" {
		http.Error(w, "reminder_id and completed_by are required", http.StatusBadRequest)
		return
	}
	if e.CompletedAt.IsZero() {
		e.CompletedAt = time.Now()
	}
	Mu.Lock()
	err = Store.CreateCompletionEvent(&e)
	Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}

func GetCompletionEventHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	e, err := Store.GetCompletionEvent(id)
	Mu.Unlock()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func ListCompletionEventsHandler(w http.ResponseWriter, r *http.Request) {
	reminderID := mux.Vars(r)["id"]
	if reminderID == "" {
		http.Error(w, "reminder_id query param required", http.StatusBadRequest)
		return
	}
	Mu.Lock()
	list, err := Store.ListCompletionEvents(reminderID)
	Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func DeleteCompletionEventHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	Mu.Lock()
	err := Store.DeleteCompletionEvent(id)
	Mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Helper function to validate weekday strings
func isValidWeekday(day string) bool {
	validDays := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	day = strings.ToLower(day)
	for _, valid := range validDays {
		if day == valid {
			return true
		}
	}
	return false
}
