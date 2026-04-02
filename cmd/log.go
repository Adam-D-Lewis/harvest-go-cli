package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"harvest/internal/cache"
	"harvest/internal/models"
	"harvest/internal/prompt"
)

// Note: getProjectsForCompletion is defined in start.go

var logDate string

var logCmd = &cobra.Command{
	Use:   "log [project] [task] [hours] [notes]",
	Short: "Log a new time entry",
	Long: `Log a new time entry by selecting project, task, hours, and notes.

Arguments are optional and support fuzzy matching:
  harvest log                                   # Interactive selection
  harvest log "myproject" "dev" 2.5             # Fuzzy match project/task
  harvest log "myproject" "dev" 2.5 "Standup"   # With notes`,
	RunE:              runLog,
	ValidArgsFunction: completeLogArgs,
}

func init() {
	logCmd.Flags().StringVar(&logDate, "date", "", "Date for the entry (YYYY-MM-DD), defaults to interactive prompt")
}

func completeLogArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projects := getProjectsForCompletion()
	if len(projects) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	switch len(args) {
	case 0:
		// Complete project names
		var completions []string
		for _, p := range projects {
			completions = append(completions, p.Project.Name)
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	case 1:
		// Complete task names
		projectQuery := args[0]
		matches := prompt.FuzzyMatchProject(projects, projectQuery)
		if len(matches) == 1 {
			var completions []string
			toCompleteLower := strings.ToLower(toComplete)
			for _, t := range matches[0].TaskAssignments {
				if t.IsActive {
					// Filter by substring match (case-insensitive)
					if toComplete == "" || strings.Contains(strings.ToLower(t.Task.Name), toCompleteLower) {
						completions = append(completions, t.Task.Name)
					}
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func runLog(cmd *cobra.Command, args []string) error {
	// Fetch projects
	fmt.Println("Fetching projects...")
	projects, err := apiClient.GetProjectAssignments()
	if err != nil {
		return fmt.Errorf("failed to fetch projects: %w", err)
	}

	// Cache projects for shell completion
	_ = cache.SaveProjects(projects)

	var project *models.ProjectAssignment
	var task *models.TaskAssignment
	var hours float64
	var notes string

	// Parse arguments
	projectQuery := ""
	taskQuery := ""
	hoursStr := ""
	if len(args) >= 1 {
		projectQuery = args[0]
	}
	if len(args) >= 2 {
		taskQuery = args[1]
	}
	if len(args) >= 3 {
		hoursStr = args[2]
	}
	if len(args) >= 4 {
		notes = args[3]
	}

	// Select/match project
	if projectQuery != "" {
		matches := prompt.FuzzyMatchProject(projects, projectQuery)
		if len(matches) == 0 {
			fmt.Printf("No project matching '%s', showing all projects...\n", projectQuery)
			p, err := prompt.SelectProject(projects)
			if err != nil {
				return fmt.Errorf("project selection cancelled: %w", err)
			}
			project = p
		} else if len(matches) == 1 {
			project = &matches[0]
			fmt.Printf("Matched project: %s\n", project.Project.Name)
		} else {
			fmt.Printf("Multiple projects match '%s', please select:\n", projectQuery)
			p, err := prompt.SelectProject(matches)
			if err != nil {
				return fmt.Errorf("project selection cancelled: %w", err)
			}
			project = p
		}
	} else {
		p, err := prompt.SelectProject(projects)
		if err != nil {
			return fmt.Errorf("project selection cancelled: %w", err)
		}
		project = p
	}

	// Select/match task
	if taskQuery != "" {
		matches := prompt.FuzzyMatchTask(project.TaskAssignments, taskQuery)
		if len(matches) == 0 {
			fmt.Printf("No task matching '%s', showing all tasks...\n", taskQuery)
			t, err := prompt.SelectTask(project.TaskAssignments)
			if err != nil {
				return fmt.Errorf("task selection cancelled: %w", err)
			}
			task = t
		} else if len(matches) == 1 {
			task = &matches[0]
			fmt.Printf("Matched task: %s\n", task.Task.Name)
		} else {
			fmt.Printf("Multiple tasks match '%s', please select:\n", taskQuery)
			t, err := prompt.SelectTask(matches)
			if err != nil {
				return fmt.Errorf("task selection cancelled: %w", err)
			}
			task = t
		}
	} else {
		t, err := prompt.SelectTask(project.TaskAssignments)
		if err != nil {
			return fmt.Errorf("task selection cancelled: %w", err)
		}
		task = t
	}

	// Parse/input hours
	if hoursStr != "" {
		h, err := prompt.ParseHours(hoursStr)
		if err != nil {
			return fmt.Errorf("invalid hours value: %s", hoursStr)
		}
		hours = h
		fmt.Printf("Hours: %.2f\n", hours)
	} else {
		h, err := prompt.InputHours()
		if err != nil {
			return fmt.Errorf("hours input cancelled: %w", err)
		}
		hours = h
	}

	// Input notes if not provided
	if notes == "" && len(args) < 4 {
		n, err := prompt.InputNotes()
		if err != nil {
			return fmt.Errorf("notes input cancelled: %w", err)
		}
		notes = n
	}

	var date string
	if logDate != "" {
		if _, err := time.Parse("2006-01-02", logDate); err != nil {
			return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		date = logDate
		fmt.Printf("Date: %s\n", date)
	} else {
		d, err := prompt.InputDate()
		if err != nil {
			return fmt.Errorf("date input cancelled: %w", err)
		}
		date = d
	}

	// Create time entry
	req := &models.CreateTimeEntryRequest{
		ProjectID: project.Project.ID,
		TaskID:    task.Task.ID,
		SpentDate: date,
		Hours:     hours,
		Notes:     notes,
	}

	fmt.Println("\nCreating time entry...")
	entry, err := apiClient.CreateTimeEntry(req)
	if err != nil {
		return fmt.Errorf("failed to create time entry: %w", err)
	}

	fmt.Printf("\nTime entry created successfully!\n")
	fmt.Printf("  Project: %s\n", entry.Project.Name)
	fmt.Printf("  Task:    %s\n", entry.Task.Name)
	fmt.Printf("  Date:    %s\n", entry.SpentDate)
	fmt.Printf("  Hours:   %.2f\n", entry.Hours)
	if entry.Notes != "" {
		fmt.Printf("  Notes:   %s\n", entry.Notes)
	}

	return nil
}
