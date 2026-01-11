package timer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type TimerState struct {
	EntryID   int       `json:"entry_id"`
	ProjectID int       `json:"project_id"`
	TaskID    int       `json:"task_id"`
	StartedAt time.Time `json:"started_at"`
	Notes     string    `json:"notes"`
	Project   string    `json:"project"`
	Task      string    `json:"task"`
}

func getTimerFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".harvest-timer.json"), nil
}

func SaveTimerState(state *TimerState) error {
	path, err := getTimerFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal timer state: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write timer state: %w", err)
	}

	return nil
}

func LoadTimerState() (*TimerState, error) {
	path, err := getTimerFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No active timer
		}
		return nil, fmt.Errorf("failed to read timer state: %w", err)
	}

	var state TimerState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse timer state: %w", err)
	}

	return &state, nil
}

func ClearTimerState() error {
	path, err := getTimerFilePath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear timer state: %w", err)
	}

	return nil
}

func IsTimerRunning() bool {
	path, err := getTimerFilePath()
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}
