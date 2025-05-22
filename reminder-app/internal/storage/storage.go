package storage

import (
	"reminder-app/internal/family"
	"reminder-app/internal/reminder"
)

// Storage defines the interface for data persistence
// for families and reminders.
type Storage interface {
	// Family operations
	CreateFamily(f *family.Family) error
	GetFamily(id string) (*family.Family, error)
	ListFamilies() ([]*family.Family, error)
	DeleteFamily(id string) error

	// Reminder operations
	CreateReminder(r *reminder.Reminder) error
	GetReminder(id string) (*reminder.Reminder, error)
	ListReminders() ([]*reminder.Reminder, error)
	DeleteReminder(id string) error

	// CompletionEvent operations
	CreateCompletionEvent(e *reminder.CompletionEvent) error
	GetCompletionEvent(id string) (*reminder.CompletionEvent, error)
	ListCompletionEvents(reminderID string) ([]*reminder.CompletionEvent, error)
	DeleteCompletionEvent(id string) error
}
