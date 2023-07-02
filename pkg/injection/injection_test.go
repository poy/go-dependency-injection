package injection_test

import (
	"context"
	"testing"

	"github.com/poy/go-dependency-injection/pkg/injection"
)

func init() {
	injection.Register[int](func(ctx context.Context) int {
		panic("This should not be invoked because it the next registration overrides it")
	})
	injection.Register[int](func(ctx context.Context) int {
		return len(injection.Resolve[string](ctx))
	})
	injection.Register[string](func(ctx context.Context) string {
		return "hello"
	})
	injection.Register[bool](func(ctx context.Context) bool {
		panic("This should not be called because it's not used")
	})

	injection.Register[rune](func(ctx context.Context) rune {
		ctx = injection.WithOverride[string](ctx, func(ctx context.Context) string {
			return "goodbye"
		})
		s := injection.Resolve[string](ctx)
		return []rune(s)[0]
	})
}

func TestInjection(t *testing.T) {
	t.Parallel()
	ctx := injection.WithInjection(context.Background())

	if actual, expected := injection.Resolve[string](ctx), "hello"; actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
	if actual, expected := injection.Resolve[int](ctx), 5; actual != expected {
		t.Fatalf("expected %d, got %d", expected, actual)
	}
}

func TestInjection_notFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("expected a panic")
		}
	}()

	injection.Resolve[int32](ctx)
}

func TestInjection_notFound_unsafeDefault(t *testing.T) {
	t.Parallel()
	ctx := injection.WithUnsafeDefaulting(context.Background())

	val := injection.Resolve[int32](ctx)
	if actual, expected := val, int32(0); actual != expected {
		t.Fatalf("expected %d, got %d", expected, actual)
	}
}

func TestTryResolve_found(t *testing.T) {
	t.Parallel()
	ctx := injection.WithInjection(context.Background())

	v, ok := injection.TryResolve[string](ctx)
	if actual, expected := ok, true; actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
	if actual, expected := v, "hello"; actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

func TestTryResolve_notFound(t *testing.T) {
	t.Parallel()
	ctx := injection.WithInjection(context.Background())

	type notRegistered struct{}

	_, ok := injection.TryResolve[notRegistered](ctx)
	if actual, expected := ok, false; actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func TestWithOverride(t *testing.T) {
	t.Parallel()

	ctx := injection.WithInjection(context.Background())
	if actual, expected := injection.Resolve[rune](ctx), 'g'; actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
