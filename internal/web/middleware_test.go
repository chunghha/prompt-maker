package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a handler that always returns an error.
	handler := func(_ echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "Something went wrong")
	}

	// Create the middleware.
	mw := ErrorMiddleware(handler)

	// Execute the middleware.
	err := mw(c)

	// Check if the error is handled correctly.
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Something went wrong")
}
