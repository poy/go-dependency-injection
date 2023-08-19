package testing

import (
	"context"
	"testing"

	"github.com/poy/go-dependency-injection/pkg/injection"
)

// WithTesting returns a context that is used for testing.
func WithTesting(t *testing.T) context.Context {
	return WithTestingContext(context.Background(), t)
}

// WithTestingContext returns a context that is used for testing.
func WithTestingContext(ctx context.Context, t *testing.T) context.Context {
	ctx = context.WithValue(ctx, testingKey{}, t)
	return injection.WithInjection(ctx)
}

// T returns the saved testing.T from the context.
func T(ctx context.Context) *testing.T {
	return ctx.Value(testingKey{}).(*testing.T)
}

type testingKey struct{}
