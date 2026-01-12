package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"harvest-cli/internal/models"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View time entries",
	Long:  `View your logged time entries for today, this week, or a custom date range.`,
}

var viewTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "View today's time entries",
	RunE:  runViewToday,
}

var viewWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "View this week's time entries",
	RunE:  runViewWeek,
}

var (
	fromDate string
	toDate   string
)

func init() {
	viewCmd.AddCommand(viewTodayCmd)
	viewCmd.AddCommand(viewWeekCmd)

	viewCmd.Flags().StringVar(&fromDate, "from", "", "Start date (YYYY-MM-DD)")
	viewCmd.Flags().StringVar(&toDate, "to", "", "End date (YYYY-MM-DD)")
	viewCmd.RunE = runViewRange
}

func runViewToday(cmd *cobra.Command, args []string) error {
	today := time.Now().Format("2006-01-02")
	return viewEntries(today, today, "Today")
}

func runViewWeek(cmd *cobra.Command, args []string) error {
	now := time.Now()
	// Get start of week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday
	}
	monday := now.AddDate(0, 0, -(weekday - 1))

	from := monday.Format("2006-01-02")
	to := now.Format("2006-01-02")

	return viewEntriesGroupedByDay(from, to)
}

func runViewRange(cmd *cobra.Command, args []string) error {
	if fromDate == "" || toDate == "" {
		return fmt.Errorf("both --from and --to flags are required for custom date range")
	}

	// Validate dates
	if _, err := time.Parse("2006-01-02", fromDate); err != nil {
		return fmt.Errorf("invalid from date format. Use YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", toDate); err != nil {
		return fmt.Errorf("invalid to date format. Use YYYY-MM-DD")
	}

	return viewEntries(fromDate, toDate, fmt.Sprintf("%s to %s", fromDate, toDate))
}

func viewEntries(from, to, label string) error {
	fmt.Printf("Fetching time entries for %s...\n", label)

	entries, err := apiClient.GetTimeEntries(from, to)
	if err != nil {
		return fmt.Errorf("failed to fetch time entries: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No time entries found.")
		return nil
	}

	renderEntriesTable(entries)
	return nil
}

func viewEntriesGroupedByDay(from, to string) error {
	fmt.Println("Fetching time entries for this week...")

	entries, err := apiClient.GetTimeEntries(from, to)
	if err != nil {
		return fmt.Errorf("failed to fetch time entries: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No time entries found this week.")
		return nil
	}

	// Group by date
	byDate := make(map[string][]models.TimeEntry)
	for _, e := range entries {
		byDate[e.SpentDate] = append(byDate[e.SpentDate], e)
	}

	// Sort dates
	var dates []string
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	var weekTotal float64

	for _, date := range dates {
		dayEntries := byDate[date]

		// Parse date for display
		t, _ := time.Parse("2006-01-02", date)
		dayName := t.Format("Monday, Jan 2")

		fmt.Printf("\n%s\n", dayName)
		fmt.Println("────────────────────────────────────────────────────────────────────────────────")

		var dayTotal float64
		for _, e := range dayEntries {
			hoursStr := fmt.Sprintf("%.2f", e.Hours)
			if e.IsRunning {
				hoursStr += "*"
			}

			fmt.Printf("  %s hrs  %s / %s\n", hoursStr, e.Project.Name, e.Task.Name)
			if e.Notes != "" {
				fmt.Printf("           %s\n", e.Notes)
			}
			dayTotal += e.Hours
		}

		fmt.Printf("                                                            Day total: %.2f hrs\n", dayTotal)
		weekTotal += dayTotal
	}

	fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("                                                           Week total: %.2f hrs\n", weekTotal)

	return nil
}

func renderEntriesTable(entries []models.TimeEntry) {
	fmt.Println()

	var total float64
	for _, e := range entries {
		hoursStr := fmt.Sprintf("%.2f", e.Hours)
		if e.IsRunning {
			hoursStr += "*"
		}

		fmt.Printf("%s  %s hrs  %s / %s\n", e.SpentDate, hoursStr, e.Project.Name, e.Task.Name)
		if e.Notes != "" {
			fmt.Printf("                  %s\n", e.Notes)
		}
		total += e.Hours
	}

	fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("Total: %.2f hours (%d entries)\n", total, len(entries))
}
