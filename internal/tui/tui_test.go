package tui

import (
	"context"
	"errors"
	"testing"

	"prompt-maker/internal/gemini"

	"github.com/charmbracelet/bubbles/list"
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

// mockModelOption is a mock implementation of the modelOption interface for testing.
type mockModelOption struct {
	name string
	desc string
}

func (m mockModelOption) Name() string {
	return m.name
}

func (m mockModelOption) Desc() string {
	return m.desc
}

func (m mockModelOption) FilterValue() string {
	return m.name
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

func TestUpdate_ModelSelection_UpdatesState(t *testing.T) {
	// Arrange
	m := New(context.Background(), &mockChatCreator{}, "v1").(*model)
	require.Equal(t, viewSelectingModel, m.state)

	// Act
	selectedModel := mockModelOption{name: "test-model", desc: "Test model description"}
	m.modelList.SetItems([]list.Item{selectedModel})
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(*model)

	// Assert
	require.Nil(t, cmd)
	require.Equal(t, viewReady, m.state)
	require.Equal(t, "test-model", m.selectedModel)
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
			require.Contains(t, parts[0].Text, userInput, "Should include user input")
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

	m := New(ctx, creator, "v1").(*model)
	// Manually advance state past model selection for the test.
	m.state = viewReady
	m.selectedModel = testModel
	m.textInput.SetValue(userInput)

	// Act
	// 1. User presses Enter, model becomes busy and returns a command.
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(*model)

	require.NotNil(t, cmd)
	require.Equal(t, viewBusy, m.state)

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
	updatedModel, cmd = m.Update(aiMsg)
	m = updatedModel.(*model)

	require.Nil(t, cmd)

	// Assert
	require.Equal(t, viewReady, m.state)
	require.Equal(t, craftedPrompt, m.craftedPrompt)
	require.Contains(t, m.viewport.View(), craftedPrompt)
	require.Equal(t, placeholderResubmit, m.textInput.Placeholder)
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
	m := New(ctx, creator, "v1").(*model)
	m.selectedModel = testModel // Set the model
	m.state = viewReady
	m.craftedPrompt = craftedPrompt
	m.textInput.Placeholder = placeholderResubmit

	// Act
	// 1. User presses 'r', model becomes busy and returns a command.
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updatedModel.(*model)

	require.NotNil(t, cmd)
	require.Equal(t, viewBusy, m.state)

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
	updatedModel, cmd = m.Update(aiMsg)
	m = updatedModel.(*model)

	require.Nil(t, cmd)

	// Assert
	require.Equal(t, viewResult, m.state)
	require.Empty(t, m.craftedPrompt)
	require.Contains(t, m.viewport.View(), finalAnswer)
	require.Equal(t, placeholderNewPrompt, m.textInput.Placeholder)
}

func TestNewStyles(t *testing.T) {
	styles := newStyles()
	require.NotNil(t, styles.header)
	require.NotNil(t, styles.appName)
	require.NotNil(t, styles.appVersion)
	require.NotNil(t, styles.modelName)
	require.NotNil(t, styles.mainContent)
	require.NotNil(t, styles.input)
	require.NotNil(t, styles.statusBar)
	require.NotNil(t, styles.statusText)
	require.NotNil(t, styles.resubmitHelp)
	require.NotNil(t, styles.selectedListItem)
	require.NotNil(t, styles.listItem)
	require.NotNil(t, styles.spinner)
}
