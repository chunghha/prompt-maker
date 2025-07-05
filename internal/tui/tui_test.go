package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"prompt-maker/internal/gemini"
	"prompt-maker/internal/prompt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

// mockChatSession implements gemini.ChatSession for testing.
var errSendMessageNotImplemented = errors.New("SendMessage not implemented")

type mockChatSession struct {
	sendMessageFunc func(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

func (m *mockChatSession) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, parts...)
	}

	return nil, errSendMessageNotImplemented
}

// mockChatCreator implements the chatCreator interface for testing.
type mockChatCreator struct {
	createFunc func(
		ctx context.Context, model string, genConfig *genai.GenerateContentConfig, history []*genai.Content,
	) (gemini.ChatSession, error)
}

func (m *mockChatCreator) Create(
	ctx context.Context, model string, genConfig *genai.GenerateContentConfig, history []*genai.Content,
) (gemini.ChatSession, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, model, genConfig, history)
	}

	return &mockChatSession{}, nil
}

// runCmds executes a tea.Cmd, handling batches, and returns all resulting messages.
func runCmds(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}

	var msgs []tea.Msg

	msg := cmd()

	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			msgs = append(msgs, c())
		}
	} else {
		msgs = append(msgs, msg)
	}

	return msgs
}

func TestUpdate_SubmitEmptyPrompt_ReturnsError(t *testing.T) {
	// Arrange
	m := New(context.Background(), &mockChatCreator{}, "v1").(*model)
	// Manually advance state past model selection for the test.
	m.state = viewReady
	m.selectedModel = "test-model"
	m.textInput.SetValue("") // Ensure input is empty

	// Act
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msgs := runCmds(cmd)

	// Assert
	var found bool

	for _, msg := range msgs {
		if err, ok := msg.(errMsg); ok {
			require.ErrorIs(t, err.err, errPromptEmpty)

			found = true

			break
		}
	}

	require.True(t, found, "Expected an errMsg")
}

func TestUpdate_SubmitRoughPrompt_GeneratesCraftedPrompt(t *testing.T) {
	// Arrange
	const (
		userInput     = "make it a poem"
		craftedPrompt = "This is the crafted poem prompt."
		testModel     = "test-model-123"
	)

	ctx := context.Background()

	mockSession := &mockChatSession{
		sendMessageFunc: func(_ context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
			require.True(t, strings.HasPrefix(parts[0].Text, prompt.LyraPrompt), "Should use Lyra system prompt")
			require.True(t, strings.HasSuffix(parts[0].Text, userInput), "Should include user input")
			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{Content: &genai.Content{Parts: []*genai.Part{{Text: craftedPrompt}}}},
				},
			}, nil
		},
	}

	creator := &mockChatCreator{
		createFunc: func(
			_ context.Context, model string, _ *genai.GenerateContentConfig, _ []*genai.Content,
		) (gemini.ChatSession, error) {
			require.Equal(t, testModel, model)
			return mockSession, nil
		},
	}

	m := New(ctx, creator, "v1")
	// Manually advance state past model selection for the test.
	m.(*model).state = viewReady
	m.(*model).selectedModel = testModel
	m.(*model).textInput.SetValue(userInput)

	// Act
	// 1. User presses Enter, model becomes busy and returns a command.
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	require.Equal(t, viewBusy, m.(*model).state)

	// 2. The command runs and returns messages.
	msgs := runCmds(cmd)

	var (
		aiMsg aiResponseMsg
		found bool
	)

	for _, msg := range msgs {
		if m, ok := msg.(aiResponseMsg); ok {
			aiMsg = m
			found = true

			break
		}
	}

	require.True(t, found, "Expected an aiResponseMsg")
	require.Equal(t, craftedPrompt, aiMsg.response)

	// 3. The model processes the AI response.
	m, cmd = m.Update(aiMsg)
	require.Nil(t, cmd)

	// Assert
	finalModel := m.(*model)
	require.Equal(t, viewReady, finalModel.state)
	require.True(t, finalModel.isPromptCrafted, "isPromptCrafted should be true")
	require.Equal(t, craftedPrompt, finalModel.craftedPrompt)
	require.Contains(t, finalModel.viewport.View(), craftedPrompt)
	require.Equal(t, placeholderResubmit, finalModel.textInput.Placeholder)
}

func TestUpdate_ResubmitCraftedPrompt_GetsFinalAnswer(t *testing.T) {
	// Arrange
	const (
		craftedPrompt = "This is the crafted prompt."
		finalAnswer   = "This is the final answer."
		testModel     = "test-model-456"
	)

	ctx := context.Background()

	mockSession := &mockChatSession{
		sendMessageFunc: func(_ context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
			require.Equal(t, craftedPrompt, parts[0].Text, "Should send the crafted prompt directly")
			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{Content: &genai.Content{Parts: []*genai.Part{{Text: finalAnswer}}}},
				},
			}, nil
		},
	}

	creator := &mockChatCreator{
		createFunc: func(
			_ context.Context, model string, _ *genai.GenerateContentConfig, _ []*genai.Content,
		) (gemini.ChatSession, error) {
			require.Equal(t, testModel, model)
			return mockSession, nil
		},
	}

	// Start the model in the state where a prompt has been crafted.
	m := New(ctx, creator, "v1")
	m.(*model).selectedModel = testModel // Set the model
	m.(*model).state = viewReady
	m.(*model).isPromptCrafted = true
	m.(*model).craftedPrompt = craftedPrompt
	m.(*model).textInput.Placeholder = placeholderResubmit

	// Act
	// 1. User presses 'r', model becomes busy and returns a command.
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	require.NotNil(t, cmd)
	require.Equal(t, viewBusy, m.(*model).state)

	// 2. The command runs and returns the final answer.
	msgs := runCmds(cmd)

	var (
		aiMsg aiResponseMsg
		found bool
	)

	for _, msg := range msgs {
		if m, ok := msg.(aiResponseMsg); ok {
			aiMsg = m
			found = true

			break
		}
	}

	require.True(t, found, "Expected an aiResponseMsg")
	require.Equal(t, finalAnswer, aiMsg.response)

	// 3. The model processes the final answer.
	m, cmd = m.Update(aiMsg)
	require.Nil(t, cmd)

	// Assert
	finalModel := m.(*model)
	require.Equal(t, viewResult, finalModel.state)
	require.False(t, finalModel.isPromptCrafted, "isPromptCrafted should be false")
	require.Empty(t, finalModel.craftedPrompt)
	require.Contains(t, finalModel.viewport.View(), finalAnswer)
	require.Equal(t, placeholderNewPrompt, finalModel.textInput.Placeholder)
}
