package tui

import (
	"context"
	"fmt"

	"prompt-maker/internal/config"
	"prompt-maker/internal/gemini"
	"prompt-maker/internal/prompt"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"google.golang.org/genai"
)

func copyToClipboardCmd(content string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(content)
		if err != nil {
			return errMsg{err: fmt.Errorf("%w: %w", errClipboardWrite, err)}
		}

		return statusMessage("Copied!")
	}
}

// sendPromptCmd creates a tea.Cmd that sends a prompt to the AI model.
// It captures ctx, chatSvc, and selectedModel by value to avoid a data
// race with the main Update goroutine.
func sendPromptCmd(ctx context.Context, chatSvc chatCreator, selectedModel, userPrompt string, useLyra bool) tea.Cmd {
	return func() tea.Msg {
		if userPrompt == "" {
			return errMsg{err: errPromptEmpty}
		}

		genConfig := &genai.GenerateContentConfig{Temperature: genai.Ptr(float32(config.DefaultModelTemperature))}

		session, err := chatSvc.Create(ctx, selectedModel, genConfig, nil)
		if err != nil {
			return errMsg{err: fmt.Errorf("creating chat session: %w", err)}
		}

		if useLyra {
			return generateCraftedPrompt(ctx, session, userPrompt)
		}

		return getFinalAnswer(ctx, session, userPrompt)
	}
}

func generateCraftedPrompt(ctx context.Context, session gemini.ChatSession, userPrompt string) tea.Msg {
	response, err := prompt.Generate(ctx, session, userPrompt)
	if err != nil {
		return errMsg{err: err}
	}

	return aiResponseMsg{response: response}
}

func getFinalAnswer(ctx context.Context, session gemini.ChatSession, userPrompt string) tea.Msg {
	response, err := prompt.Execute(ctx, session, userPrompt)
	if err != nil {
		return errMsg{err: err}
	}

	return aiResponseMsg{response: response}
}
