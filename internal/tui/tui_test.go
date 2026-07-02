package tui

import (
	"context"
	"testing"

	"prompt-maker/internal/gemini"
	"prompt-maker/internal/testutil"
	"prompt-maker/internal/tui/components"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

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

	return &testutil.MockChatSession{}, nil
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
	m := New(context.Background(), &mockChatCreator{}, "v1", "", "", 0.0).(*model)
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
	m := New(context.Background(), &mockChatCreator{}, "v1", "", "", 0.0).(*model)
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

// runUpdateAndFindAIResponse triggers an Update with keyMsg, asserts the model
// transitions to viewBusy, runs the resulting command, and returns the first
// aiResponseMsg found. Fails the test if no aiResponseMsg is produced.
func runUpdateAndFindAIResponse(t *testing.T, m *model, keyMsg tea.Msg) (*model, aiResponseMsg) {
	t.Helper()

	updatedModel, cmd := m.Update(keyMsg)
	m = updatedModel.(*model)

	require.NotNil(t, cmd)
	require.Equal(t, viewBusy, m.state)

	msgs := runCmds(cmd)

	for _, msg := range msgs {
		if ai, ok := msg.(aiResponseMsg); ok {
			return m, ai
		}
	}

	require.Fail(t, "Expected an aiResponseMsg")

	return m, aiResponseMsg{} // unreachable
}

// newMockCreator creates a mockChatCreator that verifies the model name and
// returns the given session.
func newMockCreator(t *testing.T, expectedModel string, session gemini.ChatSession) *mockChatCreator {
	t.Helper()

	return &mockChatCreator{
		createFunc: func(
			_ context.Context, model string, _ *genai.GenerateContentConfig, _ []*genai.Content,
		) (gemini.ChatSession, error) {
			require.Equal(t, expectedModel, model)
			return session, nil
		},
	}
}

func TestUpdate_SubmitRoughPrompt_GeneratesCraftedPrompt(t *testing.T) {
	// Arrange
	const (
		userInput     = "make it a poem"
		craftedPrompt = "This is the crafted poem prompt."
		testModel     = "test-model-123"
	)

	ctx := context.Background()

	mockSession := &testutil.MockChatSession{
		SendMessageFunc: func(_ context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
			require.Contains(t, parts[0].Text, userInput, "Should include user input")

			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{Content: &genai.Content{Parts: []*genai.Part{{Text: craftedPrompt}}}},
				},
			}, nil
		},
	}

	creator := newMockCreator(t, testModel, mockSession)

	m := New(ctx, creator, "v1", "", "", 0.0).(*model)
	// Manually advance state past model selection for the test.
	m.state = viewReady
	m.selectedModel = testModel
	m.textInput.SetValue(userInput)

	// Act
	// 1. User presses Enter, model becomes busy; command runs and returns the crafted prompt.
	m, aiMsg := runUpdateAndFindAIResponse(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	require.Equal(t, craftedPrompt, aiMsg.response)

	// 3. The model processes the AI response.
	updatedModel, cmd := m.Update(aiMsg)
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

	mockSession := &testutil.MockChatSession{
		SendMessageFunc: func(_ context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
			require.Equal(t, craftedPrompt, parts[0].Text, "Should send the crafted prompt directly")

			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{Content: &genai.Content{Parts: []*genai.Part{{Text: finalAnswer}}}},
				},
			}, nil
		},
	}

	creator := newMockCreator(t, testModel, mockSession)

	// Start the model in the state where a prompt has been crafted.
	m := New(ctx, creator, "v1", "", "", 0.0).(*model)
	m.selectedModel = testModel // Set the model
	m.state = viewReady
	m.craftedPrompt = craftedPrompt
	m.textInput.Placeholder = placeholderResubmit

	// Act
	// 1. User presses 'r', model becomes busy; command runs and returns the final answer.
	m, aiMsg := runUpdateAndFindAIResponse(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	require.Equal(t, finalAnswer, aiMsg.response)

	// 3. The model processes the final answer.
	updatedModel, cmd := m.Update(aiMsg)
	m = updatedModel.(*model)

	require.Nil(t, cmd)

	// Assert
	require.Equal(t, viewResult, m.state)
	require.Empty(t, m.craftedPrompt)
	require.Contains(t, m.viewport.View(), finalAnswer)
	require.Equal(t, placeholderNewPrompt, m.textInput.Placeholder)
}

func TestNewStyles(t *testing.T) {
	styles := components.NewStyles()
	require.NotNil(t, styles.Header)
	require.NotNil(t, styles.AppName)
	require.NotNil(t, styles.AppVersion)
	require.NotNil(t, styles.ModelName)
	require.NotNil(t, styles.MainContent)
	require.NotNil(t, styles.Input)
	require.NotNil(t, styles.StatusBar)
	require.NotNil(t, styles.StatusText)
	require.NotNil(t, styles.ResubmitHelp)
	require.NotNil(t, styles.SelectedListItem)
	require.NotNil(t, styles.ListItem)
	require.NotNil(t, styles.Spinner)
}
