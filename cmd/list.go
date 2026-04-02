package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Adam-D-Lewis/harvest-go-cli/internal/cache"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List projects and tasks",
	Long:    `List your assigned projects and their tasks.`,
}

var listProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List all assigned projects",
	RunE:  runListProjects,
}

var listTasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List tasks for a project",
	Long:  `List all tasks for a specific project. Use --project flag to specify project ID.`,
	RunE:  runListTasks,
}

var projectID int

func init() {
	listCmd.AddCommand(listProjectsCmd)
	listCmd.AddCommand(listTasksCmd)
	listTasksCmd.Flags().IntVarP(&projectID, "project", "p", 0, "Project ID")
}

func runListProjects(cmd *cobra.Command, args []string) error {
	fmt.Println("Fetching projects...")
	projects, err := apiClient.GetProjectAssignments()
	if err != nil {
		return fmt.Errorf("failed to fetch projects: %w", err)
	}

	// Cache projects for shell completion
	_ = cache.SaveProjects(projects)

	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	fmt.Println()
	fmt.Printf("%-10s %-40s %s\n", "ID", "PROJECT", "CLIENT")
	fmt.Println("─────────────────────────────────────────────────────────────────────────")

	for _, p := range projects {
		fmt.Printf("%-10d %-40s %s\n", p.Project.ID, truncate(p.Project.Name, 40), p.Client.Name)
	}

	fmt.Printf("\nTotal: %d projects\n", len(projects))

	return nil
}

func runListTasks(cmd *cobra.Command, args []string) error {
	if projectID == 0 {
		return fmt.Errorf("project ID is required. Use --project or -p flag")
	}

	fmt.Println("Fetching projects...")
	projects, err := apiClient.GetProjectAssignments()
	if err != nil {
		return fmt.Errorf("failed to fetch projects: %w", err)
	}

	var projectName string
	var tasks []struct {
		id       int
		name     string
		billable bool
	}

	for _, p := range projects {
		if p.Project.ID == projectID {
			projectName = p.Project.Name
			for _, t := range p.TaskAssignments {
				if t.IsActive {
					tasks = append(tasks, struct {
						id       int
						name     string
						billable bool
					}{
						id:       t.Task.ID,
						name:     t.Task.Name,
						billable: t.Billable,
					})
				}
			}
			break
		}
	}

	if projectName == "" {
		return fmt.Errorf("project with ID %d not found", projectID)
	}

	if len(tasks) == 0 {
		fmt.Printf("No active tasks found for project: %s\n", projectName)
		return nil
	}

	fmt.Printf("\nTasks for project: %s\n\n", projectName)
	fmt.Printf("%-10s %-40s %s\n", "ID", "TASK", "BILLABLE")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	for _, t := range tasks {
		billable := "No"
		if t.billable {
			billable = "Yes"
		}
		fmt.Printf("%-10d %-40s %s\n", t.id, truncate(t.name, 40), billable)
	}

	fmt.Printf("\nTotal: %d tasks\n", len(tasks))

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
