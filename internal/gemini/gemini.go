package gemini

import (
	"errors"
)

var (
	ErrNoModelSelected = errors.New("no model selected")
)

type ModelOption struct {
	Name string
	Desc string
}

// FilterValue implements the list.Item interface.
func (m ModelOption) FilterValue() string {
	return m.Name
}

// GetModelOptions is an exported function that constructs and returns the model list.
func GetModelOptions() []ModelOption {
	return []ModelOption{
		{"gemini-2.5-flash-lite-preview-06-17", "Latest fast, multi-modal preview model."},
		{"gemini-2.5-flash", "Latest stable flash model."},
		{"gemini-2.5-pro", "Latest stable pro model."},
	}
}
