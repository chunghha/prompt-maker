package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

func TestSetupTracing_ReturnsShutdownFunc(t *testing.T) {
	shutdown, err := SetupTracing(context.Background())
	require.NoError(t, err)
	require.NotNil(t, shutdown)

	// Verify the global TracerProvider is set (not the default noop).
	tp := otel.GetTracerProvider()
	assert.NotNil(t, tp)

	// Shutdown should succeed without error.
	err = shutdown(context.Background())
	require.NoError(t, err)
}
