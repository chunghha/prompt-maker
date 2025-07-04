package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// mockPromptGenerator now includes ExecuteFunc for the new interface method.
type mockPromptGenerator struct {
	GenerateFunc func(ctx context.Context, userInput string) (string, error)
	ExecuteFunc  func(ctx context.Context, userInput string) (string, error)
}

func (m *mockPromptGenerator) Generate(ctx context.Context, userInput string) (string, error) {
	return m.GenerateFunc(ctx, userInput)
}

func (m *mockPromptGenerator) Execute(ctx context.Context, userInput string) (string, error) {
	return m.ExecuteFunc(ctx, userInput)
}

func TestHandleIndex(t *testing.T) {
	mockGen := &mockPromptGenerator{}
	server, err := NewServer(Config{Generator: mockGen})
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
	)

	mockGen := &mockPromptGenerator{
		GenerateFunc: func(_ context.Context, input string) (string, error) {
			require.Equal(t, userInput, input)
			return expectedResponse, nil
		},
	}

	server, err := NewServer(Config{Generator: mockGen, Version: "test"})
	require.NoError(t, err)

	form := strings.NewReader("prompt=" + userInput)
	req := httptest.NewRequest(http.MethodPost, "/prompt", form)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	w := httptest.NewRecorder()

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), expectedResponse)
	require.Contains(t, w.Body.String(), `class="btn btn-secondary"`)
}

func TestHandleIndex_WithDaisyUI(t *testing.T) {
	mockGen := &mockPromptGenerator{}
	server, err := NewServer(Config{Generator: mockGen})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `<html lang="en" data-theme="light">`)
	require.Contains(t, w.Body.String(), `id="theme-switcher"`)
	require.Contains(t, w.Body.String(), `class="btn btn-primary mt-4"`)
}

func TestHandleExecute(t *testing.T) {
	const (
		craftedPrompt = "This is the crafted prompt."
		finalAnswer   = "This is the final answer from the AI."
	)

	mockGen := &mockPromptGenerator{
		ExecuteFunc: func(_ context.Context, input string) (string, error) {
			require.Equal(t, craftedPrompt, input)
			return finalAnswer, nil
		},
	}

	server, err := NewServer(Config{Generator: mockGen, Version: "test"})
	require.NoError(t, err)

	form := strings.NewReader("prompt=" + craftedPrompt)
	req := httptest.NewRequest(http.MethodPost, "/execute", form)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	w := httptest.NewRecorder()

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), finalAnswer)
	require.NotContains(t, w.Body.String(), `hx-post="/execute"`)
}
