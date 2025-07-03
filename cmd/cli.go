package cmd

import (
	"fmt"
	"prompt-maker/internal/config"
	"prompt-maker/internal/gemini"
	"prompt-maker/internal/tui"

	"github.com/spf13/cobra"
)

// This variable will be set by the linker during the build process.
var version = "dev" // Default value for local `go run`

// selectModelFn and startTUIFn define function types for our dependencies.
type selectModelFn func() (string, error)

// FIX: The TUI now needs the version string.
type startTUIFn func(*config.Config, string, string) error

// app holds the dependencies and logic of the application.
type app struct {
	selectModel selectModelFn
	startTUI    startTUIFn
	version     string
}

// NewRootCmd creates and configures the main command for the application.
func NewRootCmd() *cobra.Command {
	// Instantiate the app with real dependencies.
	a := &app{
		selectModel: gemini.SelectModel,
		startTUI:    tui.Start,
		version:     version, // Use the version variable.
	}

	cmd := &cobra.Command{
		Use:   "prompt-maker",
		Short: "Crafts optimized prompts for AI models.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return a.run()
		},
	}

	return cmd
}

// run contains the core application logic, making it testable and separate from cobra.
func (a *app) run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	selectedModel, err := a.selectModel()
	if err != nil {
		return fmt.Errorf("failed to select model: %w", err)
	}

	fmt.Printf("Using model: %s\n", selectedModel)

	// FIX: Pass the version to the TUI starter.
	if err := a.startTUI(cfg, selectedModel, a.version); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	return nil
}
