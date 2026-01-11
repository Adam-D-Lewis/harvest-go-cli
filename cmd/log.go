package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"harvest-cli/internal/models"
	"harvest-cli/internal/prompt"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Log a new time entry",
	Long:  `Interactively log a new time entry by selecting project, task, hours, and notes.`,
	RunE:  runLog,
}

func runLog(cmd *cobra.Command, args []string) error {
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

	// Input hours
	hours, err := prompt.InputHours()
	if err != nil {
		return fmt.Errorf("hours input cancelled: %w", err)
	}

	// Input notes
	notes, err := prompt.InputNotes()
	if err != nil {
		return fmt.Errorf("notes input cancelled: %w", err)
	}

	// Input date
	date, err := prompt.InputDate()
	if err != nil {
		return fmt.Errorf("date input cancelled: %w", err)
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
