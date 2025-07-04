package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"prompt-maker/internal/config"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// Rename the error variable to follow Go's initialism convention.
var errMockAPIFailed = errors.New("mock API failed")

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
		expectedResponse = "This is the **crafted** prompt."
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
	require.Contains(t, w.Body.String(), `onclick="copyRawText(this)"`)
	require.Contains(t, w.Body.String(), `data-target-id="raw-crafted-prompt"`)
	require.Contains(t, w.Body.String(), `<div id="raw-crafted-prompt" class="hidden">This is the **crafted** prompt.</div>`)
}

func TestHandleIndex_WithDaisyUI(t *testing.T) {
	mockGen := &mockPromptGenerator{}
	server, err := NewServer(Config{Generator: mockGen, Version: "test-version"})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "Model: "+config.DefaultModel)
}

func TestHandleExecute(t *testing.T) {
	const (
		craftedPrompt = "This is the crafted prompt."
		finalAnswer   = "This is the final answer from the AI."
		selectedModel = "gemini-2.5-flash"
	)

	mockGen := &mockPromptGenerator{
		ExecuteFunc: func(_ context.Context, model, input string) (string, error) {
			require.Equal(t, selectedModel, model)
			require.Equal(t, craftedPrompt, input)
			return finalAnswer, nil
		},
	}

	server, err := NewServer(Config{Generator: mockGen, Version: "test"})
	require.NoError(t, err)

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
	require.Contains(t, w.Body.String(), `hx-indicator="#prompt-indicator"`)
	require.Contains(t, w.Body.String(), `id="prompt-indicator"`)
}

func TestHandlePrompt_ApiError(t *testing.T) {
	const (
		userFacingErrorMessage = "The AI failed to generate a response. Please try again."
		selectedModel          = "gemini-2.5-flash"
	)

	mockGen := &mockPromptGenerator{
		GenerateFunc: func(_ context.Context, _, _ string) (string, error) {
			// Use the correctly named error variable.
			return "", errMockAPIFailed
		},
	}

	server, err := NewServer(Config{Generator: mockGen, Version: "test"})
	require.NoError(t, err)

	form := strings.NewReader("prompt=test&model=" + selectedModel)
	req := httptest.NewRequest(http.MethodPost, "/prompt", form)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	w := httptest.NewRecorder()

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `class="alert alert-error"`)
	require.Contains(t, w.Body.String(), userFacingErrorMessage)
}

func TestHandleIndex_WithClearButton(t *testing.T) {
	mockGen := &mockPromptGenerator{}
	server, err := NewServer(Config{Generator: mockGen, Version: "test"})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// Assert that the clear button exists with the correct HTMX attributes. This will fail.
	require.Contains(t, w.Body.String(), `hx-post="/clear"`)
	require.Contains(t, w.Body.String(), `hx-target="#response-container"`)
}

func TestHandleClear(t *testing.T) {
	mockGen := &mockPromptGenerator{}
	server, err := NewServer(Config{Generator: mockGen, Version: "test"})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/clear", http.NoBody)

	// This will fail because the endpoint doesn't exist.
	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// Assert that the response body is empty.
	require.Empty(t, w.Body.String())
}
