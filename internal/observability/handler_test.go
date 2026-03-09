package observability

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestTraceHandler_WithoutSpan(t *testing.T) {
	var buf bytes.Buffer

	base := slog.NewTextHandler(&buf, nil)
	handler := NewTraceHandler(base)

	logger := slog.New(handler)
	logger.InfoContext(context.Background(), "test message")

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.NotContains(t, output, "trace_id")
	assert.NotContains(t, output, "span_id")
}

func TestTraceHandler_WithValidSpan(t *testing.T) {
	var buf bytes.Buffer

	base := slog.NewTextHandler(&buf, nil)
	handler := NewTraceHandler(base)
	logger := slog.New(handler)

	// Create a span context with known IDs.
	traceID, err := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	require.NoError(t, err)

	spanID, err := trace.SpanIDFromHex("0102030405060708")
	require.NoError(t, err)

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	ctx := trace.ContextWithSpanContext(context.Background(), sc)
	logger.InfoContext(ctx, "traced message")

	output := buf.String()
	assert.Contains(t, output, "traced message")
	assert.Contains(t, output, "trace_id=0102030405060708090a0b0c0d0e0f10")
	assert.Contains(t, output, "span_id=0102030405060708")
}

func TestTraceHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer

	base := slog.NewTextHandler(&buf, nil)
	handler := NewTraceHandler(base)

	withAttrs := handler.WithAttrs([]slog.Attr{slog.String("service", "test")})
	logger := slog.New(withAttrs)
	logger.InfoContext(context.Background(), "attrs message")

	output := buf.String()
	assert.Contains(t, output, "service=test")
	assert.Contains(t, output, "attrs message")
}

func TestTraceHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer

	base := slog.NewTextHandler(&buf, nil)
	handler := NewTraceHandler(base)

	withGroup := handler.WithGroup("mygroup")
	logger := slog.New(withGroup)
	logger.InfoContext(context.Background(), "grouped message", "key", "val")

	output := buf.String()
	assert.Contains(t, output, "grouped message")
	assert.Contains(t, output, "mygroup.key=val")
}

func TestTraceHandler_Enabled(t *testing.T) {
	base := slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelWarn})
	handler := NewTraceHandler(base)

	assert.False(t, handler.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelWarn))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelError))
}
