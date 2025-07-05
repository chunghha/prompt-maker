package cmd

import (
	"errors"
	"prompt-maker/internal/config"
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

func TestApp_RunTUI(t *testing.T) {
	t.Run("ConfigLoadError", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "")

		a := &app{}
		err := a.runTUI()
		require.Error(t, err)
		assert.ErrorIs(t, err, config.ErrAPIKeyNotFound)
	})

	t.Run("TUIError", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "test-key")

		a := &app{
			startTUI: func(cfg *config.Config, version string) error {
				assert.NotNil(t, cfg)
				assert.Equal(t, "dev", version)
				return errTUI
			},
			version: "dev",
		}
		err := a.runTUI()
		require.Error(t, err)
		assert.ErrorIs(t, err, errTUI)
	})
}
