package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"harvest-cli/internal/models"
	"harvest-cli/internal/prompt"
	"harvest-cli/internal/timer"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a timer",
	Long:  `Start a timer for a project and task. The timer runs on Harvest's servers.`,
	RunE:  runStart,
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

	// Select project
	project, err := prompt.SelectProject(projects)
	if err != nil {
		return fmt.Errorf("project selection cancelled: %w", err)
	}

	// Select task
	task, err := prompt.SelectTask(project.TaskAssignments)
	if err != nil {
		return fmt.Errorf("task selection cancelled: %w", err)
	}

	// Input notes
	notes, err := prompt.InputNotes()
	if err != nil {
		return fmt.Errorf("notes input cancelled: %w", err)
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
