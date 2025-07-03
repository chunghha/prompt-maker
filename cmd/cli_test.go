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

func TestApp_Run(t *testing.T) {
	t.Run("ConfigLoadError", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "")

		a := &app{}
		err := a.run()
		require.Error(t, err)
		assert.ErrorIs(t, err, config.ErrAPIKeyNotFound)
	})

	t.Run("ModelSelectionError", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "test-key")

		a := &app{
			selectModel: func() (string, error) {
				// Use the static error from the gemini package.
				return "", gemini.ErrModelSelectionCanceled
			},
		}
		err := a.run()
		require.Error(t, err)
		assert.ErrorIs(t, err, gemini.ErrModelSelectionCanceled)
	})

	t.Run("TUIError", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "test-key")

		a := &app{
			selectModel: func() (string, error) {
				return "test-model", nil
			},
			startTUI: func(cfg *config.Config, modelName string) error {
				assert.NotNil(t, cfg)
				assert.Equal(t, "test-model", modelName)
				return errTUI
			},
		}
		err := a.run()
		require.Error(t, err)
		assert.ErrorIs(t, err, errTUI)
	})
}
