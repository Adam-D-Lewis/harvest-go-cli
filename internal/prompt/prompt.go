package prompt

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"

	"harvest-cli/internal/models"
)

func SelectProject(projects []models.ProjectAssignment) (*models.ProjectAssignment, error) {
	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects available")
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ .Project.Name | cyan }} ({{ .Client.Name | faint }})",
		Inactive: "  {{ .Project.Name }} ({{ .Client.Name | faint }})",
		Selected: "* {{ .Project.Name | green }}",
	}

	searcher := func(input string, index int) bool {
		project := projects[index]
		name := strings.ToLower(project.Project.Name + " " + project.Client.Name)
		input = strings.ToLower(input)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Select project",
		Items:     projects,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	idx, _, err := prompt.Run()
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

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ .Task.Name | cyan }}",
		Inactive: "  {{ .Task.Name }}",
		Selected: "* {{ .Task.Name | green }}",
	}

	searcher := func(input string, index int) bool {
		task := activeTasks[index]
		name := strings.ToLower(task.Task.Name)
		input = strings.ToLower(input)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Select task",
		Items:     activeTasks,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	idx, _, err := prompt.Run()
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

		// Support formats: "1.5", "1:30", "90m", "1h30m"
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

	prompt := promptui.Prompt{
		Label:    "Hours (e.g., 1.5, 1:30, 90m)",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	return ParseHours(result)
}

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

func InputNotes() (string, error) {
	prompt := promptui.Prompt{
		Label:   "Notes (optional)",
		Default: "",
	}

	return prompt.Run()
}

func InputDate() (string, error) {
	today := time.Now().Format("2006-01-02")

	validate := func(input string) error {
		if input == "" {
			return nil // Will use default
		}
		_, err := time.Parse("2006-01-02", input)
		if err != nil {
			return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Date (YYYY-MM-DD)",
		Default:  today,
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	if result == "" {
		return today, nil
	}
	return result, nil
}

func Confirm(message string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     message,
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, err
	}

	return strings.ToLower(result) == "y", nil
}

// FuzzyMatchProject finds projects matching the query string.
// Returns matched projects. If exactly one match, returns it directly.
// If multiple matches, caller should use SelectProject with filtered list.
func FuzzyMatchProject(projects []models.ProjectAssignment, query string) []models.ProjectAssignment {
	if query == "" {
		return projects
	}

	query = strings.ToLower(query)
	var matches []models.ProjectAssignment

	// First, check for exact match
	for _, p := range projects {
		if strings.ToLower(p.Project.Name) == query {
			return []models.ProjectAssignment{p}
		}
	}

	// Then, check for substring matches
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
	var matches []models.TaskAssignment

	// First, check for exact match
	for _, t := range activeTasks {
		if strings.ToLower(t.Task.Name) == query {
			return []models.TaskAssignment{t}
		}
	}

	// Then, check for substring matches
	for _, t := range activeTasks {
		name := strings.ToLower(t.Task.Name)
		if strings.Contains(name, query) {
			matches = append(matches, t)
		}
	}

	return matches
}
