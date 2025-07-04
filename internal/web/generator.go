package web

import (
	"context"
	"prompt-maker/internal/config"
	"prompt-maker/internal/prompt"

	"google.golang.org/genai"
)

// PromptGenerator now includes the Execute method for the second step.
type PromptGenerator interface {
	Generate(ctx context.Context, userInput string) (string, error)
	Execute(ctx context.Context, userInput string) (string, error)
}

// geminiPromptGenerator is the real implementation.
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

// Execute implements the second step of the workflow, getting the final answer.
func (g *geminiPromptGenerator) Execute(ctx context.Context, userInput string) (string, error) {
	genConfig := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(config.DefaultModelTemperature)),
	}

	session, err := g.client.Chats.Create(ctx, g.modelName, genConfig, nil)
	if err != nil {
		return "", err
	}
	// Use the existing core prompt.Execute function.
	return prompt.Execute(ctx, session, userInput)
}
