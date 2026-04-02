package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"harvest/internal/timer"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running timer",
	Long:  `Stop the currently running timer and log the elapsed time.`,
	RunE:  runStop,
}

func runStop(cmd *cobra.Command, args []string) error {
	// Load timer state
	state, err := timer.LoadTimerState()
	if err != nil {
		return fmt.Errorf("failed to load timer state: %w", err)
	}

	if state == nil {
		fmt.Println("No timer is currently running.")
		fmt.Println("Start a timer with: harvest start")
		return nil
	}

	elapsed := time.Since(state.StartedAt).Round(time.Minute)
	fmt.Printf("Stopping timer (running for %s)...\n", elapsed)

	// Stop the timer via API
	entry, err := apiClient.StopTimer(state.EntryID)
	if err != nil {
		// Clear local state
		_ = timer.ClearTimerState()

		// Check if timer was deleted externally (404)
		if strings.Contains(err.Error(), "404") {
			fmt.Printf("\nTimer was already stopped or deleted externally.\n")
			fmt.Printf("Local timer state cleared.\n")
			return nil
		}
		return fmt.Errorf("failed to stop timer: %w", err)
	}

	// Clear local timer state
	if err := timer.ClearTimerState(); err != nil {
		fmt.Printf("Warning: failed to clear local timer state: %v\n", err)
	}

	fmt.Printf("\nTimer stopped!\n")
	fmt.Printf("  Project: %s\n", entry.Project.Name)
	fmt.Printf("  Task:    %s\n", entry.Task.Name)
	fmt.Printf("  Hours:   %.2f\n", entry.Hours)
	if entry.Notes != "" {
		fmt.Printf("  Notes:   %s\n", entry.Notes)
	}

	return nil
}
