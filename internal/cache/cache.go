package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Adam-D-Lewis/harvest-go-cli/internal/models"
)

const cacheMaxAge = 1 * time.Hour

type ProjectCache struct {
	Projects  []models.ProjectAssignment `json:"projects"`
	UpdatedAt time.Time                  `json:"updated_at"`
}

func getCacheFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	cacheDir := filepath.Join(home, ".cache", "harvest")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	return filepath.Join(cacheDir, "projects.json"), nil
}

func SaveProjects(projects []models.ProjectAssignment) error {
	path, err := getCacheFilePath()
	if err != nil {
		return err
	}

	cache := ProjectCache{
		Projects:  projects,
		UpdatedAt: time.Now(),
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

func LoadProjects() ([]models.ProjectAssignment, bool) {
	path, err := getCacheFilePath()
	if err != nil {
		return nil, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var cache ProjectCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, false
	}

	// Check if cache is expired
	if time.Since(cache.UpdatedAt) > cacheMaxAge {
		return cache.Projects, false // Return stale data but indicate refresh needed
	}

	return cache.Projects, true
}

func ClearCache() error {
	path, err := getCacheFilePath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}
