package api

import (
	"encoding/json"
	"fmt"

	"harvest-cli/internal/models"
)

func (c *Client) GetProjectAssignments() ([]models.ProjectAssignment, error) {
	var allAssignments []models.ProjectAssignment
	page := 1

	for {
		path := fmt.Sprintf("/users/me/project_assignments?page=%d&per_page=100", page)
		body, err := c.doRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}

		var resp models.ProjectAssignmentsResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Filter to only active assignments
		for _, pa := range resp.ProjectAssignments {
			if pa.IsActive {
				allAssignments = append(allAssignments, pa)
			}
		}

		if page >= resp.TotalPages {
			break
		}
		page++
	}

	return allAssignments, nil
}
