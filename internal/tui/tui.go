package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"prompt-maker/internal/config"
	"prompt-maker/internal/gemini"
	"prompt-maker/internal/prompt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"google.golang.org/genai"
)

// --- Constants ---

const (
	appName                   = "Prompt Maker"
	modelTemperature          = 0.1
	textInputCharLimit        = 2000
	horizontalPadding         = 2
	headerPadding             = 1
	initialViewportWidth      = 130
	initialViewportHeight     = 36
	copyStatusDuration        = time.Second * 2
	placeholderRoughPrompt    = "Enter your rough prompt here..."
	placeholderNewPrompt      = "Press Enter to start a new prompt."
	placeholderResubmit       = "Press 'r' to resubmit, or type a new prompt."
	thinkingTextCrafting      = "Crafting prompt..."
	thinkingTextGettingAnswer = "Getting a response..."
	initialInstructionText    = "Enter a rough prompt for Lyra to improve."
	errorText                 = "Error: "
	goodbyeText               = "Goodbye!\n"
)

var (
	errPromptEmpty    = errors.New("prompt cannot be empty")
	errClipboardWrite = errors.New("failed to write to clipboard")
)

// --- TUI Messages ---

type aiResponseMsg struct{ response string }
type errMsg struct{ err error }
type statusMessage string
type clearStatusMsg struct{}

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
	genaiClient     *genai.Client
	selectedModel   string
	appVersion      string
	quitting        bool
	isPromptCrafted bool
	craftedPrompt   string
	busyText        string
	errorMessage    string
	statusMessage   string
	viewportContent string // Store the full content of the viewport here
	width           int
	height          int
	styles          Styles
}

type Styles struct {
	header, appName, appVersion, modelName, mainContent, input, statusBar, statusText, resubmitHelp lipgloss.Style
}

func newStyles() Styles {
	return Styles{
		header:       lipgloss.NewStyle().Padding(0, headerPadding),
		appName:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("35")),
		appVersion:   lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
		modelName:    lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
		mainContent:  lipgloss.NewStyle().Padding(0, horizontalPadding),
		input:        lipgloss.NewStyle().Padding(1, horizontalPadding),
		statusBar:    lipgloss.NewStyle().Padding(0, horizontalPadding),
		statusText:   lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		resubmitHelp: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("35")),
	}
}

func New(ctx context.Context, client *genai.Client, modelName, version string) tea.Model {
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
		genaiClient:   client,
		selectedModel: modelName,
		appVersion:    version,
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
	case statusMessage:
		m.statusMessage = string(msg)

		return m, tea.Tick(copyStatusDuration, func(_ time.Time) tea.Msg {
			return clearStatusMsg{}
		})
	case clearStatusMsg:
		m.statusMessage = ""
		return m, nil
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	return m.updateComponents(msg)
}

func (m *model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	return m, nil
}

func (m *model) handleAIResponse(msg aiResponseMsg) (tea.Model, tea.Cmd) {
	if !m.isPromptCrafted {
		m.isPromptCrafted = true
		m.craftedPrompt = msg.response
		m.textInput.Reset()
		m.textInput.Placeholder = placeholderResubmit
		m.viewportContent = msg.response // Store full content for copying
		m.viewport.SetContent(msg.response)
		m.state = viewReady
	} else {
		m.isPromptCrafted = false
		m.craftedPrompt = ""
		m.viewportContent = msg.response // Store full content for copying
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
	m.viewportContent = errorText + m.errorMessage // Store full content for copying
	m.viewport.SetContent(m.viewportContent)

	return m, nil
}

//nolint:gocyclo // This function's complexity is managed by the state machine logic.
func (m *model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Copy the full viewportContent, not the rendered view.
	if (m.state == viewResult || (m.isPromptCrafted && m.state == viewReady)) && msg.String() == "c" {
		return m, copyToClipboardCmd(m.viewportContent)
	}

	if m.isPromptCrafted && m.state == viewReady && msg.String() == "r" {
		m.state = viewBusy
		m.busyText = thinkingTextGettingAnswer

		return m, tea.Batch(m.spinner.Tick, sendPromptCmd(m, m.craftedPrompt, false))
	}

	if msg.Type == tea.KeyEnter {
		switch m.state {
		case viewReady:
			m.state = viewBusy
			m.busyText = thinkingTextCrafting
			userInput := m.textInput.Value()

			return m, tea.Batch(m.spinner.Tick, sendPromptCmd(m, userInput, !m.isPromptCrafted))
		case viewResult, viewError:
			m.state = viewReady
			m.isPromptCrafted = false
			m.craftedPrompt = ""
			m.textInput.Reset()
			m.textInput.Placeholder = placeholderRoughPrompt
			m.viewportContent = ""
			m.viewport.SetContent("")

			return m, nil
		case viewBusy:
			// Do nothing.
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

	header := m.headerView()
	footer := m.footerView()
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)

	m.viewport.Height = m.height - headerHeight - footerHeight
	m.viewport.Width = m.width
	m.textInput.Width = m.width - (horizontalPadding * 2)

	mainContent := m.styles.mainContent.Render(m.mainContentView())

	return lipgloss.JoinVertical(lipgloss.Left, header, mainContent, footer)
}

func (m *model) headerView() string {
	left := m.styles.appName.Render(appName) + " " + m.styles.appVersion.Render("("+m.appVersion+")")
	right := m.styles.modelName.Render("Model: " + m.selectedModel)

	spaceWidth := max(0, m.width-lipgloss.Width(left)-lipgloss.Width(right)-(headerPadding*2))

	space := lipgloss.NewStyle().Width(spaceWidth).Render("")

	return m.styles.header.Render(lipgloss.JoinHorizontal(lipgloss.Bottom, left, space, right))
}

func (m *model) mainContentView() string {
	switch m.state {
	case viewBusy:
		return m.spinner.View() + m.busyText
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

func (m *model) footerView() string {
	var footerContent strings.Builder
	footerContent.WriteString("\n")

	if m.state != viewResult {
		footerContent.WriteString(m.styles.input.Render(m.textInput.View()))
		footerContent.WriteString("\n")
	}

	footerContent.WriteString(m.statusBarView())

	return footerContent.String()
}

func (m *model) statusBarView() string {
	if m.statusMessage != "" {
		return m.styles.statusBar.Render(m.statusMessage)
	}

	help := "esc: quit"

	if m.isPromptCrafted && m.state == viewReady {
		resubmitHelp := m.styles.resubmitHelp.Render("r: resubmit")
		help = fmt.Sprintf("%s | c: copy | %s", resubmitHelp, help)
	} else if m.state == viewResult {
		help = "c: copy | " + help
	}

	return m.styles.statusBar.Render(m.styles.statusText.Render(help))
}

// --- Command Logic ---

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

		genConfig := &genai.GenerateContentConfig{Temperature: genai.Ptr(float32(modelTemperature))}

		session, err := m.genaiClient.Chats.Create(m.ctx, m.selectedModel, genConfig, nil)
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

// --- TUI Starter ---

func Start(cfg *config.Config, modelName, version string) error {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.APIKey, Backend: genai.BackendGeminiAPI})
	if err != nil {
		return fmt.Errorf("failed to create generative AI client: %w", err)
	}

	p := tea.NewProgram(New(ctx, client, modelName, version), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI program: %w", err)
	}

	return nil
}
