# Edit, Delete & Backdate Commands Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `harvest edit`, `harvest delete`, and `--date` flag to `harvest log` so users can manage existing time entries from the CLI.

**Architecture:** Three new features layered bottom-up: first add API methods (UpdateTimeEntry, DeleteTimeEntry) and the IsLocked model field, then build the two new commands on top, then modify the existing log command. Each command follows the existing Cobra + Bubbletea interactive pattern.

**Tech Stack:** Go, Cobra, Charmbracelet Bubbletea/Lipgloss, Harvest API v2

---

### Task 1: Add `IsLocked` field to TimeEntry model

**Files:**
- Modify: `internal/models/time_entry.go:5-16`

The Harvest API returns `is_locked` on time entries. We need this to warn users before attempting to edit/delete approved entries.

- [ ] **Step 1: Add IsLocked field**

In `internal/models/time_entry.go`, add `IsLocked` to the `TimeEntry` struct:

```go
type TimeEntry struct {
	ID        int       `json:"id"`
	SpentDate string    `json:"spent_date"`
	Hours     float64   `json:"hours"`
	Notes     string    `json:"notes"`
	IsRunning bool      `json:"is_running"`
	IsLocked  bool      `json:"is_locked"`
	Project   Project   `json:"project"`
	Task      Task      `json:"task"`
	Client    Client    `json:"client"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/balast/CodingProjects/harvest-go-cli && go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
cd /home/balast/CodingProjects/harvest-go-cli
git add internal/models/time_entry.go
git commit -m "feat: add IsLocked field to TimeEntry model"
```

---

### Task 2: Add `UpdateTimeEntryRequest` model

**Files:**
- Modify: `internal/models/time_entry.go`

- [ ] **Step 1: Add the update request struct**

Append to `internal/models/time_entry.go` after the `CreateTimeEntryRequest` struct:

```go
type UpdateTimeEntryRequest struct {
	ProjectID *int     `json:"project_id,omitempty"`
	TaskID    *int     `json:"task_id,omitempty"`
	SpentDate *string  `json:"spent_date,omitempty"`
	Hours     *float64 `json:"hours,omitempty"`
	Notes     *string  `json:"notes,omitempty"`
}
```

Pointer types so omitempty works correctly (only sends changed fields).

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/balast/CodingProjects/harvest-go-cli && go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
cd /home/balast/CodingProjects/harvest-go-cli
git add internal/models/time_entry.go
git commit -m "feat: add UpdateTimeEntryRequest model"
```

---

### Task 3: Add `UpdateTimeEntry` and `DeleteTimeEntry` API methods

**Files:**
- Modify: `internal/api/time_entries.go`

- [ ] **Step 1: Add UpdateTimeEntry method**

Append to `internal/api/time_entries.go`:

```go
func (c *Client) UpdateTimeEntry(entryID int, req *models.UpdateTimeEntryRequest) (*models.TimeEntry, error) {
	path := fmt.Sprintf("/time_entries/%d", entryID)
	body, err := c.doRequest("PATCH", path, req)
	if err != nil {
		return nil, err
	}

	var entry models.TimeEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &entry, nil
}
```

- [ ] **Step 2: Add DeleteTimeEntry method**

Append to `internal/api/time_entries.go`:

```go
func (c *Client) DeleteTimeEntry(entryID int) error {
	path := fmt.Sprintf("/time_entries/%d", entryID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /home/balast/CodingProjects/harvest-go-cli && go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
cd /home/balast/CodingProjects/harvest-go-cli
git add internal/api/time_entries.go
git commit -m "feat: add UpdateTimeEntry and DeleteTimeEntry API methods"
```

---

### Task 4: Add `SelectTimeEntry` prompt helper

**Files:**
- Modify: `internal/prompt/prompt.go`

Both `edit` and `delete` need to let the user pick a time entry from a list. Add a shared prompt function.

- [ ] **Step 1: Add SelectTimeEntry function**

Add to `internal/prompt/prompt.go`:

```go
func SelectTimeEntry(entries []models.TimeEntry) (*models.TimeEntry, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("no time entries to select from")
	}

	items := make([]selectItem, len(entries))
	for i, e := range entries {
		hoursStr := fmt.Sprintf("%.2f hrs", e.Hours)
		if e.IsRunning {
			hoursStr += " (running)"
		}
		title := fmt.Sprintf("%s  %s / %s", hoursStr, e.Project.Name, e.Task.Name)
		desc := ""
		if e.Notes != "" {
			desc = e.Notes
		}
		if e.IsLocked {
			desc = "[locked] " + desc
		}
		items[i] = selectItem{
			title: title,
			desc:  desc,
			index: i,
		}
	}

	idx, err := runSelect("Select time entry", items)
	if err != nil {
		return nil, err
	}

	return &entries[idx], nil
}
```

- [ ] **Step 2: Add SelectEditField function**

This lets the user pick what to edit on the selected entry. Add to `internal/prompt/prompt.go`:

```go
func SelectEditField() (string, error) {
	fields := []selectItem{
		{title: "Hours", index: 0},
		{title: "Notes", index: 1},
		{title: "Project & Task", index: 2},
		{title: "Date", index: 3},
		{title: "Done (save & exit)", index: 4},
	}

	idx, err := runSelect("What to edit?", fields)
	if err != nil {
		return "", err
	}

	switch idx {
	case 0:
		return "hours", nil
	case 1:
		return "notes", nil
	case 2:
		return "project", nil
	case 3:
		return "date", nil
	default:
		return "done", nil
	}
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /home/balast/CodingProjects/harvest-go-cli && go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
cd /home/balast/CodingProjects/harvest-go-cli
git add internal/prompt/prompt.go
git commit -m "feat: add SelectTimeEntry and SelectEditField prompt helpers"
```

---

### Task 5: Add `harvest edit` command

**Files:**
- Create: `cmd/edit.go`
- Modify: `cmd/root.go:65-71` (register command)

- [ ] **Step 1: Create cmd/edit.go**

```go
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
```

- [ ] **Step 2: Register the command in root.go**

In `cmd/root.go`, add `editCmd` to the init function:

```go
func init() {
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(editCmd)
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /home/balast/CodingProjects/harvest-go-cli && go build ./...`
Expected: No errors

- [ ] **Step 4: Manual test**

Run: `cd /home/balast/CodingProjects/harvest-go-cli && go run . edit --help`
Expected: Shows edit command help with `--date` flag

- [ ] **Step 5: Commit**

```bash
cd /home/balast/CodingProjects/harvest-go-cli
git add cmd/edit.go cmd/root.go
git commit -m "feat: add harvest edit command"
```

---

### Task 6: Add `harvest delete` command

**Files:**
- Create: `cmd/delete.go`
- Modify: `cmd/root.go:65-71` (register command)

- [ ] **Step 1: Create cmd/delete.go**

```go
package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"harvest-cli/internal/prompt"
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
```

- [ ] **Step 2: Register the command in root.go**

In `cmd/root.go`, add `deleteCmd` to the init function:

```go
func init() {
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(deleteCmd)
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /home/balast/CodingProjects/harvest-go-cli && go build ./...`
Expected: No errors

- [ ] **Step 4: Manual test**

Run: `cd /home/balast/CodingProjects/harvest-go-cli && go run . delete --help`
Expected: Shows delete command help with `--date` flag

- [ ] **Step 5: Commit**

```bash
cd /home/balast/CodingProjects/harvest-go-cli
git add cmd/delete.go cmd/root.go
git commit -m "feat: add harvest delete command"
```

---

### Task 7: Add `--date` flag to `harvest log`

**Files:**
- Modify: `cmd/log.go`

Currently `log` always prompts for date interactively. Add a `--date` flag so it can be passed directly, especially useful for non-interactive usage like `harvest log "proj" "task" 2.5 "notes" --date 2026-03-27`.

- [ ] **Step 1: Add the flag and modify runLog**

In `cmd/log.go`, add a package-level var and register the flag:

```go
var logDate string
```

In the existing `logCmd` definition area (or a new `init` in log.go), add:

```go
func init() {
	logCmd.Flags().StringVar(&logDate, "date", "", "Date for the entry (YYYY-MM-DD), defaults to interactive prompt")
}
```

Then in `runLog`, replace the date input section (lines 182-186):

```go
	// from:
	// Input date (always interactive for now)
	// date, err := prompt.InputDate()
	// if err != nil {
	// 	return fmt.Errorf("date input cancelled: %w", err)
	// }

	// to:
	var date string
	if logDate != "" {
		if _, err := time.Parse("2006-01-02", logDate); err != nil {
			return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		date = logDate
		fmt.Printf("Date: %s\n", date)
	} else {
		d, err := prompt.InputDate()
		if err != nil {
			return fmt.Errorf("date input cancelled: %w", err)
		}
		date = d
	}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/balast/CodingProjects/harvest-go-cli && go build ./...`
Expected: No errors

- [ ] **Step 3: Manual test**

Run: `cd /home/balast/CodingProjects/harvest-go-cli && go run . log --help`
Expected: Shows `--date` flag in the help output

- [ ] **Step 4: Commit**

```bash
cd /home/balast/CodingProjects/harvest-go-cli
git add cmd/log.go
git commit -m "feat: add --date flag to harvest log command"
```

---

### Task 8: Build, install, and end-to-end test

**Files:** None (testing only)

- [ ] **Step 1: Build the binary**

Run: `cd /home/balast/CodingProjects/harvest-go-cli && go build -o harvest .`
Expected: Binary compiles without errors

- [ ] **Step 2: Install to system path**

Run: `cd /home/balast/CodingProjects/harvest-go-cli && go install .`
Expected: `harvest` binary updated in `$GOPATH/bin` or `$GOBIN`

- [ ] **Step 3: Verify all commands show in help**

Run: `harvest --help`
Expected: Shows `edit` and `delete` in the Available Commands list

- [ ] **Step 4: Test edit help**

Run: `harvest edit --help`
Expected: Shows `--date` flag documentation

- [ ] **Step 5: Test delete help**

Run: `harvest delete --help`
Expected: Shows `--date` flag documentation

- [ ] **Step 6: Test log --date help**

Run: `harvest log --help`
Expected: Shows `--date` flag in the flags section

- [ ] **Step 7: Test view to find an unlocked entry for live testing**

Run: `harvest view today` or `harvest view --from 2026-03-30 --to 2026-04-02`
Expected: Shows current entries. Pick one for edit/delete testing if available.

- [ ] **Step 8: Push to fork**

```bash
cd /home/balast/CodingProjects/harvest-go-cli
git push fork main
```
