package main

import (
	"log"
	"net/http"

	"reminder-app/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {
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

	log.Println("Starting reminder app on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
