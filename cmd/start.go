package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"harvest-cli/internal/cache"
	"harvest-cli/internal/models"
	"harvest-cli/internal/prompt"
	"harvest-cli/internal/timer"
)

var startCmd = &cobra.Command{
	Use:   "start [project] [task] [notes]",
	Short: "Start a timer",
	Long: `Start a timer for a project and task. The timer runs on Harvest's servers.

Arguments are optional and support fuzzy matching:
  harvest start                          # Interactive selection
  harvest start "myproject"              # Fuzzy match project, interactive task
  harvest start "myproject" "dev"        # Fuzzy match both
  harvest start "myproject" "dev" "note" # With notes`,
	RunE:              runStart,
	ValidArgsFunction: completeStartArgs,
}

func completeStartArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
		// Complete task names for the selected project
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

// getProjectsForCompletion returns projects from cache, or fetches from API if cache is empty
func getProjectsForCompletion() []models.ProjectAssignment {
	// Try cache first
	projects, valid := cache.LoadProjects()
	if valid && len(projects) > 0 {
		return projects
	}

	// Cache empty or stale - fetch from API
	if err := initAPIClient(); err != nil {
		return nil
	}

	projects, err := apiClient.GetProjectAssignments()
	if err != nil {
		return nil
	}

	// Update cache
	cache.SaveProjects(projects)
	return projects
}

func runStart(cmd *cobra.Command, args []string) error {
	// Check if timer is already running
	state, err := timer.LoadTimerState()
	if err != nil {
		return fmt.Errorf("failed to check timer state: %w", err)
	}

	if state != nil {
		elapsed := time.Since(state.StartedAt).Round(time.Minute)
		fmt.Printf("Timer already running for %s\n", elapsed)
		fmt.Printf("  Project: %s\n", state.Project)
		fmt.Printf("  Task:    %s\n", state.Task)
		fmt.Println("\nStop the current timer first with: harvest stop")
		return nil
	}

	// Fetch projects
	fmt.Println("Fetching projects...")
	projects, err := apiClient.GetProjectAssignments()
	if err != nil {
		return fmt.Errorf("failed to fetch projects: %w", err)
	}

	// Cache projects for shell completion
	cache.SaveProjects(projects)

	var project *models.ProjectAssignment
	var task *models.TaskAssignment
	var notes string

	// Parse arguments
	projectQuery := ""
	taskQuery := ""
	if len(args) >= 1 {
		projectQuery = args[0]
	}
	if len(args) >= 2 {
		taskQuery = args[1]
	}
	if len(args) >= 3 {
		notes = args[2]
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

	// Input notes if not provided
	if notes == "" && len(args) < 3 {
		n, err := prompt.InputNotes()
		if err != nil {
			return fmt.Errorf("notes input cancelled: %w", err)
		}
		notes = n
	}

	// Create time entry without hours (starts the timer)
	today := time.Now().Format("2006-01-02")
	req := &models.CreateTimeEntryRequest{
		ProjectID: project.Project.ID,
		TaskID:    task.Task.ID,
		SpentDate: today,
		Notes:     notes,
	}

	fmt.Println("\nStarting timer...")
	entry, err := apiClient.CreateTimeEntry(req)
	if err != nil {
		return fmt.Errorf("failed to start timer: %w", err)
	}

	// Save timer state locally
	timerState := &timer.TimerState{
		EntryID:   entry.ID,
		ProjectID: project.Project.ID,
		TaskID:    task.Task.ID,
		StartedAt: time.Now(),
		Notes:     notes,
		Project:   project.Project.Name,
		Task:      task.Task.Name,
	}

	if err := timer.SaveTimerState(timerState); err != nil {
		fmt.Printf("Warning: failed to save local timer state: %v\n", err)
	}

	fmt.Printf("\nTimer started!\n")
	fmt.Printf("  Project: %s\n", project.Project.Name)
	fmt.Printf("  Task:    %s\n", task.Task.Name)
	if notes != "" {
		fmt.Printf("  Notes:   %s\n", notes)
	}
	fmt.Println("\nStop the timer with: harvest stop")

	return nil
}
