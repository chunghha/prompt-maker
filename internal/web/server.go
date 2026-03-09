package web

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"prompt-maker/internal/config"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// Server holds our testable interface and config values.
type Server struct {
	e         *echo.Echo
	generator PromptGenerator
	version   string
	md        goldmark.Markdown
}

// Config holds the dependencies for the server.
type Config struct {
	Generator PromptGenerator
	Version   string
}

// NewServer creates a configured Echo server with OTEL tracing,
// structured request logging, error handling middleware, and routes.
func NewServer(cfg Config) (*Server, error) {
	e := echo.New()
	e.Use(otelecho.Middleware("prompt-maker"))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:  true,
		LogURI:     true,
		LogMethod:  true,
		LogLatency: true,
		LogError:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []slog.Attr{
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
			}
			if v.Error != nil {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
			}

			slog.LogAttrs(c.Request().Context(), slog.LevelInfo, "request", attrs...)

			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(ErrorMiddleware)
	e.Static("/static", "static")

	s := &Server{
		e:         e,
		generator: cfg.Generator,
		version:   cfg.Version,
		md: goldmark.New(
			goldmark.WithRendererOptions(
				html.WithUnsafe(), // Allow raw HTML in markdown
			),
		),
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

// Shutdown gracefully shuts down the server without interrupting active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}

func (s *Server) handleIndex(c echo.Context) error {
	// Pass the model names, themes, and default theme to the index page template.
	return render(c, indexPage(s.version, config.DefaultModel, DefaultTheme, s.generator.GetModelNames(), getThemes()))
}

func (s *Server) handlePrompt(c echo.Context) error {
	userInput := c.FormValue("prompt")

	modelName := c.FormValue("model")
	if userInput == "" || modelName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Prompt and model cannot be empty.")
	}

	// Generate the crafted prompt from the user input.
	craftedPrompt, err := s.generator.Generate(c.Request().Context(), modelName, userInput)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "The AI failed to generate a response. Please try again.")
	}

	// Convert the markdown response to HTML.
	craftedPromptHTML := s.markdownToHTML(craftedPrompt)

	return render(c, craftedPromptComponent(craftedPromptHTML, craftedPrompt, modelName))
}

func (s *Server) handleExecute(c echo.Context) error {
	craftedPrompt := c.FormValue("prompt")

	modelName := c.FormValue("model")
	if craftedPrompt == "" || modelName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Prompt and model cannot be empty.")
	}

	// Execute the crafted prompt to get the final answer.
	finalAnswer, err := s.generator.Execute(c.Request().Context(), modelName, craftedPrompt)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "The AI failed to execute the prompt. Please try again.")
	}

	// Convert the markdown response to HTML.
	finalAnswerHTML := s.markdownToHTML(finalAnswer)

	return render(c, finalAnswerComponent(finalAnswerHTML, finalAnswer))
}

func (s *Server) handleUpdateFooter(c echo.Context) error {
	modelName := c.FormValue("model")
	if modelName == "" {
		modelName = config.DefaultModel
	}

	return render(c, footerComponent(s.version, modelName))
}

func handleClear(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// markdownToHTML converts a markdown string to its HTML representation.
func (s *Server) markdownToHTML(str string) string {
	var buf bytes.Buffer
	if err := s.md.Convert([]byte(str), &buf); err != nil {
		return str // Return raw text on error
	}

	return buf.String()
}

func render(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
