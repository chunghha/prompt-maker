package tui

import (
	"context"
	"errors"
	"fmt"

	"prompt-maker/internal/config"
	"prompt-maker/internal/gemini"
	"prompt-maker/internal/prompt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"google.golang.org/genai"
)

// --- Constants ---

const (
	appName                  = "Prompt Maker"
	modelTemperature         = 0.1
	textInputCharLimit       = 2000
	horizontalPadding        = 2
	initialViewportWidth     = 100
	initialViewportHeight    = 20
	placeholderRoughPrompt   = "Enter your rough prompt here..."
	placeholderCraftedPrompt = "Review the crafted prompt, or press Enter to submit."
	placeholderNewPrompt     = "Press Enter to start a new prompt."
	helpText                 = "esc: quit"
	thinkingText             = "Crafting prompt..."
	initialInstructionText   = "Enter a rough prompt for Lyra to improve."
	errorText                = "Error: "
	goodbyeText              = "Goodbye!\n"
)

var (
	errPromptEmpty          = errors.New("prompt cannot be empty")
	errNoResponseFromGemini = errors.New("no response content from Gemini")
)

// --- TUI Messages ---

type aiResponseMsg struct{ response string }
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// --- TUI Model ---

type viewState int

const (
	viewReady viewState = iota
	viewBusy
	viewResult
	viewError
)

type model struct {
	ctx             context.Context
	state           viewState
	textInput       textinput.Model
	spinner         spinner.Model
	viewport        viewport.Model
	geminiSession   gemini.ChatSession
	appVersion      string
	selectedModel   string
	quitting        bool
	isPromptCrafted bool
	errorMessage    string
	width           int
	height          int
	styles          Styles
}

type Styles struct {
	header, appName, appVersion, modelName, mainContent, input, statusBar lipgloss.Style
}

func newStyles() Styles {
	return Styles{
		header:      lipgloss.NewStyle().Padding(0, 1),
		appName:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("35")),
		appVersion:  lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
		modelName:   lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
		mainContent: lipgloss.NewStyle().Padding(0, horizontalPadding),
		input:       lipgloss.NewStyle().Padding(0, horizontalPadding),
		statusBar:   lipgloss.NewStyle().Padding(0, horizontalPadding),
	}
}

func New(ctx context.Context, session gemini.ChatSession, modelName, version string) tea.Model {
	ti := textinput.New()
	ti.Placeholder = placeholderRoughPrompt
	ti.Focus()
	ti.CharLimit = textInputCharLimit

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(initialViewportWidth, initialViewportHeight)

	return &model{
		ctx:           ctx,
		state:         viewReady,
		textInput:     ti,
		spinner:       s,
		viewport:      vp,
		geminiSession: session,
		appVersion:    version,
		selectedModel: modelName,
		styles:        newStyles(),
	}
}

func (*model) Init() tea.Cmd {
	return textinput.Blink
}

// --- Update Logic ---

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case aiResponseMsg:
		return m.handleAIResponse(msg)
	case errMsg:
		return m.handleError(msg)
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	return m.updateComponents(msg)
}

func (m *model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	headerHeight := lipgloss.Height(m.headerView())
	inputHeight := lipgloss.Height(m.inputView())
	statusHeight := lipgloss.Height(m.statusBarView())
	m.viewport.Width = m.width
	m.viewport.Height = m.height - headerHeight - inputHeight - statusHeight
	m.textInput.Width = m.width - (horizontalPadding * 2)

	return m, nil
}

func (m *model) handleAIResponse(msg aiResponseMsg) (tea.Model, tea.Cmd) {
	if !m.isPromptCrafted {
		m.isPromptCrafted = true
		m.textInput.SetValue(msg.response)
		m.textInput.Placeholder = placeholderCraftedPrompt
		m.viewport.SetContent(msg.response)
		m.state = viewReady
	} else {
		m.isPromptCrafted = false
		m.viewport.SetContent(msg.response)
		m.textInput.Reset()
		m.textInput.Placeholder = placeholderNewPrompt
		m.state = viewResult
	}

	m.viewport.GotoTop()

	return m, nil
}

func (m *model) handleError(msg errMsg) (tea.Model, tea.Cmd) {
	m.state = viewError
	m.errorMessage = msg.err.Error()
	m.viewport.SetContent(errorText + m.errorMessage)

	return m, nil
}

func (m *model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		switch m.state {
		case viewReady:
			m.state = viewBusy
			userInput := m.textInput.Value()

			return m, tea.Batch(m.spinner.Tick, sendPromptCmd(m, userInput, !m.isPromptCrafted))
		case viewResult, viewError:
			m.state = viewReady
			m.isPromptCrafted = false
			m.textInput.Reset()
			m.textInput.Placeholder = placeholderRoughPrompt
			m.viewport.SetContent("")

			return m, nil
		case viewBusy:
			// Do nothing while busy.
		}
	}

	if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
		m.quitting = true
		return m, tea.Quit
	}

	return m.updateComponents(msg)
}

func (m *model) updateComponents(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	if m.state == viewBusy {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// --- View Logic ---

func (m *model) View() string {
	if m.quitting {
		return goodbyeText
	}

	if m.width == 0 {
		return "Initializing..."
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.headerView(),
		m.styles.mainContent.Render(m.mainContentView()),
		m.inputView(),
		m.statusBarView(),
	)
}

func (m *model) headerView() string {
	left := m.styles.appName.Render(appName) + " " + m.styles.appVersion.Render("("+m.appVersion+")")
	right := m.styles.modelName.Render("Model: " + m.selectedModel)
	space := lipgloss.NewStyle().Width(m.width - lipgloss.Width(left) - lipgloss.Width(right)).Render("")

	return m.styles.header.Render(lipgloss.JoinHorizontal(lipgloss.Bottom, left, space, right))
}

func (m *model) mainContentView() string {
	switch m.state {
	case viewBusy:
		return m.spinner.View() + thinkingText
	case viewReady:
		if m.viewport.View() != "" {
			return m.viewport.View()
		}

		return initialInstructionText
	case viewResult:
		return m.viewport.View()
	case viewError:
		return m.viewport.View()
	}

	return "" // Should be unreachable
}

func (m *model) inputView() string {
	if m.state == viewResult {
		return ""
	}

	return m.styles.input.Render(m.textInput.View())
}

func (m *model) statusBarView() string {
	return m.styles.statusBar.Render(helpText)
}

// --- Command Logic ---

func sendPromptCmd(m *model, userPrompt string, useLyra bool) tea.Cmd {
	return func() tea.Msg {
		if userPrompt == "" {
			return errMsg{err: errPromptEmpty}
		}

		if useLyra {
			return generateCraftedPrompt(m, userPrompt)
		}

		return getFinalAnswer(m, userPrompt)
	}
}

func generateCraftedPrompt(m *model, userPrompt string) tea.Msg {
	response, err := prompt.Generate(m.ctx, m.geminiSession, userPrompt)
	if err != nil {
		return errMsg{err: err}
	}

	return aiResponseMsg{response: response}
}

func getFinalAnswer(m *model, userPrompt string) tea.Msg {
	resp, err := m.geminiSession.SendMessage(m.ctx, genai.Part{Text: userPrompt})
	if err != nil {
		return errMsg{err: err}
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return errMsg{err: errNoResponseFromGemini}
	}

	return aiResponseMsg{response: resp.Text()}
}

// --- TUI Starter ---

func Start(cfg *config.Config, modelName, version string) error {
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

	p := tea.NewProgram(New(ctx, chatSession, modelName, version), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI program: %w", err)
	}

	return nil
}
