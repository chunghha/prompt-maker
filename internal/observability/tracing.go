package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// SetupTracing configures the global OpenTelemetry TracerProvider and
// W3C trace-context propagator. The returned shutdown function must be
// called on application exit to flush any pending spans.
//
// By default no exporter is configured; add sdktrace.WithBatcher(exporter)
// or set OTEL_EXPORTER_* environment variables to enable trace export.
func SetupTracing(_ context.Context) (shutdown func(context.Context) error, err error) {
	tp := sdktrace.NewTracerProvider()

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
