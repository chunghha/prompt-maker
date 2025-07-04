package web

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"prompt-maker/internal/config" // Import config to use the DefaultModel constant

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// mockPromptGenerator is updated to match the new interface signatures.
type mockPromptGenerator struct {
	GenerateFunc func(ctx context.Context, modelName, userInput string) (string, error)
	ExecuteFunc  func(ctx context.Context, modelName, userInput string) (string, error)
}

func (m *mockPromptGenerator) Generate(ctx context.Context, modelName, userInput string) (string, error) {
	return m.GenerateFunc(ctx, modelName, userInput)
}

func (m *mockPromptGenerator) Execute(ctx context.Context, modelName, userInput string) (string, error) {
	return m.ExecuteFunc(ctx, modelName, userInput)
}

func TestHandleIndex(t *testing.T) {
	mockGen := &mockPromptGenerator{}
	// Add Version to the config.
	server, err := NewServer(Config{Generator: mockGen, Version: "test-version"})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `id="prompt-form"`)
}

func TestHandlePrompt(t *testing.T) {
	const (
		userInput        = "make it better"
		expectedResponse = "This is the crafted prompt."
		selectedModel    = "gemini-2.5-flash"
	)

	mockGen := &mockPromptGenerator{
		GenerateFunc: func(_ context.Context, model, input string) (string, error) {
			require.Equal(t, selectedModel, model)
			require.Equal(t, userInput, input)
			return expectedResponse, nil
		},
	}

	server, err := NewServer(Config{Generator: mockGen, Version: "test"})
	require.NoError(t, err)

	form := strings.NewReader("prompt=" + userInput + "&model=" + selectedModel)
	req := httptest.NewRequest(http.MethodPost, "/prompt", form)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	w := httptest.NewRecorder()

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), expectedResponse)
	// Remove the self-closing slash from the assertion to match HTML5 output.
	require.Contains(t, w.Body.String(), `<input type="hidden" name="model" value="gemini-2.5-flash">`)
}

func TestHandleIndex_WithDaisyUI(t *testing.T) {
	mockGen := &mockPromptGenerator{}
	// Add a version to the config.
	server, err := NewServer(Config{Generator: mockGen, Version: "test-version"})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// Change assertion to look for "Model:" instead of "Default Model:".
	require.Contains(t, w.Body.String(), "Model: "+config.DefaultModel)
}

func TestHandleExecute(t *testing.T) {
	const (
		craftedPrompt = "This is the crafted prompt."
		finalAnswer   = "This is the final answer from the AI."
		selectedModel = "gemini-2.5-flash"
	)

	mockGen := &mockPromptGenerator{
		// The signature is updated to match the new interface.
		ExecuteFunc: func(_ context.Context, model, input string) (string, error) {
			require.Equal(t, selectedModel, model)
			require.Equal(t, craftedPrompt, input)
			return finalAnswer, nil
		},
	}

	server, err := NewServer(Config{Generator: mockGen, Version: "test"})
	require.NoError(t, err)

	// The form data now includes the model.
	form := strings.NewReader("prompt=" + craftedPrompt + "&model=" + selectedModel)
	req := httptest.NewRequest(http.MethodPost, "/execute", form)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	w := httptest.NewRecorder()

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), finalAnswer)
	require.NotContains(t, w.Body.String(), `hx-post="/execute"`)
}

func TestHandleUpdateFooter(t *testing.T) {
	// FIX: Remove the 'v' from the test constant.
	const (
		testVersion   = "0.5.1"
		selectedModel = "gemini-2.5-pro"
	)

	mockGen := &mockPromptGenerator{}
	server, err := NewServer(Config{Generator: mockGen, Version: testVersion})
	require.NoError(t, err)

	form := strings.NewReader("model=" + selectedModel)
	req := httptest.NewRequest(http.MethodPost, "/update-footer", form)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	w := httptest.NewRecorder()

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// Update the expected fragment to exactly match the template's output.
	expectedFragment := fmt.Sprintf("<p>Prompt Maker v%s | Model: %s</p>", testVersion, selectedModel)
	require.Contains(t, w.Body.String(), expectedFragment)
}

func TestHandleIndex_WithLoadingIndicator(t *testing.T) {
	mockGen := &mockPromptGenerator{}
	server, err := NewServer(Config{Generator: mockGen, Version: "test"})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// 1. Assert that the form points to an indicator.
	require.Contains(t, w.Body.String(), `hx-indicator="#prompt-indicator"`)
	// 2. Assert that the indicator element exists. This will fail.
	require.Contains(t, w.Body.String(), `id="prompt-indicator"`)
}
