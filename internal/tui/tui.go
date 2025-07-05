package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"prompt-maker/internal/config"
	"prompt-maker/internal/gemini"
	"prompt-maker/internal/prompt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"google.golang.org/genai"
)

// Constants.
const (
	appName                   = "Prompt Maker"
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
	listHorizontalPadding     = 2
	modelListHeight           = 14
)

var (
	errPromptEmpty    = errors.New("prompt cannot be empty")
	errClipboardWrite = errors.New("failed to write to clipboard")
)

// TUI Messages.
type aiResponseMsg struct{ response string }
type errMsg struct{ err error }
type statusMessage string
type clearStatusMsg struct{}

func (e errMsg) Error() string { return e.err.Error() }

// --- TUI Model ---

// chatCreator defines an interface for creating chat sessions.
type chatCreator interface {
	Create(
		ctx context.Context,
		model string,
		genConfig *genai.GenerateContentConfig,
		history []*genai.Content,
	) (gemini.ChatSession, error)
}

// genaiChatCreator holds the genai.Client to satisfy the chatCreator interface.
type genaiChatCreator struct {
	client *genai.Client
}

// Create satisfies the chatCreator interface for the real implementation.
func (c *genaiChatCreator) Create(
	ctx context.Context,
	model string,
	genConfig *genai.GenerateContentConfig,
	history []*genai.Content,
) (gemini.ChatSession, error) {
	return c.client.Chats.Create(ctx, model, genConfig, history)
}

type viewState int

const (
	viewSelectingModel viewState = iota // New initial state
	viewReady
	viewBusy
	viewResult
	viewError
)

type model struct {
	ctx                context.Context
	state              viewState
	modelList          list.Model // New list component for model selection
	textInput          textinput.Model
	spinner            spinner.Model
	viewport           viewport.Model
	glamourRenderer    *glamour.TermRenderer
	chatSvc            chatCreator
	selectedModel      string
	appVersion         string
	quitting           bool
	isPromptCrafted    bool
	craftedPrompt      string
	busyText           string
	errorMessage       string
	statusMessage      string
	rawViewportContent string
	width              int
	height             int
	styles             Styles
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

// itemDelegate for the model selection list.
type itemDelegate struct{}

func (itemDelegate) Height() int                             { return 1 }
func (itemDelegate) Spacing() int                            { return 0 }
func (itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

//nolint:gocritic // The signature is defined by the list.ItemDelegate interface, which requires passing list.Model by value.
func (itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(gemini.ModelOption)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s (%s)", index+1, i.Name, i.Desc)

	fn := lipgloss.NewStyle().Padding(0, 0, 0, listHorizontalPadding).Render
	if index == m.Index() {
		fn = func(s ...string) string {
			style := lipgloss.NewStyle().Padding(0, 0, 0, listHorizontalPadding).Foreground(lipgloss.Color("208"))
			return style.Render("> " + strings.Join(s, " "))
		}
	}

	_, _ = io.WriteString(w, fn(str))
}

func New(ctx context.Context, chatSvc chatCreator, version string) tea.Model {
	// Create items for the list.
	modelOptions := gemini.GetModelOptions()

	items := make([]list.Item, len(modelOptions))
	for i, opt := range modelOptions {
		items[i] = opt
	}

	// Setup the list component.
	l := list.New(items, itemDelegate{}, initialViewportWidth, modelListHeight)
	l.Title = "Select a Gemini Model"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	ti := textinput.New()
	ti.Placeholder = placeholderRoughPrompt
	ti.Focus()
	ti.CharLimit = textInputCharLimit

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(initialViewportWidth, initialViewportHeight)

	renderer, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())

	return &model{
		ctx:             ctx,
		state:           viewSelectingModel, // Start at the new selection view
		modelList:       l,
		textInput:       ti,
		spinner:         s,
		viewport:        vp,
		glamourRenderer: renderer,
		chatSvc:         chatSvc,
		appVersion:      version,
		styles:          newStyles(),
	}
}

func (*model) Init() tea.Cmd {
	return textinput.Blink
}

// --- Update Logic ---

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle top-level messages that apply to all states.
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		// Global quit works in any state.
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Route all other messages based on the current view state.
	switch m.state {
	case viewSelectingModel:
		return m.updateModelSelection(msg)
	case viewReady, viewBusy, viewResult, viewError:
		return m.updateMain(msg)
	default:
		return m, nil
	}
}

// updateModelSelection handles logic for the new initial view.
func (m *model) updateModelSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEnter {
		if i, ok := m.modelList.SelectedItem().(gemini.ModelOption); ok {
			m.selectedModel = i.Name
			m.state = viewReady // Transition to the main view
		}

		return m, nil
	}

	var cmd tea.Cmd

	m.modelList, cmd = m.modelList.Update(msg)

	return m, cmd
}

// updateMain contains the original update logic for all states after model selection.
func (m *model) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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

	m.modelList.SetWidth(msg.Width)

	// Re-create the glamour renderer with the new width.
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.width-(horizontalPadding*2)),
	)
	if err == nil {
		m.glamourRenderer = renderer
	}

	// Re-render the content with the new renderer settings.
	if m.rawViewportContent != "" {
		rendered, _ := m.glamourRenderer.Render(m.rawViewportContent)
		m.viewport.SetContent(rendered)
	}

	return m, nil
}

func (m *model) handleAIResponse(msg aiResponseMsg) (tea.Model, tea.Cmd) {
	var renderedContent string

	if m.glamourRenderer != nil {
		rendered, err := m.glamourRenderer.Render(msg.response)
		if err == nil {
			renderedContent = rendered
		} else {
			renderedContent = msg.response
		}
	} else {
		renderedContent = msg.response
	}

	if !m.isPromptCrafted {
		m.isPromptCrafted = true
		m.craftedPrompt = msg.response
		m.textInput.Reset()
		m.textInput.Placeholder = placeholderResubmit
		m.rawViewportContent = msg.response
		m.viewport.SetContent(renderedContent)
		m.state = viewReady
	} else {
		m.isPromptCrafted = false
		m.craftedPrompt = ""
		m.rawViewportContent = msg.response
		m.viewport.SetContent(renderedContent)
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
	m.rawViewportContent = errorText + m.errorMessage
	m.viewport.SetContent(m.rawViewportContent)

	return m, nil
}

func (m *model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if (m.state == viewResult || (m.isPromptCrafted && m.state == viewReady)) && msg.String() == "c" {
		return m, copyToClipboardCmd(m.rawViewportContent)
	}

	if m.isPromptCrafted && m.state == viewReady && msg.String() == "r" {
		m.state = viewBusy
		m.busyText = thinkingTextGettingAnswer

		return m, tea.Batch(m.spinner.Tick, sendPromptCmd(m, m.craftedPrompt, false))
	}

	if msg.Type == tea.KeyEnter {
		//nolint:exhaustive // This switch is inside the main update loop, which already filters by state.
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
			m.rawViewportContent = ""
			m.viewport.SetContent("")

			return m, nil
		case viewBusy:
			// Do nothing.
		}
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

	if m.state == viewSelectingModel {
		return m.styles.mainContent.Render(m.modelList.View())
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

	spaceWidth := m.width - lipgloss.Width(left) - lipgloss.Width(right) - (headerPadding * 2)
	if spaceWidth < 0 {
		spaceWidth = 0
	}

	space := lipgloss.NewStyle().Width(spaceWidth).Render("")

	return m.styles.header.Render(lipgloss.JoinHorizontal(lipgloss.Bottom, left, space, right))
}

func (m *model) mainContentView() string {
	//nolint:exhaustive // The viewSelectingModel state is handled in the parent View() function.
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

	return ""
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

// --- TUI Starter ---

// Start no longer takes a modelName.
func Start(cfg *config.Config, version string) error {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.APIKey, Backend: genai.BackendGeminiAPI})
	if err != nil {
		return fmt.Errorf("failed to create generative AI client: %w", err)
	}

	creator := &genaiChatCreator{client: client}

	p := tea.NewProgram(New(ctx, creator, version), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI program: %w", err)
	}

	return nil
}
