package models

import "time"

type TimeEntry struct {
	ID        int       `json:"id"`
	SpentDate string    `json:"spent_date"`
	Hours     float64   `json:"hours"`
	Notes     string    `json:"notes"`
	Billable     bool   `json:"billable"`
	IsRunning    bool   `json:"is_running"`
	IsLocked     bool   `json:"is_locked"`
	LockedReason string `json:"locked_reason"`
	Project   Project   `json:"project"`
	Task      Task      `json:"task"`
	Client    Client    `json:"client"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TimeEntriesResponse struct {
	TimeEntries  []TimeEntry `json:"time_entries"`
	PerPage      int         `json:"per_page"`
	TotalPages   int         `json:"total_pages"`
	TotalEntries int         `json:"total_entries"`
	Page         int         `json:"page"`
}

type CreateTimeEntryRequest struct {
	ProjectID int     `json:"project_id"`
	TaskID    int     `json:"task_id"`
	SpentDate string  `json:"spent_date"`
	Hours     float64 `json:"hours,omitempty"`
	Notes     string  `json:"notes,omitempty"`
}

type UpdateTimeEntryRequest struct {
	ProjectID *int     `json:"project_id,omitempty"`
	TaskID    *int     `json:"task_id,omitempty"`
	SpentDate *string  `json:"spent_date,omitempty"`
	Hours     *float64 `json:"hours,omitempty"`
	Notes     *string  `json:"notes,omitempty"`
}
