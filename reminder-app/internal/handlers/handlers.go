package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	fam "reminder-app/internal/family"
	"reminder-app/internal/reminder"
	"reminder-app/internal/storage"

	"github.com/gorilla/mux"
)

var (
	// Remove old maps, use storage instead
	Store storage.Storage
)

// errorHandler provides consistent error handling and logging
func errorHandler(w http.ResponseWriter, r *http.Request, message string, statusCode int, err error) {
	if err != nil {
		log.Printf("%s %s %s %d - %s: %v", r.Method, r.URL.Path, r.UserAgent(), statusCode, message, err)
	} else {
		log.Printf("%s %s %s %d - %s", r.Method, r.URL.Path, r.UserAgent(), statusCode, message)
	}
	http.Error(w, message, statusCode)
}

// Family Handlers
func CreateFamilyHandler(w http.ResponseWriter, r *http.Request) {
	var f fam.Family
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errorHandler(w, r, "failed to read request body", http.StatusBadRequest, err)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset body for further reading

	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		errorHandler(w, r, fmt.Sprintf("invalid JSON: %v, Body: %s", err, string(body)), http.StatusBadRequest, err)
		return
	}
	f.ID = storage.GenerateFamilyID(Store)
	err = Store.CreateFamily(&f)
	if err != nil {
		errorHandler(w, r, "failed to create family", http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(f)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusCreated)
}

func GetFamilyHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	f, err := Store.GetFamily(id)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("family not found: %s", id), http.StatusNotFound, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusOK)
}

func ListFamiliesHandler(w http.ResponseWriter, r *http.Request) {
	list, err := Store.ListFamilies()
	if err != nil {
		errorHandler(w, r, "failed to list families", http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusOK)
}

func DeleteFamilyHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	err := Store.DeleteFamily(id)
	if err != nil {
		errorHandler(w, r, "failed to delete family", http.StatusInternalServerError, err)
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
		errorHandler(w, r, "failed to read request body", http.StatusBadRequest, err)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset body for further reading

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorHandler(w, r, fmt.Sprintf("invalid JSON: %v, Body: %s", err, string(body)), http.StatusBadRequest, err)
		return
	}

	var dueDate *time.Time
	if req.DueDate != "" {
		due, err := time.Parse(time.RFC3339, req.DueDate)
		if err != nil {
			errorHandler(w, r, "invalid due_date format", http.StatusBadRequest, err)
			return
		}
		dueDate = &due
	}

	if req.FamilyID == "" || req.FamilyMember == "" {
		errorHandler(w, r, "family_id and family_member are required", http.StatusBadRequest, nil)
		return
	}

	family, err := Store.GetFamily(req.FamilyID)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("family not found: %s", req.FamilyID), http.StatusBadRequest, err)
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
		errorHandler(w, r, fmt.Sprintf("family member not found: %s", req.FamilyMember), http.StatusBadRequest, nil)
		return
	}
	if req.Recurrence.Type == "" {
		req.Recurrence.Type = "once"
	}

	// Validate recurrence pattern
	switch req.Recurrence.Type {
	case "once":
		// No additional validation needed
	case "daily":
		// No additional validation needed for daily recurrence
	case "weekly":
		if len(req.Recurrence.Days) == 0 {
			errorHandler(w, r, "weekly recurrence requires at least one day", http.StatusBadRequest, nil)
			return
		}
		for _, day := range req.Recurrence.Days {
			if !isValidWeekday(day) {
				errorHandler(w, r, "invalid weekday in recurrence pattern", http.StatusBadRequest, nil)
				return
			}
		}
	case "monthly":
		if req.Recurrence.Date < 1 || req.Recurrence.Date > 31 {
			errorHandler(w, r, "monthly recurrence requires a date between 1 and 31", http.StatusBadRequest, nil)
			return
		}
	default:
		errorHandler(w, r, "invalid recurrence type", http.StatusBadRequest, nil)
		return
	}

	if req.Recurrence.EndDate != "" {
		_, err = time.Parse(time.RFC3339, req.Recurrence.EndDate)
		if err != nil {
			errorHandler(w, r, "invalid end_date format", http.StatusBadRequest, err)
			return
		}
	}

	id := storage.GenerateReminderID(Store)
	re := reminder.NewReminderWithNullableDueDate(id, req.Title, req.Description, dueDate, req.FamilyID, req.FamilyMember, req.Recurrence)
	err = Store.CreateReminder(re)
	if err != nil {
		errorHandler(w, r, "failed to create reminder", http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(re)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusCreated)
}

func GetReminderHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	reminder, err := Store.GetReminder(id)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("reminder not found: %s", id), http.StatusNotFound, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reminder)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusOK)
}

func ListRemindersHandler(w http.ResponseWriter, r *http.Request) {
	list, err := Store.ListReminders()
	if err != nil {
		errorHandler(w, r, "failed to list reminders", http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusOK)
}

func DeleteReminderHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	err := Store.DeleteReminder(id)
	if err != nil {
		errorHandler(w, r, "failed to delete reminder", http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.UserAgent(), http.StatusNoContent)
}

func UpdateReminderHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	r, err := Store.GetReminder(id)
	if err != nil {
		errorHandler(w, req, fmt.Sprintf("reminder not found: %s", id), http.StatusNotFound, err)
		return
	}
	// Read and decode partial update
	var patch map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&patch); err != nil {
		errorHandler(w, req, "invalid JSON", http.StatusBadRequest, err)
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
				if s == "" {
					// Empty string means null due date
					r.DueDate = nil
					updated = true
				} else if t, err := time.Parse(time.RFC3339, s); err == nil {
					r.DueDate = &t
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
					errorHandler(w, req, "failed to create completion event", http.StatusInternalServerError, err)
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
	if err != nil {
		errorHandler(w, req, "failed to update reminder", http.StatusInternalServerError, err)
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
		errorHandler(w, r, "failed to read request body", http.StatusBadRequest, err)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		errorHandler(w, r, "invalid JSON", http.StatusBadRequest, err)
		return
	}
	if e.ID == "" {
		e.ID = storage.GenerateCompletionEventID(Store)
	}
	if e.ReminderID == "" || e.CompletedBy == "" {
		errorHandler(w, r, "reminder_id and completed_by are required", http.StatusBadRequest, nil)
		return
	}
	if e.CompletedAt.IsZero() {
		e.CompletedAt = time.Now()
	}
	err = Store.CreateCompletionEvent(&e)
	if err != nil {
		errorHandler(w, r, "failed to create completion event", http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}

func GetCompletionEventHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	e, err := Store.GetCompletionEvent(id)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("completion event not found: %s", id), http.StatusNotFound, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func ListCompletionEventsHandler(w http.ResponseWriter, r *http.Request) {
	reminderID := mux.Vars(r)["id"]
	if reminderID == "" {
		errorHandler(w, r, "reminder_id query param required", http.StatusBadRequest, nil)
		return
	}
	list, err := Store.ListCompletionEvents(reminderID)
	if err != nil {
		errorHandler(w, r, "failed to list completion events", http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func DeleteCompletionEventHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	err := Store.DeleteCompletionEvent(id)
	if err != nil {
		errorHandler(w, r, "failed to delete completion event", http.StatusInternalServerError, err)
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
