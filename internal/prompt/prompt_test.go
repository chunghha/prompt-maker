package prompt

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

var errSendMessageNotImplemented = errors.New("sendMessageFunc not implemented")

// mockChat is a test double for the gemini.ChatSession.
type mockChat struct {
	sendMessageFunc func(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

// SendMessage implements the gemini.ChatSession interface for our mock.
func (m *mockChat) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, parts...)
	}

	return nil, errSendMessageNotImplemented
}

func TestGenerate(t *testing.T) {
	ctx := context.Background()
	userInput := "convert a function to a class"
	expectedAnswer := "This is the optimized prompt."

	mockCS := &mockChat{
		sendMessageFunc: func(_ context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
			require.Len(t, parts, 1)

			sentText := parts[0].Text
			require.True(t, strings.HasPrefix(sentText, "You are Lyra"), "The prompt must start with the Lyra system prompt.")
			require.True(t, strings.HasSuffix(sentText, userInput), "The prompt must end with the user's input.")

			// Return a simulated response.
			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							// THIS IS THE FIX: The Parts slice must be a slice of POINTERS (*genai.Part).
							Parts: []*genai.Part{{Text: expectedAnswer}},
						},
					},
				},
			}, nil
		},
	}

	answer, err := Generate(ctx, mockCS, userInput)

	require.NoError(t, err)
	require.Equal(t, expectedAnswer, answer)
}
