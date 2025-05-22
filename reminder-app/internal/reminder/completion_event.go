package reminder

import "time"

type CompletionEvent struct {
	ID          string    `json:"id"`
	ReminderID  string    `json:"reminder_id"`
	CompletedAt time.Time `json:"completed_at"`
	CompletedBy string    `json:"completed_by"`
}
