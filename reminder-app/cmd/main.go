package main

import (
	"flag"
	"log"
	"mime"
	"net/http"
	"path/filepath"

	"reminder-app/internal/handlers"
	"reminder-app/internal/storage"

	"github.com/gorilla/mux"
)

func main() {
	staticDir := flag.String("static", "./static", "directory to serve static files from")
	tlsCert := flag.String("tls-cert", "", "path to TLS certificate file (optional)")
	tlsKey := flag.String("tls-key", "", "path to TLS key file (optional)")
	flag.Parse()

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
	r.HandleFunc("/reminders/{id}/completion-events", handlers.ListCompletionEventsHandler).Methods("GET")
	r.HandleFunc("/completion-events/{id}", handlers.GetCompletionEventHandler).Methods("GET")
	r.HandleFunc("/completion-events/{id}", handlers.DeleteCompletionEventHandler).Methods("DELETE")

	// Static file server for frontend at "/"
	staticFs := http.FileServer(http.Dir(*staticDir))
	r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		ext := filepath.Ext(path)
		if ext != "" {
			if ctype := mime.TypeByExtension(ext); ctype != "" {
				w.Header().Set("Content-Type", ctype)
			}
		}
		staticFs.ServeHTTP(w, req)
	}))

	if *tlsCert != "" && *tlsKey != "" {
		addr := ":443"
		log.Println("Starting reminder app with HTTPS on", addr, "serving static files from", *staticDir)
		if err := http.ListenAndServeTLS(addr, *tlsCert, *tlsKey, r); err != nil {
			log.Fatalf("Could not start HTTPS server: %s\n", err)
		}
	} else {
		addr := ":8080"
		log.Println("Starting reminder app with HTTP on", addr, "serving static files from", *staticDir)
		if err := http.ListenAndServe(addr, r); err != nil {
			log.Fatalf("Could not start HTTP server: %s\n", err)
		}
	}
}
