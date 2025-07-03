package gemini

import (
	"context"

	"google.golang.org/genai"
)

// ChatSession defines the interface for a chat session.
type ChatSession interface {
	SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}
