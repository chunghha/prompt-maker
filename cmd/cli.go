package cmd

import (
	"context"
	"fmt"
	"prompt-maker/internal/config"
	"prompt-maker/internal/gemini"
	"prompt-maker/internal/tui"
	"prompt-maker/internal/web"

	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

var version = "dev"

type selectModelFn func() (string, error)
type startTUIFn func(*config.Config, string, string) error

type app struct {
	selectModel selectModelFn
	startTUI    startTUIFn
	version     string
}

func NewRootCmd() *cobra.Command {
	a := &app{
		selectModel: gemini.SelectModel,
		startTUI:    tui.Start,
		version:     version,
	}

	var webMode bool

	cmd := &cobra.Command{
		Use:   "prompt-maker",
		Short: "Crafts optimized prompts for AI models.",
		RunE: func(_ *cobra.Command, _ []string) error {
			if webMode {
				return a.runWeb()
			}
			return a.runTUI()
		},
	}

	cmd.Flags().BoolVar(&webMode, "web", false, "Run in web server mode on port 8080")

	return cmd
}

// runWeb configures and starts the web server.
func (a *app) runWeb() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.APIKey})
	if err != nil {
		return fmt.Errorf("failed to create genai client: %w", err)
	}

	// The generator constructor no longer takes a model name.
	promptGenerator := web.NewGeminiPromptGenerator(client)

	// The web config no longer takes a model name.
	webCfg := web.Config{
		Generator: promptGenerator,
		Version:   a.version,
	}

	server, err := web.NewServer(webCfg)
	if err != nil {
		return fmt.Errorf("failed to create web server: %w", err)
	}

	fmt.Printf("Starting web server on http://localhost:8080\n")

	return server.Start(":8080")
}

// runTUI contains the original logic for the terminal UI.
func (a *app) runTUI() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	selectedModel, err := a.selectModel()
	if err != nil {
		return fmt.Errorf("failed to select model: %w", err)
	}

	fmt.Printf("Using model: %s\n", selectedModel)

	return a.startTUI(cfg, selectedModel, a.version)
}
