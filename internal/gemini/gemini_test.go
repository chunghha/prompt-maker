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
	assert.NotEmpty(t, options, "GetModelOptions should return at least one model")

	for _, opt := range options {
		assert.NotEmpty(t, opt.ModelName, "ModelName should not be empty")
		assert.NotEmpty(t, opt.ModelDesc, "ModelDesc should not be empty")
		assert.Equal(t, opt.ModelName, opt.FilterValue(), "FilterValue should return ModelName")
	}
}
