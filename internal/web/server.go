package web

import (
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
	s.e.POST("/clear", handleClear)
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

	modelName := c.FormValue("model")
	if userInput == "" || modelName == "" {
		return render(c, errorComponent("Prompt and model cannot be empty."))
	}

	craftedPrompt, err := s.generator.Generate(c.Request().Context(), modelName, userInput)
	if err != nil {
		c.Logger().Errorf("Failed to generate prompt: %v", err)
		return render(c, errorComponent("The AI failed to generate a response. Please try again."))
	}

	return render(c, craftedPromptComponent(craftedPrompt, modelName))
}

func (s *Server) handleExecute(c echo.Context) error {
	craftedPrompt := c.FormValue("prompt")

	modelName := c.FormValue("model")
	if craftedPrompt == "" || modelName == "" {
		return render(c, errorComponent("Prompt and model cannot be empty."))
	}

	finalAnswer, err := s.generator.Execute(c.Request().Context(), modelName, craftedPrompt)
	if err != nil {
		c.Logger().Errorf("Failed to execute prompt: %v", err)
		return render(c, errorComponent("The AI failed to execute the prompt. Please try again."))
	}

	return render(c, finalAnswerComponent(finalAnswer))
}

func (s *Server) handleUpdateFooter(c echo.Context) error {
	modelName := c.FormValue("model")
	if modelName == "" {
		modelName = config.DefaultModel
	}

	return render(c, footerComponent(s.version, modelName))
}

func handleClear(c echo.Context) error {
	return c.HTML(http.StatusOK, "")
}

func render(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
