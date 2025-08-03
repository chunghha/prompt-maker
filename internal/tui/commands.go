package tui

import (
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

func sendPromptCmd(m *model, userPrompt string, useLyra bool) tea.Cmd {
	return func() tea.Msg {
		if userPrompt == "" {
			return errMsg{err: errPromptEmpty}
		}

		genConfig := &genai.GenerateContentConfig{Temperature: genai.Ptr(float32(config.DefaultModelTemperature))}

		session, err := m.chatSvc.Create(m.ctx, m.selectedModel, genConfig, nil)
		if err != nil {
			return errMsg{err: err}
		}

		if useLyra {
			return generateCraftedPrompt(m, session, userPrompt)
		}

		return getFinalAnswer(m, session, userPrompt)
	}
}

func generateCraftedPrompt(m *model, session gemini.ChatSession, userPrompt string) tea.Msg {
	response, err := prompt.Generate(m.ctx, session, userPrompt)
	if err != nil {
		return errMsg{err: err}
	}

	return aiResponseMsg{response: response}
}

func getFinalAnswer(m *model, session gemini.ChatSession, userPrompt string) tea.Msg {
	response, err := prompt.Execute(m.ctx, session, userPrompt)
	if err != nil {
		return errMsg{err: err}
	}

	return aiResponseMsg{response: response}
}
