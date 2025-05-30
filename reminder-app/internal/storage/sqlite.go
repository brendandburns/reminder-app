package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"reminder-app/internal/family"
	"reminder-app/internal/reminder"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
	mu sync.Mutex
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	s := &SQLiteStorage{db: db}

	// Create tables if they don't exist
	if err := s.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return s, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// createTables creates the necessary tables
func (s *SQLiteStorage) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS families (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			members TEXT NOT NULL -- JSON array of members
		)`,
		`CREATE TABLE IF NOT EXISTS reminders (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			due_date TEXT, -- ISO 8601 format, nullable
			recurrence_type TEXT NOT NULL,
			recurrence_days TEXT, -- JSON array for weekly days
			recurrence_date INTEGER, -- Day of month for monthly
			recurrence_end_date TEXT, -- ISO 8601 format
			completed BOOLEAN NOT NULL DEFAULT 0,
			completed_at TEXT, -- ISO 8601 format
			family_id TEXT NOT NULL,
			family_member TEXT NOT NULL,
			FOREIGN KEY (family_id) REFERENCES families(id)
		)`,
		`CREATE TABLE IF NOT EXISTS completion_events (
			id TEXT PRIMARY KEY,
			reminder_id TEXT NOT NULL,
			completed_at TEXT NOT NULL, -- ISO 8601 format
			completed_by TEXT NOT NULL,
			FOREIGN KEY (reminder_id) REFERENCES reminders(id)
		)`,
		`CREATE TABLE IF NOT EXISTS counters (
			name TEXT PRIMARY KEY,
			value INTEGER NOT NULL DEFAULT 0
		)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %q: %w", query, err)
		}
	}

	// Initialize counters if they don't exist
	counterNames := []string{"family_id", "reminder_id", "completion_event_id"}
	for _, name := range counterNames {
		_, err := s.db.Exec("INSERT OR IGNORE INTO counters (name, value) VALUES (?, 0)", name)
		if err != nil {
			return fmt.Errorf("failed to initialize counter %s: %w", name, err)
		}
	}

	return nil
}

// Family operations
func (s *SQLiteStorage) CreateFamily(f *family.Family) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	membersJSON, err := json.Marshal(f.Members)
	if err != nil {
		return fmt.Errorf("failed to marshal family members: %w", err)
	}

	_, err = s.db.Exec("INSERT INTO families (id, name, members) VALUES (?, ?, ?)",
		f.ID, f.Name, string(membersJSON))
	if err != nil {
		return fmt.Errorf("failed to create family: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) GetFamily(id string) (*family.Family, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var f family.Family
	var membersJSON string

	err := s.db.QueryRow("SELECT id, name, members FROM families WHERE id = ?", id).
		Scan(&f.ID, &f.Name, &membersJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("family not found")
		}
		return nil, fmt.Errorf("failed to get family: %w", err)
	}

	if err := json.Unmarshal([]byte(membersJSON), &f.Members); err != nil {
		return nil, fmt.Errorf("failed to unmarshal family members: %w", err)
	}

	return &f, nil
}

func (s *SQLiteStorage) ListFamilies() ([]*family.Family, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query("SELECT id, name, members FROM families")
	if err != nil {
		return nil, fmt.Errorf("failed to list families: %w", err)
	}
	defer rows.Close()

	var families []*family.Family
	for rows.Next() {
		var f family.Family
		var membersJSON string

		if err := rows.Scan(&f.ID, &f.Name, &membersJSON); err != nil {
			return nil, fmt.Errorf("failed to scan family: %w", err)
		}

		if err := json.Unmarshal([]byte(membersJSON), &f.Members); err != nil {
			return nil, fmt.Errorf("failed to unmarshal family members: %w", err)
		}

		families = append(families, &f)
	}

	return families, nil
}

func (s *SQLiteStorage) DeleteFamily(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM families WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete family: %w", err)
	}

	return nil
}

// Reminder operations
func (s *SQLiteStorage) CreateReminder(r *reminder.Reminder) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	recurrenceDaysJSON, err := json.Marshal(r.Recurrence.Days)
	if err != nil {
		return fmt.Errorf("failed to marshal recurrence days: %w", err)
	}

	var completedAtStr *string
	if r.CompletedAt != nil {
		str := r.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		completedAtStr = &str
	}

	var dueDateStr *string
	if r.DueDate != nil {
		str := r.DueDate.Format("2006-01-02T15:04:05Z07:00")
		dueDateStr = &str
	}

	// Handle empty end date by setting it to a very far future date
	endDate := r.Recurrence.EndDate
	if endDate == "" {
		// Set to year 2099 for "no end date" recurring reminders
		endDate = "2099-12-31T23:59:59Z"
	}

	_, err = s.db.Exec(`INSERT OR REPLACE INTO reminders 
		(id, title, description, due_date, recurrence_type, recurrence_days, 
		recurrence_date, recurrence_end_date, completed, completed_at, family_id, family_member) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.Title, r.Description, dueDateStr,
		r.Recurrence.Type, string(recurrenceDaysJSON), r.Recurrence.Date,
		endDate, r.Completed, completedAtStr, r.FamilyID, r.FamilyMember)
	if err != nil {
		return fmt.Errorf("failed to create/update reminder: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) GetReminder(id string) (*reminder.Reminder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var r reminder.Reminder
	var dueDateStr *string
	var recurrenceDaysJSON string
	var completedAtStr *string

	err := s.db.QueryRow(`SELECT id, title, description, due_date, recurrence_type, 
		recurrence_days, recurrence_date, recurrence_end_date, completed, completed_at, 
		family_id, family_member FROM reminders WHERE id = ?`, id).
		Scan(&r.ID, &r.Title, &r.Description, &dueDateStr, &r.Recurrence.Type,
			&recurrenceDaysJSON, &r.Recurrence.Date, &r.Recurrence.EndDate,
			&r.Completed, &completedAtStr, &r.FamilyID, &r.FamilyMember)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("reminder not found")
		}
		return nil, fmt.Errorf("failed to get reminder: %w", err)
	}

	// Parse due date if not null
	if dueDateStr != nil {
		dueDate, err := parseTimeString(*dueDateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse due date: %w", err)
		}
		r.DueDate = &dueDate
	}

	// Parse completed at
	if completedAtStr != nil {
		completedAt, err := parseTimeString(*completedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse completed at: %w", err)
		}
		r.CompletedAt = &completedAt
	}

	// Convert far future end date back to empty string for API consistency
	if r.Recurrence.EndDate == "2099-12-31T23:59:59Z" {
		r.Recurrence.EndDate = ""
	}

	// Parse recurrence days
	if err := json.Unmarshal([]byte(recurrenceDaysJSON), &r.Recurrence.Days); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recurrence days: %w", err)
	}

	return &r, nil
}

func (s *SQLiteStorage) ListReminders() ([]*reminder.Reminder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`SELECT id, title, description, due_date, recurrence_type, 
		recurrence_days, recurrence_date, recurrence_end_date, completed, completed_at, 
		family_id, family_member FROM reminders`)
	if err != nil {
		return nil, fmt.Errorf("failed to list reminders: %w", err)
	}
	defer rows.Close()

	var reminders []*reminder.Reminder
	for rows.Next() {
		var r reminder.Reminder
		var dueDateStr *string
		var recurrenceDaysJSON string
		var completedAtStr *string

		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &dueDateStr, &r.Recurrence.Type,
			&recurrenceDaysJSON, &r.Recurrence.Date, &r.Recurrence.EndDate,
			&r.Completed, &completedAtStr, &r.FamilyID, &r.FamilyMember); err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}

		// Parse due date if not null
		if dueDateStr != nil {
			dueDate, err := parseTimeString(*dueDateStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse due date: %w", err)
			}
			r.DueDate = &dueDate
		}

		// Parse completed at
		if completedAtStr != nil {
			completedAt, err := parseTimeString(*completedAtStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse completed at: %w", err)
			}
			r.CompletedAt = &completedAt
		}

		// Convert far future end date back to empty string for API consistency
		if r.Recurrence.EndDate == "2099-12-31T23:59:59Z" {
			r.Recurrence.EndDate = ""
		}

		// Parse recurrence days
		if err := json.Unmarshal([]byte(recurrenceDaysJSON), &r.Recurrence.Days); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recurrence days: %w", err)
		}

		reminders = append(reminders, &r)
	}

	return reminders, nil
}

func (s *SQLiteStorage) DeleteReminder(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM reminders WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}

	return nil
}

// CompletionEvent operations
func (s *SQLiteStorage) CreateCompletionEvent(e *reminder.CompletionEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("INSERT OR REPLACE INTO completion_events (id, reminder_id, completed_at, completed_by) VALUES (?, ?, ?, ?)",
		e.ID, e.ReminderID, e.CompletedAt.Format("2006-01-02T15:04:05Z07:00"), e.CompletedBy)
	if err != nil {
		return fmt.Errorf("failed to create/update completion event: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) GetCompletionEvent(id string) (*reminder.CompletionEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var e reminder.CompletionEvent
	var completedAtStr string

	err := s.db.QueryRow("SELECT id, reminder_id, completed_at, completed_by FROM completion_events WHERE id = ?", id).
		Scan(&e.ID, &e.ReminderID, &completedAtStr, &e.CompletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("completion event not found")
		}
		return nil, fmt.Errorf("failed to get completion event: %w", err)
	}

	// Parse completed at
	if e.CompletedAt, err = parseTimeString(completedAtStr); err != nil {
		return nil, fmt.Errorf("failed to parse completed at: %w", err)
	}

	return &e, nil
}

func (s *SQLiteStorage) ListCompletionEvents(reminderID string) ([]*reminder.CompletionEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query("SELECT id, reminder_id, completed_at, completed_by FROM completion_events WHERE reminder_id = ?", reminderID)
	if err != nil {
		return nil, fmt.Errorf("failed to list completion events: %w", err)
	}
	defer rows.Close()

	var events []*reminder.CompletionEvent
	for rows.Next() {
		var e reminder.CompletionEvent
		var completedAtStr string

		if err := rows.Scan(&e.ID, &e.ReminderID, &completedAtStr, &e.CompletedBy); err != nil {
			return nil, fmt.Errorf("failed to scan completion event: %w", err)
		}

		// Parse completed at
		if e.CompletedAt, err = parseTimeString(completedAtStr); err != nil {
			return nil, fmt.Errorf("failed to parse completed at: %w", err)
		}

		events = append(events, &e)
	}

	return events, nil
}

func (s *SQLiteStorage) DeleteCompletionEvent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM completion_events WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete completion event: %w", err)
	}

	return nil
}

// ID counter operations
func (s *SQLiteStorage) GetFamilyIDCounter() int {
	return s.getCounter("family_id")
}

func (s *SQLiteStorage) SetFamilyIDCounter(counter int) error {
	return s.setCounter("family_id", counter)
}

func (s *SQLiteStorage) GetReminderIDCounter() int {
	return s.getCounter("reminder_id")
}

func (s *SQLiteStorage) SetReminderIDCounter(counter int) error {
	return s.setCounter("reminder_id", counter)
}

func (s *SQLiteStorage) GetCompletionEventIDCounter() int {
	return s.getCounter("completion_event_id")
}

func (s *SQLiteStorage) SetCompletionEventIDCounter(counter int) error {
	return s.setCounter("completion_event_id", counter)
}

// Helper methods
func (s *SQLiteStorage) getCounter(name string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	var value int
	err := s.db.QueryRow("SELECT value FROM counters WHERE name = ?", name).Scan(&value)
	if err != nil {
		return 0
	}
	return value
}

func (s *SQLiteStorage) setCounter(name string, value int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("UPDATE counters SET value = ? WHERE name = ?", value, name)
	return err
}

// parseTimeString parses a time string in ISO 8601 format
func parseTimeString(timeStr string) (time.Time, error) {
	// Try multiple time formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time string: %s", timeStr)
}
