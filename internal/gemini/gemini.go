package gemini

import "errors"

// This file now only contains model selection logic.
// The ChatSession interface has been moved to chat.go to avoid redeclaration.

var ErrModelSelectionCanceled = errors.New("model selection canceled")

// SelectModel will eventually show an interactive list. For now, it returns a default.
func SelectModel() (string, error) {
	return "gemini-2.5-flash", nil
}
