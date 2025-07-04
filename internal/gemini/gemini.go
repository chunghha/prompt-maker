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

type ModelOption struct {
	Name string
	Desc string
}

// GetModelOptions is an exported function that constructs and returns the model list.
// There is no package-level variable anymore.
func GetModelOptions() []ModelOption {
	return []ModelOption{
		{"gemini-2.5-flash-lite-preview-06-17", "Latest fast, multi-modal preview model."},
		{"gemini-2.5-flash", "Latest stable flash model."},
		{"gemini-2.5-pro", "Latest stable pro model."},
	}
}

// SelectModel now calls GetModelOptions() to get the list of models.
func SelectModel() (string, error) {
	var selectedModelName string

	// Get the models from our single source of truth function.
	opts := GetModelOptions()

	huhOptions := make([]huh.Option[string], len(opts))
	optionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))

	for i, opt := range opts {
		huhOptions[i] = huh.Option[string]{
			Key:   fmt.Sprintf("%d: %s (%s)", i+1, optionStyle.Render(opt.Name), opt.Desc),
			Value: opt.Name,
		}
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Gemini Model").
				Options(huhOptions...).
				Value(&selectedModelName),
		),
	)

	err := form.Run()
	if err != nil {
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
