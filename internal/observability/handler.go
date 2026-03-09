package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// traceHandler wraps a slog.Handler to enrich log records with
// OpenTelemetry trace_id and span_id extracted from the context.
type traceHandler struct {
	inner slog.Handler
}

// NewTraceHandler returns a slog.Handler that adds trace_id and span_id
// attributes to every log record when a valid span is present in the context.
func NewTraceHandler(base slog.Handler) slog.Handler {
	return &traceHandler{inner: base}
}

// Enabled reports whether the inner handler handles records at the given level.
func (h *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle extracts the OpenTelemetry span context and, if valid, attaches
// trace_id and span_id as structured log attributes before delegating
// to the inner handler.
//
//nolint:gocritic // slog.Handler interface requires slog.Record by value.
func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}

	return h.inner.Handle(ctx, r)
}

// WithAttrs returns a new handler whose inner handler has the given attributes.
func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{inner: h.inner.WithAttrs(attrs)}
}

// WithGroup returns a new handler whose inner handler has the given group.
func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{inner: h.inner.WithGroup(name)}
}
