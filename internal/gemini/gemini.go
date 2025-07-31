package gemini

import (
	"errors"
)

var (
	ErrNoModelSelected = errors.New("no model selected")
)

type ModelOption struct {
	ModelName string
	ModelDesc string
}

// Name returns the name of the model.
func (m ModelOption) Name() string {
	return m.ModelName
}

// Desc returns the description of the model.
func (m ModelOption) Desc() string {
	return m.ModelDesc
}

// FilterValue implements the list.Item interface.
func (m ModelOption) FilterValue() string {
	return m.ModelName
}

// GetModelOptions is an exported function that constructs and returns the model list.
func GetModelOptions() []ModelOption {
	return []ModelOption{
		{ModelName: "gemini-2.5-flash-lite", ModelDesc: "Latest fast, multi-modal model."},
		{ModelName: "gemini-2.5-flash", ModelDesc: "Latest stable flash model."},
		{ModelName: "gemini-2.5-pro", ModelDesc: "Latest stable pro model."},
	}
}
