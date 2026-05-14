package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Adam-D-Lewis/harvest-go-cli/config"
)

const baseURL = "https://api.harvestapp.com/v2"

type Client struct {
	httpClient *http.Client
	token      string
	accountID  string
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{},
		token:      cfg.HarvestToken,
		accountID:  cfg.HarvestAccountID,
	}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Harvest-Account-Id", c.accountID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Harvest CLI")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    getErrorMessage(resp.StatusCode, respBody),
		}
	}

	return respBody, nil
}

func getErrorMessage(statusCode int, body []byte) string {
	switch statusCode {
	case 401:
		return "Invalid credentials. Check HARVEST_TOKEN in .env"
	case 403:
		return "Access denied. Check HARVEST_ACCOUNT_ID in .env"
	case 404:
		return "Resource not found"
	case 429:
		return "Rate limited. Please wait and try again"
	default:
		if len(body) > 0 {
			var errResp map[string]interface{}
			if json.Unmarshal(body, &errResp) == nil {
				if msg, ok := errResp["message"].(string); ok {
					return msg
				}
			}
			return string(body)
		}
		return fmt.Sprintf("HTTP %d error", statusCode)
	}
}
