package main

import (
	"log"
	"net/http"
	"sync"

	"reminder-app/internal/handlers"
	"reminder-app/internal/storage"

	"github.com/gorilla/mux"
)

var (
	mu                sync.Mutex
	familyIDCounter   int
	reminderIDCounter int
)

func main() {
	// Choose storage implementation here
	// For memory-based:
	// handlers.Store = storage.NewMemoryStorage()
	// For file-based:
	// handlers.Store = storage.NewFileStorage("families.json", "reminders.json")

	//handlers.Store = storage.NewMemoryStorage() // Default to memory, change to file as needed
	handlers.Store = storage.NewFileStorage("families.json", "reminders.json", "completion_events.json")

	r := mux.NewRouter()

	// Family routes
	r.HandleFunc("/families", handlers.CreateFamilyHandler).Methods("POST")
	r.HandleFunc("/families", handlers.ListFamiliesHandler).Methods("GET")
	r.HandleFunc("/families/{id}", handlers.GetFamilyHandler).Methods("GET")
	r.HandleFunc("/families/{id}", handlers.DeleteFamilyHandler).Methods("DELETE")

	// Reminder routes
	r.HandleFunc("/reminders", handlers.CreateReminderHandler).Methods("POST")
	r.HandleFunc("/reminders", handlers.ListRemindersHandler).Methods("GET")
	r.HandleFunc("/reminders/{id}", handlers.GetReminderHandler).Methods("GET")
	r.HandleFunc("/reminders/{id}", handlers.DeleteReminderHandler).Methods("DELETE")
	r.HandleFunc("/reminders/{id}", handlers.UpdateReminderHandler).Methods("PATCH")

	// CompletionEvent routes
	r.HandleFunc("/completion-events", handlers.CreateCompletionEventHandler).Methods("POST")
	r.HandleFunc("/completion-events", handlers.ListCompletionEventsHandler).Methods("GET")
	r.HandleFunc("/completion-events/{id}", handlers.GetCompletionEventHandler).Methods("GET")
	r.HandleFunc("/completion-events/{id}", handlers.DeleteCompletionEventHandler).Methods("DELETE")

	log.Println("Starting reminder app on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
