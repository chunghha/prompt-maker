package web

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
)

// ErrorMiddleware catches errors from downstream handlers, logs them
// with structured slog output, and writes a plain-text HTTP response.
func ErrorMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		err := next(c)
		if err == nil {
			return nil
		}

		slog.ErrorContext(c.Request().Context(), "handler error", "error", err)

		he, ok := errors.AsType[*echo.HTTPError](err)
		if !ok {
			he = &echo.HTTPError{
				Code:    http.StatusInternalServerError,
				Message: "Internal Server Error",
			}
		}

		if writeErr := c.String(he.Code, he.Message); writeErr != nil {
			slog.ErrorContext(c.Request().Context(), "failed to write error response", "error", writeErr)
		}

		return nil
	}
}
