package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"harvest/internal/prompt"
)

var deleteDate string

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a time entry",
	Long: `Delete an existing time entry. Shows today's entries by default.

  harvest delete                    # Delete from today's entries
  harvest delete --date 2026-03-27  # Delete from a specific day's entries`,
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().StringVar(&deleteDate, "date", "", "Date to show entries for (YYYY-MM-DD)")
}

func runDelete(cmd *cobra.Command, args []string) error {
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
