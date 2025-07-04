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

// mockPromptGenerator is a test double for our generator interface.
type mockPromptGenerator struct {
	GenerateFunc func(ctx context.Context, userInput string) (string, error)
}

func (m *mockPromptGenerator) Generate(ctx context.Context, userInput string) (string, error) {
	return m.GenerateFunc(ctx, userInput)
}

func TestHandleIndex(t *testing.T) {
	// Setup a mock generator (it's not used by this handler, but the server requires it).
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

	// 1. Setup the mock to return a predictable response.
	mockGen := &mockPromptGenerator{
		// Use the blank identifier "_" for the unused context parameter.
		GenerateFunc: func(_ context.Context, input string) (string, error) {
			require.Equal(t, userInput, input) // Assert the correct input was received.
			return expectedResponse, nil
		},
	}

	server, err := NewServer(Config{Generator: mockGen, Version: "test"})
	require.NoError(t, err)

	// 2. Create the request.
	form := strings.NewReader("prompt=" + userInput)
	req := httptest.NewRequest(http.MethodPost, "/prompt", form)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	w := httptest.NewRecorder()

	// 3. Execute the request.
	server.e.ServeHTTP(w, req)

	// 4. Assert the outcome.
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), expectedResponse)
	require.Contains(t, w.Body.String(), `hx-post="/execute"`)
}
