package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Adam-D-Lewis/harvest-go-cli/internal/prompt"
)

var deleteDate string
var deleteID int

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a time entry",
	Long: `Delete an existing time entry. Shows today's entries by default.

  harvest delete                    # Interactive: select from today's entries
  harvest delete --date 2026-03-27  # Interactive: select from a specific day
  harvest delete --id 12345678      # Non-interactive: delete by entry ID`,
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().StringVar(&deleteDate, "date", "", "Date to show entries for (YYYY-MM-DD)")
	deleteCmd.Flags().IntVar(&deleteID, "id", 0, "Time entry ID to delete (skips interactive selection)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteID != 0 {
		return runDeleteNonInteractive()
	}
	return runDeleteInteractive()
}

func runDeleteNonInteractive() error {
	fmt.Printf("Fetching entry %d...\n", deleteID)
	entry, err := apiClient.GetTimeEntry(deleteID)
	if err != nil {
		return fmt.Errorf("failed to fetch entry: %w", err)
	}

	if entry.IsLocked {
		return fmt.Errorf("entry %d is locked/approved and cannot be deleted", deleteID)
	}

	fmt.Println("Deleting time entry...")
	if err := apiClient.DeleteTimeEntry(deleteID); err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	fmt.Printf("Deleted: %.2f hrs  %s / %s", entry.Hours, entry.Project.Name, entry.Task.Name)
	if entry.Notes != "" {
		fmt.Printf("  [%s]", entry.Notes)
	}
	fmt.Println()
	return nil
}

func runDeleteInteractive() error {
	// Determine date
	date := time.Now().Format("2006-01-02")
	if deleteDate != "" {
		if _, err := time.Parse("2006-01-02", deleteDate); err != nil {
			return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		date = deleteDate
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
		fmt.Println("This entry is locked/approved and cannot be deleted.")
		return nil
	}

	// Confirm deletion
	fmt.Printf("\nAbout to delete: %.2f hrs  %s / %s\n", entry.Hours, entry.Project.Name, entry.Task.Name)
	if entry.Notes != "" {
		fmt.Printf("  Notes: %s\n", entry.Notes)
	}

	confirmed, err := prompt.Confirm("Delete this entry?")
	if err != nil {
		return fmt.Errorf("confirmation error: %w", err)
	}
	if !confirmed {
		fmt.Println("Cancelled.")
		return nil
	}

	// Delete
	fmt.Println("Deleting time entry...")
	if err := apiClient.DeleteTimeEntry(entry.ID); err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	fmt.Println("Time entry deleted.")
	return nil
}
