package storage

import (
	"errors"
	"sync"

	"reminder-app/internal/family"
	"reminder-app/internal/reminder"
)

type MemoryStorage struct {
	families                 map[string]*family.Family
	reminders                map[string]*reminder.Reminder
	completionEvents         map[string]*reminder.CompletionEvent // new
	familyIDCounter          int
	reminderIDCounter        int
	completionEventIDCounter int
	mu                       sync.Mutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		families:         make(map[string]*family.Family),
		reminders:        make(map[string]*reminder.Reminder),
		completionEvents: make(map[string]*reminder.CompletionEvent),
	}
}

// Family operations
func (m *MemoryStorage) CreateFamily(f *family.Family) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.families[f.ID] = f
	return nil
}

func (m *MemoryStorage) GetFamily(id string) (*family.Family, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, ok := m.families[id]
	if !ok {
		return nil, errors.New("family not found")
	}
	return f, nil
}

func (m *MemoryStorage) ListFamilies() ([]*family.Family, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []*family.Family
	for _, f := range m.families {
		list = append(list, f)
	}
	return list, nil
}

func (m *MemoryStorage) DeleteFamily(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.families, id)
	return nil
}

// Reminder operations
func (m *MemoryStorage) CreateReminder(r *reminder.Reminder) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reminders[r.ID] = r
	return nil
}

func (m *MemoryStorage) GetReminder(id string) (*reminder.Reminder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.reminders[id]
	if !ok {
		return nil, errors.New("reminder not found")
	}
	return r, nil
}

func (m *MemoryStorage) ListReminders() ([]*reminder.Reminder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []*reminder.Reminder
	for _, r := range m.reminders {
		list = append(list, r)
	}
	return list, nil
}

func (m *MemoryStorage) DeleteReminder(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.reminders, id)
	return nil
}

// CompletionEvent operations
func (m *MemoryStorage) CreateCompletionEvent(e *reminder.CompletionEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.completionEvents[e.ID] = e
	return nil
}

func (m *MemoryStorage) GetCompletionEvent(id string) (*reminder.CompletionEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.completionEvents[id]
	if !ok {
		return nil, errors.New("completion event not found")
	}
	return e, nil
}

func (m *MemoryStorage) ListCompletionEvents(reminderID string) ([]*reminder.CompletionEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []*reminder.CompletionEvent
	for _, e := range m.completionEvents {
		if e.ReminderID == reminderID {
			list = append(list, e)
		}
	}
	return list, nil
}

func (m *MemoryStorage) DeleteCompletionEvent(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.completionEvents, id)
	return nil
}
func (fs *MemoryStorage) GetCompletionEventIDCounter() int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.completionEventIDCounter
}

func (fs *MemoryStorage) GetFamilyIDCounter() int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.familyIDCounter
}

func (fs *MemoryStorage) GetReminderIDCounter() int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.reminderIDCounter
}

// Counter setter methods (useful for restoring state or testing)
func (m *MemoryStorage) SetFamilyIDCounter(counter int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.familyIDCounter = counter
	return nil
}

func (m *MemoryStorage) SetReminderIDCounter(counter int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reminderIDCounter = counter
	return nil
}

func (m *MemoryStorage) SetCompletionEventIDCounter(counter int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.completionEventIDCounter = counter
	return nil
}
