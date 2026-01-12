package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	HarvestToken     string
	HarvestAccountID string
}

func Load() (*Config, error) {
	// Try to load .env from current directory or parent directories
	if err := loadEnvFile(); err != nil {
		// .env file is optional if env vars are already set
	}

	token := os.Getenv("HARVEST_TOKEN")
	accountID := os.Getenv("HARVEST_ACCOUNT_ID")

	if token == "" {
		return nil, fmt.Errorf("HARVEST_TOKEN not set. Add it to .env file or set as environment variable")
	}

	if accountID == "" {
		return nil, fmt.Errorf("HARVEST_ACCOUNT_ID not set. Add it to .env file or set as environment variable")
	}

	return &Config{
		HarvestToken:     token,
		HarvestAccountID: accountID,
	}, nil
}

func loadEnvFile() error {
	// Try ~/.config/harvest/.env first (for tmux/global access)
	if home, err := os.UserHomeDir(); err == nil {
		configPath := filepath.Join(home, ".config", "harvest", ".env")
		if _, err := os.Stat(configPath); err == nil {
			return godotenv.Load(configPath)
		}
	}

	// Try current directory
	if err := godotenv.Load(); err == nil {
		return nil
	}

	// Try to find .env in parent directories (up to 5 levels)
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	for i := 0; i < 5; i++ {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return godotenv.Load(envPath)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return fmt.Errorf(".env file not found")
}
