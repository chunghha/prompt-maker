package web

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

func ErrorMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err == nil {
			return nil
		}

		c.Logger().Error(err)

		var he *echo.HTTPError
		if !errors.As(err, &he) {
			he = &echo.HTTPError{
				Code:    http.StatusInternalServerError,
				Message: "Internal Server Error",
			}
		}

		if err := c.String(he.Code, he.Message.(string)); err != nil {
			c.Logger().Error(err)
		}

		return nil
	}
}
