package tui

import (
	"context"
	"errors"
	"fmt"
	"time"

	"prompt-maker/internal/config"
	"prompt-maker/internal/gemini"

	tea "github.com/charmbracelet/bubbletea"
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
	goodbyeText               = "Goodbye!\n"
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

// --- TUI Starter ---

// Start no longer takes a modelName.
func Start(cfg *config.Config, version, modelName, history string, temperature float32) error {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.APIKey, Backend: genai.BackendGeminiAPI})
	if err != nil {
		return fmt.Errorf("failed to create generative AI client: %w", err)
	}

	creator := &genaiChatCreator{client: client}

	p := tea.NewProgram(New(ctx, creator, version, modelName, history, temperature), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI program: %w", err)
	}

	return nil
}
