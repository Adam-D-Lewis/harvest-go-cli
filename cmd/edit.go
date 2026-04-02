package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"harvest-cli/internal/cache"
	"harvest-cli/internal/models"
	"harvest-cli/internal/prompt"
)

var editDate string

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a time entry",
	Long: `Edit an existing time entry. Shows today's entries by default.

  harvest edit                    # Edit from today's entries
  harvest edit --date 2026-03-27  # Edit from a specific day's entries`,
	RunE: runEdit,
}

func init() {
	editCmd.Flags().StringVar(&editDate, "date", "", "Date to show entries for (YYYY-MM-DD)")
}

func runEdit(cmd *cobra.Command, args []string) error {
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
