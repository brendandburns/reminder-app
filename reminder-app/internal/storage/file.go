package storage

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"reminder-app/internal/family"
	"reminder-app/internal/reminder"
)

type FileStorage struct {
	familyFile          string
	reminderFile        string
	completionEventFile string
	mu                  sync.Mutex
}

func NewFileStorage(familyFile, reminderFile, completionFile string) *FileStorage {
	return &FileStorage{
		familyFile:          familyFile,
		reminderFile:        reminderFile,
		completionEventFile: completionFile,
	}
}

// Helper functions for file IO
func (fs *FileStorage) loadFamilies() (map[string]*family.Family, error) {
	families := make(map[string]*family.Family)
	if _, err := os.Stat(fs.familyFile); os.IsNotExist(err) {
		return families, nil
	}
	data, err := ioutil.ReadFile(fs.familyFile)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return families, nil
	}
	if err := json.Unmarshal(data, &families); err != nil {
		return nil, err
	}
	return families, nil
}

func (fs *FileStorage) saveFamilies(families map[string]*family.Family) error {
	data, err := json.MarshalIndent(families, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fs.familyFile, data, 0644)
}

func (fs *FileStorage) loadReminders() (map[string]*reminder.Reminder, error) {
	reminders := make(map[string]*reminder.Reminder)
	if _, err := os.Stat(fs.reminderFile); os.IsNotExist(err) {
		return reminders, nil
	}
	data, err := ioutil.ReadFile(fs.reminderFile)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return reminders, nil
	}
	if err := json.Unmarshal(data, &reminders); err != nil {
		return nil, err
	}
	return reminders, nil
}

func (fs *FileStorage) saveReminders(reminders map[string]*reminder.Reminder) error {
	data, err := json.MarshalIndent(reminders, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fs.reminderFile, data, 0644)
}

// Helper functions for CompletionEvent file IO
func (fs *FileStorage) loadCompletionEvents() (map[string]*reminder.CompletionEvent, error) {
	events := make(map[string]*reminder.CompletionEvent)
	if _, err := os.Stat(fs.completionEventFile); os.IsNotExist(err) {
		return events, nil
	}
	data, err := ioutil.ReadFile(fs.completionEventFile)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return events, nil
	}
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (fs *FileStorage) saveCompletionEvents(events map[string]*reminder.CompletionEvent) error {
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fs.completionEventFile, data, 0644)
}

// Family operations
func (fs *FileStorage) CreateFamily(f *family.Family) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	families, err := fs.loadFamilies()
	if err != nil {
		return err
	}
	families[f.ID] = f
	return fs.saveFamilies(families)
}

func (fs *FileStorage) GetFamily(id string) (*family.Family, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	families, err := fs.loadFamilies()
	if err != nil {
		return nil, err
	}
	f, ok := families[id]
	if !ok {
		return nil, errors.New("family not found")
	}
	return f, nil
}

func (fs *FileStorage) ListFamilies() ([]*family.Family, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	families, err := fs.loadFamilies()
	if err != nil {
		return nil, err
	}
	var list []*family.Family
	for _, f := range families {
		list = append(list, f)
	}
	return list, nil
}

func (fs *FileStorage) DeleteFamily(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	families, err := fs.loadFamilies()
	if err != nil {
		return err
	}
	delete(families, id)
	return fs.saveFamilies(families)
}

// Reminder operations
func (fs *FileStorage) CreateReminder(r *reminder.Reminder) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	reminders, err := fs.loadReminders()
	if err != nil {
		return err
	}
	reminders[r.ID] = r
	return fs.saveReminders(reminders)
}

func (fs *FileStorage) GetReminder(id string) (*reminder.Reminder, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	reminders, err := fs.loadReminders()
	if err != nil {
		return nil, err
	}
	r, ok := reminders[id]
	if !ok {
		return nil, errors.New("reminder not found")
	}
	return r, nil
}

func (fs *FileStorage) ListReminders() ([]*reminder.Reminder, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	reminders, err := fs.loadReminders()
	if err != nil {
		return nil, err
	}
	var list []*reminder.Reminder
	for _, r := range reminders {
		list = append(list, r)
	}
	return list, nil
}

func (fs *FileStorage) DeleteReminder(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	reminders, err := fs.loadReminders()
	if err != nil {
		return err
	}
	delete(reminders, id)
	return fs.saveReminders(reminders)
}

// CompletionEvent operations
func (fs *FileStorage) CreateCompletionEvent(e *reminder.CompletionEvent) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	events, err := fs.loadCompletionEvents()
	if err != nil {
		return err
	}
	events[e.ID] = e
	return fs.saveCompletionEvents(events)
}

func (fs *FileStorage) GetCompletionEvent(id string) (*reminder.CompletionEvent, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	events, err := fs.loadCompletionEvents()
	if err != nil {
		return nil, err
	}
	e, ok := events[id]
	if !ok {
		return nil, errors.New("completion event not found")
	}
	return e, nil
}

func (fs *FileStorage) ListCompletionEvents(reminderID string) ([]*reminder.CompletionEvent, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	events, err := fs.loadCompletionEvents()
	if err != nil {
		return nil, err
	}
	var list []*reminder.CompletionEvent
	for _, e := range events {
		if e.ReminderID == reminderID {
			list = append(list, e)
		}
	}
	return list, nil
}

func (fs *FileStorage) DeleteCompletionEvent(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	events, err := fs.loadCompletionEvents()
	if err != nil {
		return err
	}
	delete(events, id)
	return fs.saveCompletionEvents(events)
}
