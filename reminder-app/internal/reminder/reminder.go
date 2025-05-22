package reminder

import (
	"strings"
	"time"
)

type RecurrencePattern struct {
	Type    string   `json:"type"`     // "once", "weekly", "monthly"
	Days    []string `json:"days"`     // ["monday", "wednesday", etc] for weekly
	Date    int      `json:"date"`     // 1-31 for monthly
	EndDate string   `json:"end_date"` // Optional end date for recurrence
}

type Reminder struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	DueDate      time.Time         `json:"due_date"`
	Recurrence   RecurrencePattern `json:"recurrence"`
	Completed    bool              `json:"completed"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
	FamilyID     string            `json:"family_id"`
	FamilyMember string            `json:"family_member"`
}

func NewReminder(id, title, description string, dueDate time.Time, familyID, familyMember string, recurrence RecurrencePattern) *Reminder {
	return &Reminder{
		ID:           id,
		Title:        title,
		Description:  description,
		DueDate:      dueDate,
		Recurrence:   recurrence,
		Completed:    false,
		CompletedAt:  nil,
		FamilyID:     familyID,
		FamilyMember: familyMember,
	}
}

// IsRecurring returns true if the reminder is a recurring reminder
func (r *Reminder) IsRecurring() bool {
	return r.Recurrence.Type != "once"
}

// NextOccurrence returns the next occurrence of the reminder after the given time
func (r *Reminder) NextOccurrence(after time.Time) *time.Time {
	if r.Recurrence.Type == "once" {
		if r.DueDate.After(after) {
			return &r.DueDate
		}
		return nil
	}

	if r.Recurrence.EndDate != "" {
		endDate, err := time.Parse(time.RFC3339, r.Recurrence.EndDate)
		if err == nil && after.After(endDate) {
			return nil
		}
	}

	next := after
	switch r.Recurrence.Type {
	case "weekly":
		// Find the next matching day of the week
		for i := 0; i < 7; i++ {
			next = next.AddDate(0, 0, 1)
			weekday := strings.ToLower(next.Weekday().String())
			for _, day := range r.Recurrence.Days {
				if day == weekday {
					result := time.Date(
						next.Year(), next.Month(), next.Day(),
						r.DueDate.Hour(), r.DueDate.Minute(), r.DueDate.Second(),
						0, next.Location(),
					)
					return &result
				}
			}
		}
	case "monthly":
		// Find the next matching date of the month
		next = time.Date(
			after.Year(), after.Month(), r.Recurrence.Date,
			r.DueDate.Hour(), r.DueDate.Minute(), r.DueDate.Second(),
			0, after.Location(),
		)
		if next.Before(after) {
			next = next.AddDate(0, 1, 0)
		}
		return &next
	}
	return nil
}

func (r *Reminder) Update(title, description string, dueDate time.Time) {
	r.Title = title
	r.Description = description
	r.DueDate = dueDate
}

func (r *Reminder) MarkCompleted() {
	now := time.Now()
	r.Completed = true
	r.CompletedAt = &now
}

func (r *Reminder) Delete() {
	// Logic to delete the reminder
}
