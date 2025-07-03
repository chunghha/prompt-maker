package tui

import (
	"context"
	"fmt"
	"prompt-maker/internal/config"
	"prompt-maker/internal/gemini"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"google.golang.org/genai"
)

const (
	viewportWidth      = 80
	viewportHeight     = 20
	modelTemperature   = 0.1
	textInputCharLimit = 1000
	textInputWidth     = 80
)

type viewState int

const (
	viewReady viewState = iota
	viewBusy
	viewResult
	viewError
)

type model struct {
	state         viewState
	textInput     textinput.Model
	spinner       spinner.Model
	viewport      viewport.Model
	geminiSession gemini.ChatSession
	quitting      bool
}

func New(session gemini.ChatSession) tea.Model {
	ti := textinput.New()
	ti.Placeholder = "Enter your rough prompt here..."
	ti.Focus()
	ti.CharLimit = textInputCharLimit
	ti.Width = textInputWidth

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(viewportWidth, viewportHeight)

	return &model{
		state:         viewReady,
		textInput:     ti,
		spinner:       s,
		viewport:      vp,
		geminiSession: session,
	}
}

func (*model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		//nolint:exhaustive // We only care about a few key presses.
		switch keyMsg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd

	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m *model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	return fmt.Sprintf(
		"Enter your prompt and press Enter.\n\n%s\n\n(esc to quit)",
		m.textInput.View(),
	)
}

func Start(cfg *config.Config, modelName string) error {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.APIKey, Backend: genai.BackendGeminiAPI})
	if err != nil {
		return fmt.Errorf("failed to create generative AI client: %w", err)
	}

	genConfig := &genai.GenerateContentConfig{Temperature: genai.Ptr(float32(modelTemperature))}

	chatSession, err := client.Chats.Create(ctx, modelName, genConfig, nil)
	if err != nil {
		return fmt.Errorf("failed to create chat session: %w", err)
	}

	m := New(chatSession)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI program: %w", err)
	}

	return nil
}
