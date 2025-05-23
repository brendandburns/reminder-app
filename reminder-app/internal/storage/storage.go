package storage

import (
	"fmt"
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

	// ID counter operations
	GetFamilyIDCounter() int
	SetFamilyIDCounter(counter int) error
	GetReminderIDCounter() int
	SetReminderIDCounter(counter int) error
	GetCompletionEventIDCounter() int
	SetCompletionEventIDCounter(counter int) error
}

func GenerateFamilyID(s Storage) string {
	// Generate a new family ID
	counter := s.GetFamilyIDCounter()
	counter++
	s.SetFamilyIDCounter(counter)
	return fmt.Sprintf("fam%d", counter)
}

func GenerateReminderID(s Storage) string {
	// Generate a new reminder ID
	counter := s.GetReminderIDCounter()
	counter++
	s.SetReminderIDCounter(counter)
	return fmt.Sprintf("rem%d", counter)
}

func GenerateCompletionEventID(s Storage) string {
	// Generate a new completion event ID
	counter := s.GetCompletionEventIDCounter()
	counter++
	s.SetCompletionEventIDCounter(counter)
	return fmt.Sprintf("cev%d", counter)
}
