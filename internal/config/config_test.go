package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name         string
		setupEnv     func()
		expectedConf *Config
		expectedErr  error
	}{
		{
			name: "successful loading",
			setupEnv: func() {
				if err := os.Setenv(apiKeyEnvVar, "test_api_key"); err != nil {
					t.Fatalf("failed to set env: %v", err)
				}
			},
			expectedConf: &Config{APIKey: "test_api_key"},
			expectedErr:  nil,
		},
		{
			name: "API key not found",
			setupEnv: func() {
				if err := os.Unsetenv(apiKeyEnvVar); err != nil {
					t.Fatalf("failed to unset env: %v", err)
				}
			},
			expectedConf: nil,
			expectedErr:  ErrAPIKeyNotFound,
		},
		{
			name: "empty API key",
			setupEnv: func() {
				if err := os.Setenv(apiKeyEnvVar, ""); err != nil {
					t.Fatalf("failed to set env: %v", err)
				}
			},
			expectedConf: nil,
			expectedErr:  ErrAPIKeyNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variables after each test
			defer func() {
				if err := os.Unsetenv(apiKeyEnvVar); err != nil {
					t.Fatalf("failed to unset env in defer: %v", err)
				}
			}()

			tt.setupEnv()

			conf, err := Load()

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				assert.Contains(t, err.Error(), apiKeyEnvVar)
				assert.Nil(t, conf)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedConf, conf)
			}
		})
	}
}
