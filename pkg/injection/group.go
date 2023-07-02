package injection

import (
	"context"
	"reflect"
	"sync"
)

// AddToGroup is used within the resolution function for the type.
func AddToGroup[T any](ctx context.Context, obj T) Group[T] {
	x := ctx.Value(initResolveKey{})
	if x == nil {
		panic("AddToGroup can only be invoked within type resolution functions")
	}
	ir := x.(initResolve)

	gi, ok := ir.objs[groupKey[T]{}]
	if !ok {
		gi = Group[T]{}
		ir.objs[groupKey[T]{}] = gi
	}
	g := gi.(Group[T])
	g.vals = append(g.vals, obj)
	ir.objs[groupKey[T]{}] = g
	return g
}

// Group is used by injection to allow for multiple instances of the same type
// to be injected. This is useful for things like http.Handlers.
type Group[T any] struct {
	once      sync.Once
	vals      []T
	instances []*instance
}

// Verify that Group[T] implements registrar.
var _ registrar = Group[int]{}

type groupKey[T any] struct{}

// Vals returns the slice of values in the Group.
func (g Group[T]) Vals() []T {
	g.once.Do(func() {
		for _, inst := range g.instances {
			// NOTE: We are replacing the value each time (as opposed to
			// appending) because we aren't storing a pointer to the value and
			// therefore the group is getting updated with the last value each
			// time.
			g.vals = inst.Val().(Group[T]).vals
		}
	})
	return g.vals
}

func (g Group[T]) register(ctx context.Context, m map[reflect.Type]*instance, t reflect.Type, inst *instance) {
	x := ctx.Value(initResolveKey{})
	if x == nil {
		panic("register can only be invoked within type resolution functions")
	}
	ir := x.(initResolve)

	gi, ok := ir.objs[groupKey[T]{}]
	if !ok {
		gi = Group[T]{}
		ir.objs[groupKey[T]{}] = gi
	}
	g = gi.(Group[T])
	g.instances = append(g.instances, inst)
	ir.objs[groupKey[T]{}] = g

	m[t] = newInstance(ctx, func(ctx context.Context) Group[T] {
		return g
	})
}
