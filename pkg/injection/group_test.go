package injection_test

import (
	"context"
	"testing"

	"github.com/poy/go-dependency-injection/pkg/injection"
)

func init() {
	injection.Register[injection.Group[int]](func(ctx context.Context) injection.Group[int] {
		return injection.AddToGroup[int](ctx, 1)
	})
	injection.Register[injection.Group[int]](func(ctx context.Context) injection.Group[int] {
		return injection.AddToGroup[int](ctx, 2)
	})
}

func TestGroup(t *testing.T) {
	t.Parallel()

	ctx := injection.WithInjection(context.Background())

	group := injection.Resolve[injection.Group[int]](ctx)
	if actual, expected := len(group.Vals()), 2; actual != expected {
		t.Fatalf("expected %d, got %d", expected, actual)
	}
}
