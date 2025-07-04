package web

import (
	"context"
	"prompt-maker/internal/config"
	"prompt-maker/internal/prompt"

	"google.golang.org/genai"
)

// PromptGenerator defines the interface for our prompt generation logic.
type PromptGenerator interface {
	Generate(ctx context.Context, userInput string) (string, error)
}

// geminiPromptGenerator is the real implementation that talks to the Gemini API.
type geminiPromptGenerator struct {
	client    *genai.Client
	modelName string
}

func NewGeminiPromptGenerator(client *genai.Client, modelName string) PromptGenerator {
	return &geminiPromptGenerator{
		client:    client,
		modelName: modelName,
	}
}

func (g *geminiPromptGenerator) Generate(ctx context.Context, userInput string) (string, error) {
	genConfig := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(config.DefaultModelTemperature)),
	}

	session, err := g.client.Chats.Create(ctx, g.modelName, genConfig, nil)
	if err != nil {
		return "", err
	}

	return prompt.Generate(ctx, session, userInput)
}
