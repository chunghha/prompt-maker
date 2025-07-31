package prompt

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"prompt-maker/internal/gemini"

	"google.golang.org/genai"
)

//go:embed lyra.txt
var LyraPrompt string

var (
	ErrSendMessage          = errors.New("error sending message to Gemini")
	ErrNoResponseCandidates = errors.New("received no response candidates from model")
)

// Generate creates an optimized prompt by sending the user's input along with the Lyra system prompt to the Gemini model.
func Generate(ctx context.Context, cs gemini.ChatSession, userInput string) (string, error) {
	fullPrompt := LyraPrompt + userInput

	resp, err := cs.SendMessage(ctx, genai.Part{Text: fullPrompt})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSendMessage, err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", ErrNoResponseCandidates
	}

	var b strings.Builder

	for _, part := range resp.Candidates[0].Content.Parts {
		if txt := part.Text; txt != "" {
			b.WriteString(txt)
		}
	}

	return b.String(), nil
}

// Execute sends a prompt to the Gemini model without any system prompt.
func Execute(ctx context.Context, cs gemini.ChatSession, userInput string) (string, error) {
	resp, err := cs.SendMessage(ctx, genai.Part{Text: userInput})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSendMessage, err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		// Corrected the error variable name.
		return "", ErrNoResponseCandidates
	}

	return resp.Text(), nil
}
