package prompt

import (
	"fmt"
	"time"

	"harvest/internal/models"
)

func SelectProject(projects []models.ProjectAssignment) (*models.ProjectAssignment, error) {
	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects available")
	}

	items := make([]selectItem, len(projects))
	for i, p := range projects {
		items[i] = selectItem{
			title: p.Project.Name,
			desc:  p.Client.Name,
			index: i,
		}
	}

	idx, err := runSelect("Select project", items)
	if err != nil {
		return nil, err
	}

	return &projects[idx], nil
}

func SelectTask(tasks []models.TaskAssignment) (*models.TaskAssignment, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no tasks available for this project")
	}

	// Filter to only active tasks
	var activeTasks []models.TaskAssignment
	for _, t := range tasks {
		if t.IsActive {
			activeTasks = append(activeTasks, t)
		}
	}

	if len(activeTasks) == 0 {
		return nil, fmt.Errorf("no active tasks available for this project")
	}

	items := make([]selectItem, len(activeTasks))
	for i, t := range activeTasks {
		items[i] = selectItem{
			title: t.Task.Name,
			index: i,
		}
	}

	idx, err := runSelect("Select task", items)
	if err != nil {
		return nil, err
	}

	return &activeTasks[idx], nil
}

func InputHours() (float64, error) {
	validate := func(input string) error {
		if input == "" {
			return fmt.Errorf("hours is required")
		}

		hours, err := ParseHours(input)
		if err != nil {
			return err
		}

		if hours <= 0 {
			return fmt.Errorf("hours must be greater than 0")
		}
		if hours > 24 {
			return fmt.Errorf("hours cannot exceed 24")
		}

		return nil
	}

	result, err := runInput("Hours (e.g., 1.5, 1:30, 90m)", "", "", validate)
	if err != nil {
		return 0, err
	}

	return ParseHours(result)
}

func InputNotes() (string, error) {
	return runInput("Notes (optional)", "", "", nil)
}

func InputDate() (string, error) {
	today := time.Now().Format("2006-01-02")

	validate := func(input string) error {
		if input == "" {
			return nil
		}
		_, err := time.Parse("2006-01-02", input)
		if err != nil {
			return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		return nil
	}

	result, err := runInput("Date (YYYY-MM-DD)", "", today, validate)
	if err != nil {
		return "", err
	}

	if result == "" {
		return today, nil
	}
	return result, nil
}

func Confirm(message string) (bool, error) {
	return runConfirm(message)
}

func SelectTimeEntry(entries []models.TimeEntry) (*models.TimeEntry, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("no time entries to select from")
	}

	items := make([]selectItem, len(entries))
	for i, e := range entries {
		hoursStr := fmt.Sprintf("%.2f hrs", e.Hours)
		if e.IsRunning {
			hoursStr += " (running)"
		}
		title := fmt.Sprintf("%s  %s / %s", hoursStr, e.Project.Name, e.Task.Name)
		desc := ""
		if e.Notes != "" {
			desc = e.Notes
		}
		if e.IsLocked {
			desc = "[locked] " + desc
		}
		items[i] = selectItem{
			title: title,
			desc:  desc,
			index: i,
		}
	}

	idx, err := runSelect("Select time entry", items)
	if err != nil {
		return nil, err
	}

	return &entries[idx], nil
}

func SelectEditField() (string, error) {
	fields := []selectItem{
		{title: "Hours", index: 0},
		{title: "Notes", index: 1},
		{title: "Project & Task", index: 2},
		{title: "Date", index: 3},
		{title: "Done (save & exit)", index: 4},
	}

	idx, err := runSelect("What to edit?", fields)
	if err != nil {
		return "", err
	}

	switch idx {
	case 0:
		return "hours", nil
	case 1:
		return "notes", nil
	case 2:
		return "project", nil
	case 3:
		return "date", nil
	default:
		return "done", nil
	}
}
