package web

import (
	"fmt"
	"net/http"
	"prompt-maker/internal/config"
	"prompt-maker/internal/gemini"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server holds our testable interface and config values.
type Server struct {
	e         *echo.Echo
	generator PromptGenerator
	version   string
}

// Config holds the dependencies for the server.
type Config struct {
	Generator PromptGenerator
	Version   string
}

func NewServer(cfg Config) (*Server, error) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/static", "static")

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
	s.e.POST("/execute", s.handleExecute)
	s.e.POST("/update-footer", s.handleUpdateFooter)
}

func (s *Server) Start(addr string) error {
	return s.e.Start(addr)
}

func (s *Server) handleIndex(c echo.Context) error {
	opts := gemini.GetModelOptions()

	modelNames := make([]string, len(opts))
	for i, opt := range opts {
		modelNames[i] = opt.Name
	}

	return render(c, indexPage(s.version, config.DefaultModel, modelNames))
}

func (s *Server) handlePrompt(c echo.Context) error {
	userInput := c.FormValue("prompt")

	modelName := c.FormValue("model") // Get model from form
	if userInput == "" || modelName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Prompt and model cannot be empty")
	}

	// Pass the modelName to the Generate method.
	craftedPrompt, err := s.generator.Generate(c.Request().Context(), modelName, userInput)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to generate prompt: %v", err))
	}

	// Pass the modelName to the template component.
	return render(c, craftedPromptComponent(craftedPrompt, modelName))
}

func (s *Server) handleExecute(c echo.Context) error {
	craftedPrompt := c.FormValue("prompt")

	modelName := c.FormValue("model") // Get model from form
	if craftedPrompt == "" || modelName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Prompt and model cannot be empty")
	}

	// Pass the modelName to the Execute method.
	finalAnswer, err := s.generator.Execute(c.Request().Context(), modelName, craftedPrompt)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to execute prompt: %v", err))
	}

	return render(c, finalAnswerComponent(finalAnswer))
}

func (s *Server) handleUpdateFooter(c echo.Context) error {
	modelName := c.FormValue("model")
	if modelName == "" {
		// Fallback to the default if the model is somehow empty.
		modelName = config.DefaultModel
	}

	return render(c, footerComponent(s.version, modelName))
}

func render(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
