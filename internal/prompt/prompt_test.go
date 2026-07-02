package prompt

import (
	"context"
	"strings"
	"testing"

	"prompt-maker/internal/testutil"

	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

func TestGenerate(t *testing.T) {
	ctx := context.Background()
	userInput := "convert a function to a class"
	expectedAnswer := "This is the optimized prompt."

	mockCS := &testutil.MockChatSession{
		SendMessageFunc: func(_ context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
			require.Len(t, parts, 1)

			sentText := parts[0].Text
			require.True(t, strings.HasPrefix(sentText, "You are Lyra"), "The prompt must start with the Lyra system prompt.")
			require.True(t, strings.HasSuffix(sentText, userInput), "The prompt must end with the user's input.")

			// Return a simulated response.
			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
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

func TestLyraPrompt(t *testing.T) {
	require.NotEmpty(t, LyraPrompt, "LyraPrompt should not be empty")
}
