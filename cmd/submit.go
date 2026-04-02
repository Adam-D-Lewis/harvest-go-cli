package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:                "submit [week offset]",
	Short:              "Submit timesheet for approval (opens Harvest in browser)",
	RunE:               runSubmit,
	DisableFlagParsing: true,
}

func runSubmit(cmd *cobra.Command, args []string) error {
	offset := 0
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid week offset %q — use a number like -1, -2", args[0])
		}
		offset = n
	}

	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday-1)+(offset*7))
	friday := monday.AddDate(0, 0, 4)

	from := monday.Format("2006-01-02")
	to := friday.Format("2006-01-02")

	entries, err := apiClient.GetTimeEntries(from, to)
	if err != nil {
		return fmt.Errorf("failed to fetch time entries: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No time entries found for this week. Nothing to submit.")
		return nil
	}

	// Check if already submitted/locked
	allLocked := true
	for _, e := range entries {
		if !e.IsLocked {
			allLocked = false
			break
		}
	}

	if allLocked {
		fmt.Println("All entries for this week are already locked/approved.")
		return nil
	}

	// Show summary
	fmt.Printf("Week of %s to %s\n\n", monday.Format("Mon, Jan 2"), friday.Format("Mon, Jan 2"))
	renderGrouped(entries, false)

	// Get the company's Harvest URL
	baseURI, err := apiClient.GetCompanyBaseURI()
	if err != nil {
		return fmt.Errorf("failed to get Harvest URL: %w", err)
	}

	harvestURL := fmt.Sprintf("%s/time/week/%s/%s",
		baseURI,
		monday.Format("2006"),
		monday.Format("01/02"))

	fmt.Printf("\nOpening Harvest to submit: %s\n", harvestURL)
	fmt.Println("(The Harvest API does not support submitting timesheets — please click Submit in the browser.)")
	return openBrowser(harvestURL)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform — open this URL manually: %s", url)
	}
	return cmd.Start()
}
