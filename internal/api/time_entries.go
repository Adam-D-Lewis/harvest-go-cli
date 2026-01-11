package api

import (
	"encoding/json"
	"fmt"

	"harvest-cli/internal/models"
)

func (c *Client) GetTimeEntries(from, to string) ([]models.TimeEntry, error) {
	var allEntries []models.TimeEntry
	page := 1

	for {
		path := fmt.Sprintf("/time_entries?from=%s&to=%s&page=%d&per_page=100", from, to, page)
		body, err := c.doRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}

		var resp models.TimeEntriesResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		allEntries = append(allEntries, resp.TimeEntries...)

		if page >= resp.TotalPages {
			break
		}
		page++
	}

	return allEntries, nil
}

func (c *Client) CreateTimeEntry(req *models.CreateTimeEntryRequest) (*models.TimeEntry, error) {
	body, err := c.doRequest("POST", "/time_entries", req)
	if err != nil {
		return nil, err
	}

	var entry models.TimeEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &entry, nil
}

func (c *Client) StopTimer(entryID int) (*models.TimeEntry, error) {
	path := fmt.Sprintf("/time_entries/%d/stop", entryID)
	body, err := c.doRequest("PATCH", path, nil)
	if err != nil {
		return nil, err
	}

	var entry models.TimeEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &entry, nil
}

func (c *Client) RestartTimer(entryID int) (*models.TimeEntry, error) {
	path := fmt.Sprintf("/time_entries/%d/restart", entryID)
	body, err := c.doRequest("PATCH", path, nil)
	if err != nil {
		return nil, err
	}

	var entry models.TimeEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &entry, nil
}
