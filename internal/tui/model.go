package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"prompt-maker/internal/gemini"
	"prompt-maker/internal/tui/components"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
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
	craftedPrompt      string
	busyText           string
	errorMessage       string
	statusMessage      string
	rawViewportContent string
	width              int
	height             int
	styles             components.Styles
}

func New(ctx context.Context, chatSvc chatCreator, version string) tea.Model {
	// Create items for the list.
	modelOptions := gemini.GetModelOptions()

	items := make([]list.Item, len(modelOptions))
	for i, opt := range modelOptions {
		items[i] = opt
	}

	// Setup the list component.
	l := list.New(items, components.ItemDelegate{}, initialViewportWidth, modelListHeight)
	l.Title = "Select a Gemini Model"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	ti := textinput.New()
	ti.Placeholder = placeholderRoughPrompt
	ti.Focus()
	ti.CharLimit = textInputCharLimit

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = components.NewStyles().Spinner

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
		styles:          components.NewStyles(),
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
	case viewReady:
		return m.updateReady(msg)
	case viewBusy:
		return m.updateBusy(msg)
	case viewResult:
		return m.updateResult(msg)
	case viewError:
		return m.updateError(msg)
	default:
		return m, nil
	}
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
		return m.styles.MainContent.Render(m.modelList.View())
	}

	header := m.headerView()
	footer := m.footerView()
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)

	m.viewport.Height = m.height - headerHeight - footerHeight
	m.viewport.Width = m.width
	m.textInput.Width = m.width - (horizontalPadding * 2)

	mainContent := m.styles.MainContent.Render(m.mainContentView())

	return lipgloss.JoinVertical(lipgloss.Left, header, mainContent, footer)
}

// updateModelSelection handles logic for the new initial view.
func (m *model) updateModelSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEnter {
		if i, ok := m.modelList.SelectedItem().(components.ModelOption); ok {
			m.selectedModel = i.Name()
			m.state = viewReady // Transition to the main view
		}

		return m, nil
	}

	var cmd tea.Cmd

	m.modelList, cmd = m.modelList.Update(msg)

	return m, cmd
}

func (m *model) updateReady(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *model) updateBusy(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case aiResponseMsg:
		return m.handleAIResponse(msg)
	case errMsg:
		return m.handleError(msg)
	}

	return m.updateComponents(msg)
}

func (m *model) updateResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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

func (m *model) updateError(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return m.handleKeyMsg(keyMsg)
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

	if m.craftedPrompt == "" {
		m.craftedPrompt = msg.response
		m.textInput.Reset()
		m.textInput.Placeholder = placeholderResubmit
		m.rawViewportContent = msg.response
		m.viewport.SetContent(renderedContent)
		m.state = viewReady
	} else {
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
	switch {
	case msg.String() == "c" && (m.state == viewResult || (m.craftedPrompt != "" && m.state == viewReady)):
		return m, copyToClipboardCmd(m.rawViewportContent)
	case msg.String() == "r" && m.craftedPrompt != "" && m.state == viewReady:
		return m.resubmitPrompt()
	case msg.Type == tea.KeyEnter:
		return m.handleEnterKey()
	}

	return m.updateComponents(msg)
}

func (m *model) resubmitPrompt() (tea.Model, tea.Cmd) {
	m.state = viewBusy
	m.busyText = thinkingTextGettingAnswer

	return m, tea.Batch(m.spinner.Tick, sendPromptCmd(m, m.craftedPrompt, false))
}

func (m *model) handleEnterKey() (tea.Model, tea.Cmd) {
	switch m.state {
	case viewReady:
		return m.submitPrompt()
	case viewResult, viewError:
		m.resetToReady()
		return m, nil
	case viewSelectingModel, viewBusy:
		// Do nothing in these states.
	}

	return m, nil
}

func (m *model) submitPrompt() (tea.Model, tea.Cmd) {
	m.state = viewBusy
	m.busyText = thinkingTextCrafting
	userInput := m.textInput.Value()

	return m, tea.Batch(m.spinner.Tick, sendPromptCmd(m, userInput, m.craftedPrompt == ""))
}

func (m *model) resetToReady() {
	m.state = viewReady
	m.craftedPrompt = ""
	m.textInput.Reset()
	m.textInput.Placeholder = placeholderRoughPrompt
	m.rawViewportContent = ""
	m.viewport.SetContent("")
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

func (m *model) headerView() string {
	left := m.styles.AppName.Render(appName) + " " + m.styles.AppVersion.Render("("+m.appVersion+")")
	right := m.styles.ModelName.Render("Model: " + m.selectedModel)

	spaceWidth := max(0, m.width-lipgloss.Width(left)-lipgloss.Width(right)-(headerPadding*2))
	space := lipgloss.NewStyle().Width(spaceWidth).Render("")

	return m.styles.Header.Render(lipgloss.JoinHorizontal(lipgloss.Bottom, left, space, right))
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
		footerContent.WriteString(m.styles.Input.Render(m.textInput.View()))
		footerContent.WriteString("\n")
	}

	footerContent.WriteString(m.statusBarView())

	return footerContent.String()
}

func (m *model) statusBarView() string {
	if m.statusMessage != "" {
		return m.styles.StatusBar.Render(m.statusMessage)
	}

	help := "esc: quit"

	if m.craftedPrompt != "" && m.state == viewReady {
		resubmitHelp := m.styles.ResubmitHelp.Render("r: resubmit")
		help = fmt.Sprintf("%s | c: copy | %s", resubmitHelp, help)
	} else if m.state == viewResult {
		help = "c: copy | " + help
	}

	return m.styles.StatusBar.Render(m.styles.StatusText.Render(help))
}
