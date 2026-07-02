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

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

// Rename the error variable to follow Go's initialism convention.
var errMockAPIFailed = errors.New("mock API failed")

// mockPromptGenerator is updated to match the new interface signatures.
type mockPromptGenerator struct {
	GenerateFunc      func(ctx context.Context, modelName, userInput string) (string, error)
	ExecuteFunc       func(ctx context.Context, modelName, userInput string) (string, error)
	GetModelNamesFunc func() []string
}

func (m *mockPromptGenerator) Generate(ctx context.Context, modelName, userInput string) (string, error) {
	return m.GenerateFunc(ctx, modelName, userInput)
}

func (m *mockPromptGenerator) Execute(ctx context.Context, modelName, userInput string) (string, error) {
	return m.ExecuteFunc(ctx, modelName, userInput)
}

func (m *mockPromptGenerator) GetModelNames() []string {
	return m.GetModelNamesFunc()
}

// newTestServer creates a Server with the given mock generator and version for testing.
func newTestServer(t *testing.T, gen *mockPromptGenerator, version string) *Server {
	t.Helper()

	server, err := NewServer(Config{Generator: gen, Version: version})
	require.NoError(t, err)

	return server
}

// doGETIndex performs a GET request to "/" on the server and returns the recorder.
func doGETIndex(server *Server) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
	server.e.ServeHTTP(w, req)

	return w
}

func TestHandleIndex(t *testing.T) {
	mockGen := &mockPromptGenerator{
		GetModelNamesFunc: func() []string {
			return []string{"test-model-1", "test-model-2"}
		},
	}
	server := newTestServer(t, mockGen, "test-version")
	w := doGETIndex(server)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	require.Contains(t, body, `id="prompt-form"`)
	require.Contains(t, body, "test-model-1")
	require.Contains(t, body, "test-model-2")
}

func TestHandlePrompt(t *testing.T) {
	const (
		userInput     = "make it better"
		rawResponse   = "This is the **crafted** prompt."
		selectedModel = "gemini-2.5-flash"
		// The goldmark library wraps the output in <p> tags by default.
		expectedHTMLPart = "<p>This is the <strong>crafted</strong> prompt.</p>"
	)

	mockGen := &mockPromptGenerator{
		GenerateFunc: func(_ context.Context, model, input string) (string, error) {
			require.Equal(t, selectedModel, model)
			require.Equal(t, userInput, input)

			return rawResponse, nil
		},
	}

	server := newTestServer(t, mockGen, "test")

	form := strings.NewReader("prompt=" + userInput + "&model=" + selectedModel)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/prompt", form)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	w := httptest.NewRecorder()

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	// 1. Check that the visible part is rendered correctly as HTML.
	require.Contains(t, body, expectedHTMLPart, "The rendered HTML should contain the strong tag inside a p tag.")

	// 2. Check that the raw markdown is correctly placed in the hidden div for the copy button.
	require.Contains(
		t,
		body,
		`<div id="raw-crafted-prompt" class="hidden">`+rawResponse+`</div>`,
		"The hidden div should contain the raw markdown.",
	)

	// 3. Check that the raw markdown is correctly placed in the hidden input for the resubmit form.
	require.Contains(
		t,
		body,
		`<input type="hidden" name="prompt" value="`+rawResponse+`">`,
		"The hidden input should contain the raw markdown for resubmission.",
	)
}

func TestHandleIndex_WithDaisyUI(t *testing.T) {
	mockGen := &mockPromptGenerator{
		GetModelNamesFunc: func() []string {
			return []string{"test-model-1"}
		},
	}
	server := newTestServer(t, mockGen, "test-version")
	w := doGETIndex(server)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), config.DefaultModel)
}

func TestHandleExecute(t *testing.T) {
	const (
		craftedPrompt = "This is the crafted prompt."
		finalAnswer   = "This is the **final** answer from the AI."
		selectedModel = "gemini-2.5-flash"
		expectedHTML  = "<p>This is the <strong>final</strong> answer from the AI.</p>"
	)

	mockGen := &mockPromptGenerator{
		ExecuteFunc: func(_ context.Context, model, input string) (string, error) {
			require.Equal(t, selectedModel, model)
			require.Equal(t, craftedPrompt, input)

			return finalAnswer, nil
		},
	}

	server := newTestServer(t, mockGen, "test")

	form := strings.NewReader("prompt=" + craftedPrompt + "&model=" + selectedModel)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/execute", form)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	w := httptest.NewRecorder()

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	require.Contains(t, body, expectedHTML)
	require.NotContains(t, body, `hx-post="/execute"`)
}

func TestHandleUpdateFooter(t *testing.T) {
	const (
		testVersion   = "0.5.1"
		selectedModel = "gemini-2.5-pro"
	)

	mockGen := &mockPromptGenerator{}
	server := newTestServer(t, mockGen, testVersion)

	form := strings.NewReader("model=" + selectedModel)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/update-footer", form)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	w := httptest.NewRecorder()

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	expectedFragment := fmt.Sprintf("prompt-maker v%s / %s", testVersion, selectedModel)
	require.Contains(t, w.Body.String(), expectedFragment)
}

func TestHandleIndex_WithLoadingIndicator(t *testing.T) {
	mockGen := &mockPromptGenerator{
		GetModelNamesFunc: func() []string {
			return []string{"test-model-1"}
		},
	}
	server := newTestServer(t, mockGen, "test")
	w := doGETIndex(server)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `hx-indicator="#prompt-indicator"`)
	require.Contains(t, w.Body.String(), `id="prompt-indicator"`)
}

// assertAPIError verifies that posting "prompt=test&model=<model>" to the
// given path returns HTTP 500 with the expected error message.
func assertAPIError(t *testing.T, server *Server, path, expectedMessage string) {
	t.Helper()

	form := strings.NewReader("prompt=test&model=gemini-2.5-flash")
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, path, form)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	w := httptest.NewRecorder()

	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Contains(t, w.Body.String(), expectedMessage)
}

func TestHandlePrompt_ApiError(t *testing.T) {
	mockGen := &mockPromptGenerator{
		GenerateFunc: func(_ context.Context, _, _ string) (string, error) {
			return "", errMockAPIFailed
		},
	}
	server := newTestServer(t, mockGen, "test")
	assertAPIError(t, server, "/prompt", "The AI failed to generate a response. Please try again.")
}

func TestHandleIndex_WithClearButton(t *testing.T) {
	mockGen := &mockPromptGenerator{
		GetModelNamesFunc: func() []string {
			return []string{"test-model-1"}
		},
	}
	server := newTestServer(t, mockGen, "test")
	w := doGETIndex(server)

	require.Equal(t, http.StatusOK, w.Code)
	// Assert that the clear button exists with the correct HTMX attributes. This will fail.
	require.Contains(t, w.Body.String(), `hx-post="/clear"`)
	require.Contains(t, w.Body.String(), `hx-target="#response-container"`)
}

func TestHandleClear(t *testing.T) {
	mockGen := &mockPromptGenerator{}
	server := newTestServer(t, mockGen, "test")

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/clear", http.NoBody)

	// This will fail because the endpoint doesn't exist.
	server.e.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	// Assert that the response body is empty.
	require.Empty(t, w.Body.String())
}

func TestHandleExecute_ApiError(t *testing.T) {
	mockGen := &mockPromptGenerator{
		ExecuteFunc: func(_ context.Context, _, _ string) (string, error) {
			return "", errMockAPIFailed
		},
	}
	server := newTestServer(t, mockGen, "test")
	assertAPIError(t, server, "/execute", "The AI failed to execute the prompt. Please try again.")
}
