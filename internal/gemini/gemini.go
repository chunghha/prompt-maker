package gemini

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	ErrModelSelectionCanceled = errors.New("model selection canceled")
	ErrNoModelSelected        = errors.New("no model selected")
	ErrUserInput              = errors.New("failed to get user input")
)

// ModelOption defines the structure for our model choices.
type ModelOption struct {
	Name string
	Desc string
}

// SelectModel displays an interactive form for the user to select a Gemini model.
func SelectModel() (string, error) {
	var selectedModelName string

	// Updated the model list to match the image.
	modelOptions := []ModelOption{
		{"gemini-2.5-flash-lite-preview-06-17", "Latest fast, multi-modal preview model."},
		{"gemini-2.5-flash", "Latest stable flash model."},
		{"gemini-2.5-pro", "Latest stable pro model."},
	}

	// Create the options for the 'huh' form.
	huhOptions := make([]huh.Option[string], len(modelOptions))
	optionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // An orange color

	for i, opt := range modelOptions {
		huhOptions[i] = huh.Option[string]{
			Key:   fmt.Sprintf("%d: %s (%s)", i+1, optionStyle.Render(opt.Name), opt.Desc),
			Value: opt.Name,
		}
	}

	// Create the interactive form.
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Gemini Model").
				Options(huhOptions...).
				Value(&selectedModelName),
		),
	)

	// Run the form and handle the result.
	err := form.Run()
	if err != nil {
		// Handle cases where the user cancels the selection.
		if errors.Is(err, context.Canceled) || errors.Is(err, huh.ErrUserAborted) {
			return "", ErrModelSelectionCanceled
		}

		return "", fmt.Errorf("%w: %w", ErrUserInput, err)
	}

	if selectedModelName == "" {
		return "", ErrNoModelSelected
	}

	return selectedModelName, nil
}
