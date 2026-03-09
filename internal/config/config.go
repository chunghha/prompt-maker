package config

import (
	"errors"
	"fmt"
	"os"
)

//nolint:gosec // This is a false positive. We are defining the name of an env var, not a credential.
const apiKeyEnvVar = "GEMINI_API_KEY"

// ErrAPIKeyNotFound is returned when the API key environment variable is not set.
var ErrAPIKeyNotFound = errors.New("API key not found in environment variable")

// Config holds the application configuration loaded from the environment.
type Config struct {
	APIKey string
}

// Load reads configuration from environment variables and returns a Config.
func Load() (*Config, error) {
	apiKey := os.Getenv(apiKeyEnvVar)
	if apiKey == "" {
		return nil, fmt.Errorf("%w (checked environment variable: %s)", ErrAPIKeyNotFound, apiKeyEnvVar)
	}

	return &Config{APIKey: apiKey}, nil
}
