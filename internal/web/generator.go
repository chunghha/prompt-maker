package web

import (
	"context"
	"prompt-maker/internal/config"
	"prompt-maker/internal/prompt"

	"google.golang.org/genai"
)

// PromptGenerator methods now accept the modelName for each request.
type PromptGenerator interface {
	Generate(ctx context.Context, modelName, userInput string) (string, error)
	Execute(ctx context.Context, modelName, userInput string) (string, error)
}

// The struct no longer needs to store the modelName.
type geminiPromptGenerator struct {
	client *genai.Client
}

// The constructor no longer needs the modelName.
func NewGeminiPromptGenerator(client *genai.Client) PromptGenerator {
	return &geminiPromptGenerator{
		client: client,
	}
}

// Generate now uses the passed-in modelName.
func (g *geminiPromptGenerator) Generate(ctx context.Context, modelName, userInput string) (string, error) {
	genConfig := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(config.DefaultModelTemperature)),
	}

	session, err := g.client.Chats.Create(ctx, modelName, genConfig, nil)
	if err != nil {
		return "", err
	}

	return prompt.Generate(ctx, session, userInput)
}

// Execute now uses the passed-in modelName.
func (g *geminiPromptGenerator) Execute(ctx context.Context, modelName, userInput string) (string, error) {
	genConfig := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(config.DefaultModelTemperature)),
	}

	session, err := g.client.Chats.Create(ctx, modelName, genConfig, nil)
	if err != nil {
		return "", err
	}

	return prompt.Execute(ctx, session, userInput)
}
