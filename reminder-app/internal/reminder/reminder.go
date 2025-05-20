package reminder

import "time"

type Reminder struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Completed   bool      `json:"completed"`
}

func NewReminder(id, title, description string, dueDate time.Time) *Reminder {
	return &Reminder{
		ID:          id,
		Title:       title,
		Description: description,
		DueDate:     dueDate,
		Completed:   false,
	}
}

func (r *Reminder) Update(title, description string, dueDate time.Time) {
	r.Title = title
	r.Description = description
	r.DueDate = dueDate
}

func (r *Reminder) MarkCompleted() {
	r.Completed = true
}

func (r *Reminder) Delete() {
	// Logic to delete the reminder
}