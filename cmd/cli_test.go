package cmd

import (
	"errors"
	"prompt-maker/internal/config"
	"prompt-maker/internal/gemini"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errTUI = errors.New("tui failed to start")

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()
	require.NotNil(t, cmd)
	assert.Equal(t, "prompt-maker", cmd.Use)
	assert.Equal(t, "Crafts optimized prompts for AI models.", cmd.Short)
}

func TestApp_RunTUI(t *testing.T) { // <-- Renamed test function for clarity
	t.Run("ConfigLoadError", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "")

		a := &app{}
		err := a.runTUI() // <-- Use runTUI() instead of run()
		require.Error(t, err)
		assert.ErrorIs(t, err, config.ErrAPIKeyNotFound)
	})

	t.Run("ModelSelectionError", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "test-key")

		a := &app{
			selectModel: func() (string, error) {
				return "", gemini.ErrModelSelectionCanceled
			},
		}
		err := a.runTUI() // <-- Use runTUI() instead of run()
		require.Error(t, err)
		assert.ErrorIs(t, err, gemini.ErrModelSelectionCanceled)
	})

	t.Run("TUIError", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "test-key")

		a := &app{
			selectModel: func() (string, error) {
				return "test-model", nil
			},
			startTUI: func(cfg *config.Config, modelName, version string) error {
				assert.NotNil(t, cfg)
				assert.Equal(t, "test-model", modelName)
				assert.Equal(t, "dev", version)
				return errTUI
			},
			version: "dev",
		}
		err := a.runTUI() // <-- Use runTUI() instead of run()
		require.Error(t, err)
		assert.ErrorIs(t, err, errTUI)
	})
}
