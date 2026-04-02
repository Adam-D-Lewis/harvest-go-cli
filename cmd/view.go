package cmd

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"harvest/internal/models"
)

// ANSI color codes — high-contrast colors readable on dark backgrounds.
var projectColors = []string{
	"\033[92m",       // bright green
	"\033[93m",       // bright yellow
	"\033[96m",       // bright cyan
	"\033[38;5;81m",  // bright sky blue
	"\033[38;5;208m", // orange
	"\033[38;5;117m", // sky blue
	"\033[38;5;219m", // pink
	"\033[38;5;156m", // light lime
}

const colorReset = "\033[0m"

// colorForProject returns a consistent ANSI color for a given project name.
func colorForProject(name string) string {
	h := fnv.New32a()
	h.Write([]byte(name))
	return projectColors[int(h.Sum32())%len(projectColors)]
}

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
	Use:                "week [offset]",
	Short:              "View this week's time entries (use -1 for last week, -2 for 2 weeks ago, etc.)",
	RunE:               runViewWeek,
	DisableFlagParsing: true,
}

var viewMonthCmd = &cobra.Command{
	Use:                "month [offset]",
	Short:              "View this month's time entries (use -1 for last month, -2 for 2 months ago, etc.)",
	RunE:               runViewMonth,
	DisableFlagParsing: true,
}

var (
	fromDate string
	toDate   string
)

func init() {
	viewCmd.AddCommand(viewTodayCmd)
	viewCmd.AddCommand(viewWeekCmd)
	viewCmd.AddCommand(viewMonthCmd)

	viewCmd.Flags().StringVar(&fromDate, "from", "", "Start date (YYYY-MM-DD)")
	viewCmd.Flags().StringVar(&toDate, "to", "", "End date (YYYY-MM-DD)")
	viewCmd.RunE = runViewRange
}

func runViewToday(cmd *cobra.Command, args []string) error {
	today := time.Now().Format("2006-01-02")
	return viewEntries(today, today, "Today")
}

func runViewWeek(cmd *cobra.Command, args []string) error {
	offset := 0
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid week offset %q — use a number like -1, -2", args[0])
		}
		offset = n
	}

	now := time.Now()
	// Get start of week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday
	}
	monday := now.AddDate(0, 0, -(weekday-1)+(offset*7))
	sunday := monday.AddDate(0, 0, 6)

	// Fetch Mon–Sun to check for weekend entries
	from := monday.Format("2006-01-02")
	to := sunday.Format("2006-01-02")

	entries, err := apiClient.GetTimeEntries(from, to)
	if err != nil {
		return fmt.Errorf("failed to fetch time entries: %w", err)
	}

	// Check if any entries fall on Saturday or Sunday
	hasWeekend := false
	for _, e := range entries {
		t, _ := time.Parse("2006-01-02", e.SpentDate)
		if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
			hasWeekend = true
			break
		}
	}

	// Filter to Mon–Fri unless weekend has entries
	if !hasWeekend {
		friday := monday.AddDate(0, 0, 4).Format("2006-01-02")
		var filtered []models.TimeEntry
		for _, e := range entries {
			if e.SpentDate <= friday {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	return renderGroupedFromEntries(entries)
}

func runViewMonth(cmd *cobra.Command, args []string) error {
	offset := 0
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid month offset %q — use a number like -1, -2", args[0])
		}
		offset = n
	}

	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	firstOfMonth = firstOfMonth.AddDate(0, offset, 0)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	from := firstOfMonth.Format("2006-01-02")
	to := lastOfMonth.Format("2006-01-02")
	label := firstOfMonth.Format("January 2006")

	return viewEntriesByWeek(from, to, label)
}

// parseDate accepts:
//   - YYYY-MM-DD (standard date)
//   - relative days: -7, +4, -1, +0
//   - day abbreviations: mon, tue, wed, thu, fri, sat, sun (most recent occurrence)
//   - "today", "yesterday"
func parseDate(s string) (string, error) {
	now := time.Now()

	// Standard date format
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return s, nil
	}

	lower := strings.ToLower(strings.TrimSpace(s))

	// today / yesterday
	if lower == "today" {
		return now.Format("2006-01-02"), nil
	}
	if lower == "yesterday" {
		return now.AddDate(0, 0, -1).Format("2006-01-02"), nil
	}

	// Relative days: -7, +4, 4, etc.
	if n, err := strconv.Atoi(lower); err == nil {
		return now.AddDate(0, 0, n).Format("2006-01-02"), nil
	}

	// Day name abbreviations
	dayMap := map[string]time.Weekday{
		"sun": time.Sunday, "sunday": time.Sunday,
		"mon": time.Monday, "monday": time.Monday,
		"tue": time.Tuesday, "tuesday": time.Tuesday,
		"wed": time.Wednesday, "wednesday": time.Wednesday,
		"thu": time.Thursday, "thursday": time.Thursday,
		"fri": time.Friday, "friday": time.Friday,
		"sat": time.Saturday, "saturday": time.Saturday,
	}

	if target, ok := dayMap[lower]; ok {
		// Find the most recent occurrence of this day (including today)
		current := now.Weekday()
		diff := int(current) - int(target)
		if diff < 0 {
			diff += 7
		}
		return now.AddDate(0, 0, -diff).Format("2006-01-02"), nil
	}

	return "", fmt.Errorf("unrecognized date format %q — use YYYY-MM-DD, -7/+4, mon/tue/.../fri, today, or yesterday", s)
}

func runViewRange(cmd *cobra.Command, args []string) error {
	// Default to this week if no flags provided
	if fromDate == "" && toDate == "" {
		return runViewWeek(cmd, args)
	}
	if fromDate == "" || toDate == "" {
		return fmt.Errorf("both --from and --to flags are required for custom date range")
	}

	from, err := parseDate(fromDate)
	if err != nil {
		return fmt.Errorf("invalid --from: %w", err)
	}
	to, err := parseDate(toDate)
	if err != nil {
		return fmt.Errorf("invalid --to: %w", err)
	}

	return viewEntries(from, to, fmt.Sprintf("%s to %s", from, to))
}

func viewEntries(from, to, label string) error {
	fmt.Printf("Fetching time entries for %s...\n\n", label)

	entries, err := apiClient.GetTimeEntries(from, to)
	if err != nil {
		return fmt.Errorf("failed to fetch time entries: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No time entries found.")
		return nil
	}

	renderGrouped(entries)
	return nil
}

func renderGroupedFromEntries(entries []models.TimeEntry) error {
	if len(entries) == 0 {
		fmt.Println("No time entries found.")
		return nil
	}

	renderGrouped(entries)
	return nil
}

// formatHours formats hours as a right-aligned string with optional running indicator.
func formatHours(hours float64, running bool) string {
	s := fmt.Sprintf("%5.2f", hours)
	if running {
		s += "*"
	} else {
		s += " "
	}
	return s
}

// renderGrouped displays time entries grouped by day with daily subtotals and a grand total.
func renderGrouped(entries []models.TimeEntry) {
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

	var grandTotal float64
	const lineWidth = 66

	for i, date := range dates {
		dayEntries := byDate[date]

		// Calculate day total
		var dayTotal float64
		for _, e := range dayEntries {
			dayTotal += e.Hours
		}

		// Day header: ── Monday, Mar 30 ──────────────────── 8.00 hrs ──
		t, _ := time.Parse("2006-01-02", date)
		dayName := t.Format("Monday, Jan 2")
		totalStr := fmt.Sprintf("%.2f hrs", dayTotal)
		prefix := "── " + dayName + " "
		suffix := " " + totalStr + " ──"
		padLen := lineWidth - len(prefix) - len(suffix)
		if padLen < 2 {
			padLen = 2
		}
		pad := ""
		for j := 0; j < padLen; j++ {
			pad += "─"
		}
		fmt.Printf("%s%s%s\n", prefix, pad, suffix)

		// Entries
		for _, e := range dayEntries {
			hoursStr := formatHours(e.Hours, e.IsRunning)

			lockedStr := ""
			if e.IsLocked {
				lockedStr = " 🔒"
				if e.LockedReason != "" {
					lockedStr += " " + e.LockedReason
				}
			}

			color := colorForProject(e.Project.Name)
			projectTask := e.Project.Name + " / " + e.Task.Name
			padWidth := 40
			padding := padWidth - len(projectTask)
			if padding < 1 {
				padding = 1
			}
			fmt.Printf("  %s hrs  %s%s%s%*s[%d]%s\n", hoursStr, color, projectTask, colorReset, padding, "", e.ID, lockedStr)
			if e.Notes != "" {
				fmt.Printf("          📝 %s\n", e.Notes)
			}
		}

		grandTotal += dayTotal

		// Blank line between days, but not after the last
		if i < len(dates)-1 {
			fmt.Println()
		}
	}

	// Grand total
	divider := ""
	for j := 0; j < lineWidth; j++ {
		divider += "─"
	}
	fmt.Printf("\n%s\n", divider)
	fmt.Printf("  Total: %.2f hrs (%d entries)\n", grandTotal, len(entries))
}

// mondayOf returns the Monday of the week containing the given date.
func mondayOf(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -(weekday - 1))
}

// viewEntriesByWeek fetches entries and renders them summarized by week with project breakdowns.
func viewEntriesByWeek(from, to, label string) error {
	fmt.Printf("Fetching time entries for %s...\n\n", label)

	entries, err := apiClient.GetTimeEntries(from, to)
	if err != nil {
		return fmt.Errorf("failed to fetch time entries: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No time entries found.")
		return nil
	}

	// Group entries by week (Monday date)
	type projectTask struct {
		project string
		task    string
	}
	type weekData struct {
		monday   time.Time
		lastDate time.Time
		hours    map[projectTask]float64
		total    float64
	}

	weekMap := make(map[string]*weekData)
	var weekKeys []string

	for _, e := range entries {
		t, _ := time.Parse("2006-01-02", e.SpentDate)
		mon := mondayOf(t)
		key := mon.Format("2006-01-02")

		wd, ok := weekMap[key]
		if !ok {
			wd = &weekData{
				monday: mon,
				hours:  make(map[projectTask]float64),
			}
			weekMap[key] = wd
			weekKeys = append(weekKeys, key)
		}

		if t.After(wd.lastDate) {
			wd.lastDate = t
		}

		pt := projectTask{project: e.Project.Name, task: e.Task.Name}
		wd.hours[pt] += e.Hours
		wd.total += e.Hours
	}

	sort.Strings(weekKeys)

	const lineWidth = 66
	var grandTotal float64

	for i, key := range weekKeys {
		wd := weekMap[key]

		// Week header — show actual date range
		friday := wd.monday.AddDate(0, 0, 4)
		endDate := wd.lastDate
		if friday.Before(endDate) {
			endDate = friday
		}
		weekLabel := wd.monday.Format("Jan 2") + " – " + endDate.Format("Jan 2")
		totalStr := fmt.Sprintf("%.2f hrs", wd.total)
		prefix := "── " + weekLabel + " "
		suffix := " " + totalStr + " ──"
		padLen := lineWidth - len(prefix) - len(suffix)
		if padLen < 2 {
			padLen = 2
		}
		pad := ""
		for j := 0; j < padLen; j++ {
			pad += "─"
		}
		fmt.Printf("%s%s%s\n", prefix, pad, suffix)

		// Sort project/tasks by hours descending
		type ptHours struct {
			pt    projectTask
			hours float64
		}
		var items []ptHours
		for pt, h := range wd.hours {
			items = append(items, ptHours{pt: pt, hours: h})
		}
		sort.Slice(items, func(a, b int) bool {
			return items[a].hours > items[b].hours
		})

		for _, item := range items {
			color := colorForProject(item.pt.project)
			name := item.pt.project + " / " + item.pt.task
			fmt.Printf("  %5.2f  hrs  %s%s%s\n", item.hours, color, name, colorReset)
		}

		grandTotal += wd.total

		if i < len(weekKeys)-1 {
			fmt.Println()
		}
	}

	divider := ""
	for j := 0; j < lineWidth; j++ {
		divider += "─"
	}
	fmt.Printf("\n%s\n", divider)
	fmt.Printf("  Total: %.2f hrs\n", grandTotal)

	return nil
}
