package web

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server holds our testable interface.
type Server struct {
	e         *echo.Echo
	generator PromptGenerator
	version   string
}

// Config now includes the generator.
type Config struct {
	Generator PromptGenerator
	Version   string
}

func NewServer(cfg Config) (*Server, error) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	s := &Server{
		e:         e,
		generator: cfg.Generator,
		version:   cfg.Version,
	}
	s.registerRoutes()

	return s, nil
}

func (s *Server) registerRoutes() {
	s.e.GET("/", s.handleIndex)
	s.e.POST("/prompt", s.handlePrompt)
}

func (s *Server) Start(addr string) error {
	return s.e.Start(addr)
}

func (s *Server) handleIndex(c echo.Context) error {
	return render(c, indexPage(s.version))
}

func (s *Server) handlePrompt(c echo.Context) error {
	userInput := c.FormValue("prompt")
	if userInput == "" {
		// Use the standard http constant for Bad Request.
		return echo.NewHTTPError(http.StatusBadRequest, "Prompt cannot be empty")
	}

	craftedPrompt, err := s.generator.Generate(c.Request().Context(), userInput)
	if err != nil {
		// Use the standard http constant for Internal Server Error.
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to generate prompt: %v", err))
	}

	return render(c, craftedPromptComponent(craftedPrompt))
}

func render(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
