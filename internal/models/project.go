package models

type Project struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type Client struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TaskAssignment struct {
	ID       int  `json:"id"`
	Billable bool `json:"billable"`
	IsActive bool `json:"is_active"`
	Task     Task `json:"task"`
}

type ProjectAssignment struct {
	ID              int              `json:"id"`
	IsActive        bool             `json:"is_active"`
	Project         Project          `json:"project"`
	Client          Client           `json:"client"`
	TaskAssignments []TaskAssignment `json:"task_assignments"`
}

type ProjectAssignmentsResponse struct {
	ProjectAssignments []ProjectAssignment `json:"project_assignments"`
	PerPage            int                 `json:"per_page"`
	TotalPages         int                 `json:"total_pages"`
	TotalEntries       int                 `json:"total_entries"`
	Page               int                 `json:"page"`
}
