package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/Adam-D-Lewis/harvest-go-cli/internal/cache"
	"github.com/Adam-D-Lewis/harvest-go-cli/internal/models"
	"github.com/Adam-D-Lewis/harvest-go-cli/internal/prompt"
)

var (
	editDate    string
	editID      int
	editHours   string
	editNotes   string
	editProject string
	editTask    string
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a time entry",
	Long: `Edit an existing time entry. Shows today's entries by default.

  harvest edit                                    # Interactive: select and edit
  harvest edit --date 2026-03-27                  # Interactive: from a specific day
  harvest edit --id 12345678 --hours 2.5          # Non-interactive: update hours
  harvest edit --id 12345678 --notes "Standup"    # Non-interactive: update notes
  harvest edit --id 12345678 --project "My Proj" --task "Dev"  # Update project/task`,
	RunE: runEdit,
}

func init() {
	editCmd.Flags().StringVar(&editDate, "date", "", "Date to show entries for (YYYY-MM-DD)")
	editCmd.Flags().IntVar(&editID, "id", 0, "Time entry ID to edit (skips interactive selection)")
	editCmd.Flags().StringVar(&editHours, "hours", "", "New hours value (e.g. 1.5, 1:30, 90m)")
	editCmd.Flags().StringVar(&editNotes, "notes", "", "New notes")
	editCmd.Flags().StringVar(&editProject, "project", "", "New project (fuzzy match)")
	editCmd.Flags().StringVar(&editTask, "task", "", "New task (fuzzy match, requires --project)")
}

func runEdit(cmd *cobra.Command, args []string) error {
	if editID != 0 {
		return runEditNonInteractive(cmd)
	}
	return runEditInteractive()
}

func runEditNonInteractive(cmd *cobra.Command) error {
	notesChanged := cmd.Flags().Changed("notes")
	if editHours == "" && !notesChanged && editProject == "" && editDate == "" {
		return fmt.Errorf("--id requires at least one field to update: --hours, --notes, --project/--task, or --date")
	}

	// Fetch entry to verify it exists and check lock status
	fmt.Printf("Fetching entry %d...\n", editID)
	entry, err := apiClient.GetTimeEntry(editID)
	if err != nil {
		return fmt.Errorf("failed to fetch entry: %w", err)
	}

	if entry.IsLocked {
		return fmt.Errorf("entry %d is locked/approved and cannot be edited", editID)
	}

	req := &models.UpdateTimeEntryRequest{}

	// Hours
	if editHours != "" {
		h, err := prompt.ParseHours(editHours)
		if err != nil {
			return fmt.Errorf("invalid hours: %w", err)
		}
		req.Hours = &h
	}

	// Notes
	if notesChanged {
		req.Notes = &editNotes
	}

	// Date
	if editDate != "" {
		if _, err := time.Parse("2006-01-02", editDate); err != nil {
			return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		req.SpentDate = &editDate
	}

	// Project & Task
	if editProject != "" {
		if editTask == "" {
			return fmt.Errorf("--project requires --task")
		}

		fmt.Println("Fetching projects...")
		projects, err := apiClient.GetProjectAssignments()
		if err != nil {
			return fmt.Errorf("failed to fetch projects: %w", err)
		}
		_ = cache.SaveProjects(projects)

		matches := prompt.FuzzyMatchProject(projects, editProject)
		if len(matches) == 0 {
			return fmt.Errorf("no project matching '%s'", editProject)
		}
		if len(matches) > 1 {
			fmt.Fprintf(os.Stderr, "Multiple projects match '%s':\n", editProject)
			for _, m := range matches {
				fmt.Fprintf(os.Stderr, "  - %s (%s)\n", m.Project.Name, m.Client.Name)
			}
			return fmt.Errorf("ambiguous project match, be more specific")
		}
		project := matches[0]

		taskMatches := prompt.FuzzyMatchTask(project.TaskAssignments, editTask)
		if len(taskMatches) == 0 {
			return fmt.Errorf("no task matching '%s' in project '%s'", editTask, project.Project.Name)
		}
		if len(taskMatches) > 1 {
			fmt.Fprintf(os.Stderr, "Multiple tasks match '%s':\n", editTask)
			for _, m := range taskMatches {
				fmt.Fprintf(os.Stderr, "  - %s\n", m.Task.Name)
			}
			return fmt.Errorf("ambiguous task match, be more specific")
		}

		req.ProjectID = &project.Project.ID
		req.TaskID = &taskMatches[0].Task.ID
	} else if editTask != "" {
		return fmt.Errorf("--task requires --project")
	}

	// Apply update
	fmt.Println("Updating time entry...")
	updated, err := apiClient.UpdateTimeEntry(editID, req)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	fmt.Printf("Time entry updated!\n")
	fmt.Printf("  Project: %s\n", updated.Project.Name)
	fmt.Printf("  Task:    %s\n", updated.Task.Name)
	fmt.Printf("  Date:    %s\n", updated.SpentDate)
	fmt.Printf("  Hours:   %.2f\n", updated.Hours)
	if updated.Notes != "" {
		fmt.Printf("  Notes:   %s\n", updated.Notes)
	}

	return nil
}

func runEditInteractive() error {
	// Determine date
	date := time.Now().Format("2006-01-02")
	if editDate != "" {
		if _, err := time.Parse("2006-01-02", editDate); err != nil {
			return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		date = editDate
	}

	// Fetch entries for the date
	fmt.Printf("Fetching entries for %s...\n", date)
	entries, err := apiClient.GetTimeEntries(date, date)
	if err != nil {
		return fmt.Errorf("failed to fetch entries: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No time entries found for this date.")
		return nil
	}

	// Select entry
	entry, err := prompt.SelectTimeEntry(entries)
	if err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	if entry.IsLocked {
		fmt.Println("This entry is locked/approved and cannot be edited.")
		return nil
	}

	fmt.Printf("\nEditing: %.2f hrs  %s / %s\n", entry.Hours, entry.Project.Name, entry.Task.Name)
	if entry.Notes != "" {
		fmt.Printf("  Notes: %s\n", entry.Notes)
	}

	// Edit loop - let user change multiple fields
	req := &models.UpdateTimeEntryRequest{}
	changed := false

	for {
		field, err := prompt.SelectEditField()
		if err != nil {
			return fmt.Errorf("field selection cancelled: %w", err)
		}

		if field == "done" {
			break
		}

		switch field {
		case "hours":
			h, err := prompt.InputHours()
			if err != nil {
				return fmt.Errorf("hours input cancelled: %w", err)
			}
			req.Hours = &h
			changed = true

		case "notes":
			n, err := prompt.InputNotes()
			if err != nil {
				return fmt.Errorf("notes input cancelled: %w", err)
			}
			req.Notes = &n
			changed = true

		case "date":
			d, err := prompt.InputDate()
			if err != nil {
				return fmt.Errorf("date input cancelled: %w", err)
			}
			req.SpentDate = &d
			changed = true

		case "project":
			fmt.Println("Fetching projects...")
			projects, err := apiClient.GetProjectAssignments()
			if err != nil {
				return fmt.Errorf("failed to fetch projects: %w", err)
			}
			_ = cache.SaveProjects(projects)

			p, err := prompt.SelectProject(projects)
			if err != nil {
				return fmt.Errorf("project selection cancelled: %w", err)
			}
			t, err := prompt.SelectTask(p.TaskAssignments)
			if err != nil {
				return fmt.Errorf("task selection cancelled: %w", err)
			}
			req.ProjectID = &p.Project.ID
			req.TaskID = &t.Task.ID
			changed = true
		}
	}

	if !changed {
		fmt.Println("No changes made.")
		return nil
	}

	// Apply the update
	fmt.Println("\nUpdating time entry...")
	updated, err := apiClient.UpdateTimeEntry(entry.ID, req)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	fmt.Printf("\nTime entry updated!\n")
	fmt.Printf("  Project: %s\n", updated.Project.Name)
	fmt.Printf("  Task:    %s\n", updated.Task.Name)
	fmt.Printf("  Date:    %s\n", updated.SpentDate)
	fmt.Printf("  Hours:   %.2f\n", updated.Hours)
	if updated.Notes != "" {
		fmt.Printf("  Notes:   %s\n", updated.Notes)
	}

	return nil
}
