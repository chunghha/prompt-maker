package main

import (
	"fmt"
	"log/slog"
	"os"

	"prompt-maker/cmd"
	"prompt-maker/internal/observability"
)

func main() {
	handler := observability.NewTraceHandler(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(slog.New(handler))

	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
