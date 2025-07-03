package cmd

import (
	"fmt"
	"prompt-maker/internal/config"
	"prompt-maker/internal/gemini"
	"prompt-maker/internal/tui"

	"github.com/spf13/cobra"
)

// selectModelFn and startTUIFn define function types for our dependencies.
type selectModelFn func() (string, error)
type startTUIFn func(*config.Config, string) error

// app holds the dependencies and logic of the application.
// This avoids using global variables and makes testing cleaner.
type app struct {
	selectModel selectModelFn
	startTUI    startTUIFn
}

// NewRootCmd creates and configures the main command for the application.
func NewRootCmd() *cobra.Command {
	// Instantiate the app with real dependencies.
	a := &app{
		selectModel: gemini.SelectModel,
		startTUI:    tui.Start,
	}

	cmd := &cobra.Command{
		Use:   "prompt-maker",
		Short: "Crafts optimized prompts for AI models.",
		RunE: func(_ *cobra.Command, _ []string) error {
			// The RunE function now calls the run method on our app instance.
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

	if err := a.startTUI(cfg, selectedModel); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	return nil
}
