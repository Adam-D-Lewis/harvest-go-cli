package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"harvest-cli/internal/timer"
)

const weeklyTargetHours = 36.0

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status for tmux/status bar",
	Long:  `Output a compact status line showing current timer and weekly progress.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Get weekly hours
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	from := monday.Format("2006-01-02")
	to := now.Format("2006-01-02")

	entries, err := apiClient.GetTimeEntries(from, to)
	if err != nil {
		fmt.Print("harvest: error")
		return nil
	}

	var weekTotal float64
	for _, e := range entries {
		weekTotal += e.Hours
	}

	pct := (weekTotal / weeklyTargetHours) * 100
	if pct > 100 {
		pct = 100
	}

	// Check if timer is running
	state, _ := timer.LoadTimerState()

	if state != nil {
		elapsed := time.Since(state.StartedAt)
		elapsedHours := elapsed.Hours()
		totalWithTimer := weekTotal + elapsedHours

		pctWithTimer := (totalWithTimer / weeklyTargetHours) * 100
		if pctWithTimer > 100 {
			pctWithTimer = 100
		}

		// Format: "> Project/Task (1h23m) | 25.5h/36h (71%)"
		fmt.Printf("> %s/%s (%s) | %.1fh/%.0fh (%.0f%%)\n",
			state.Project,
			state.Task,
			formatDuration(elapsed),
			totalWithTimer,
			weeklyTargetHours,
			pctWithTimer,
		)
	} else {
		// Format: "25.5h/36h (71%)"
		fmt.Printf("%.1fh/%.0fh (%.0f%%)\n",
			weekTotal,
			weeklyTargetHours,
			pct,
		)
	}

	return nil
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	m := (d % time.Hour) / time.Minute

	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
