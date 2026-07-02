package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"prompt-maker/internal/config"
	"prompt-maker/internal/observability"
	"prompt-maker/internal/tui"
	"prompt-maker/internal/web"

	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

var version = "dev"

const tracerShutdownTimeout = 5 * time.Second

type startTUIFn func(cfg *config.Config, version, modelName, history string, temperature float32) error

type app struct {
	startTUI    startTUIFn
	version     string
	model       string
	history     string
	temperature float32
}

// NewRootCmd creates the root Cobra command for the prompt-maker CLI.
// It supports TUI mode (default) and web server mode (--web).
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

// runWeb configures and starts the web server with OTEL tracing.
// In Echo v5, Start blocks until an OS signal is received and
// performs graceful shutdown automatically.
func (a *app) runWeb() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()

	shutdown, err := observability.SetupTracing(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup tracing: %w", err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), tracerShutdownTimeout)
		defer cancel()

		if shutdownErr := shutdown(shutdownCtx); shutdownErr != nil {
			slog.ErrorContext(shutdownCtx, "failed to shutdown tracer provider", "error", shutdownErr)
		}
	}()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.APIKey})
	if err != nil {
		return fmt.Errorf("failed to create genai client: %w", err)
	}

	promptGenerator := web.NewGeminiPromptGenerator(client)

	webCfg := web.Config{
		Generator: promptGenerator,
		Version:   a.version,
	}

	server, err := web.NewServer(webCfg)
	if err != nil {
		return fmt.Errorf("failed to create web server: %w", err)
	}

	slog.InfoContext(ctx, "starting web server", "addr", "http://localhost:8080")

	// Echo v5's Start blocks until Interrupt/SIGTERM and handles graceful shutdown.
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
