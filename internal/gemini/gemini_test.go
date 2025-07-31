package gemini

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModelOption_Name(t *testing.T) {
	m := ModelOption{ModelName: "test-model"}
	assert.Equal(t, "test-model", m.Name())
}

func TestModelOption_Desc(t *testing.T) {
	m := ModelOption{ModelDesc: "Test Description"}
	assert.Equal(t, "Test Description", m.Desc())
}

func TestModelOption_FilterValue(t *testing.T) {
	m := ModelOption{ModelName: "filter-model"}
	assert.Equal(t, "filter-model", m.FilterValue())
}

func TestGetModelOptions(t *testing.T) {
	options := GetModelOptions()
	assert.Len(t, options, 3)

	expectedModels := []ModelOption{
		{ModelName: "gemini-2.5-flash-lite", ModelDesc: "Latest fast, multi-modal model."},
		{ModelName: "gemini-2.5-flash", ModelDesc: "Latest stable flash model."},
		{ModelName: "gemini-2.5-pro", ModelDesc: "Latest stable pro model."},
	}

	assert.ElementsMatch(t, expectedModels, options)
}
