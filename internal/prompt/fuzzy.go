package prompt

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Adam-D-Lewis/harvest-go-cli/internal/models"
)

// ParseHours parses various hour input formats into a float64.
// Supported formats: decimal (1.5), time (1:30), duration (90m, 1h30m).
func ParseHours(input string) (float64, error) {
	input = strings.TrimSpace(input)

	// Try decimal format first (1.5)
	if hours, err := strconv.ParseFloat(input, 64); err == nil {
		return hours, nil
	}

	// Try time format (1:30)
	if strings.Contains(input, ":") {
		parts := strings.Split(input, ":")
		if len(parts) == 2 {
			hours, err1 := strconv.ParseFloat(parts[0], 64)
			minutes, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return hours + (minutes / 60), nil
			}
		}
	}

	// Try duration format (90m, 1h30m, 1h)
	input = strings.ToLower(input)
	if strings.Contains(input, "h") || strings.Contains(input, "m") {
		duration, err := time.ParseDuration(input)
		if err == nil {
			return duration.Hours(), nil
		}
	}

	return 0, fmt.Errorf("invalid format. Use decimal (1.5), time (1:30), or duration (90m)")
}

// FuzzyMatchProject finds projects matching the query string.
// Returns matched projects. If exactly one match, returns it directly.
// If multiple matches, caller should use SelectProject with filtered list.
func FuzzyMatchProject(projects []models.ProjectAssignment, query string) []models.ProjectAssignment {
	if query == "" {
		return projects
	}

	query = strings.ToLower(query)

	// First, check for exact match
	for _, p := range projects {
		if strings.ToLower(p.Project.Name) == query {
			return []models.ProjectAssignment{p}
		}
	}

	// Then, check for substring matches
	var matches []models.ProjectAssignment
	for _, p := range projects {
		name := strings.ToLower(p.Project.Name)
		client := strings.ToLower(p.Client.Name)
		if strings.Contains(name, query) || strings.Contains(client, query) {
			matches = append(matches, p)
		}
	}

	return matches
}

// FuzzyMatchTask finds tasks matching the query string.
// Returns matched tasks. Only considers active tasks.
func FuzzyMatchTask(tasks []models.TaskAssignment, query string) []models.TaskAssignment {
	// Filter to active tasks first
	var activeTasks []models.TaskAssignment
	for _, t := range tasks {
		if t.IsActive {
			activeTasks = append(activeTasks, t)
		}
	}

	if query == "" {
		return activeTasks
	}

	query = strings.ToLower(query)

	// First, check for exact match
	for _, t := range activeTasks {
		if strings.ToLower(t.Task.Name) == query {
			return []models.TaskAssignment{t}
		}
	}

	// Then, check for substring matches
	var matches []models.TaskAssignment
	for _, t := range activeTasks {
		name := strings.ToLower(t.Task.Name)
		if strings.Contains(name, query) {
			matches = append(matches, t)
		}
	}

	return matches
}
