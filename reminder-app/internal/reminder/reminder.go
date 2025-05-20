package reminder

import "time"

type Reminder struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	DueDate      time.Time `json:"due_date"`
	Completed    bool      `json:"completed"`
	FamilyID     string    `json:"family_id"`
	FamilyMember string    `json:"family_member"`
}

func NewReminder(id, title, description string, dueDate time.Time, familyID, familyMember string) *Reminder {
	return &Reminder{
		ID:           id,
		Title:        title,
		Description:  description,
		DueDate:      dueDate,
		Completed:    false,
		FamilyID:     familyID,
		FamilyMember: familyMember,
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
