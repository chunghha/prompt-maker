package cmd

import (
	"context"
	"fmt"
	"prompt-maker/internal/config"
	"prompt-maker/internal/tui"
	"prompt-maker/internal/web"

	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

var version = "dev"

type startTUIFn func(cfg *config.Config, version, modelName, history string, temperature float32) error

type app struct {
	startTUI    startTUIFn
	version     string
	model       string
	history     string
	temperature float32
}

func NewRootCmd() *cobra.Command {
	a := &app{
		startTUI: tui.Start,
		version:  version,
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
	cmd.Flags().StringVar(&a.model, "model", "", "Specify the model to use")
	cmd.Flags().Float32Var(&a.temperature, "temperature", 0.0, "Specify the model temperature")
	cmd.Flags().StringVar(&a.history, "history", "", "Path to a file containing chat history")

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

// runTUI is simplified. It no longer selects a model.
func (a *app) runTUI() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return a.startTUI(cfg, a.version, a.model, a.history, a.temperature)
}
