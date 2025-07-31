package config

import (
	"errors"
	"fmt"
	"os"
)

//nolint:gosec // This is a false positive. We are defining the name of an env var, not a credential.
const (
	apiKeyEnvVar = "GEMINI_API_KEY"
	EMPTY_STRING = "" //nolint:revive // this is intentional for code readability
)

var (
	ErrAPIKeyNotFound = errors.New("API key not found in environment variable")
)

type Config struct {
	APIKey string
}

func Load() (*Config, error) {
	apiKey := os.Getenv(apiKeyEnvVar)
	if apiKey == EMPTY_STRING {
		return nil, fmt.Errorf("%w (checked environment variable: %s)", ErrAPIKeyNotFound, apiKeyEnvVar)
	}

	return &Config{APIKey: apiKey}, nil
}
