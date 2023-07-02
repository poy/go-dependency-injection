package injection

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type builderList []func(context.Context, map[reflect.Type]*instance)

func (bs builderList) build(ctx context.Context, m map[reflect.Type]*instance) {
	// We want to do this in reverse order so that we can override types.
	for i := len(bs) - 1; i >= 0; i-- {
		bs[i](ctx, m)
	}
}

type registrar interface {
	register(context.Context, map[reflect.Type]*instance, reflect.Type, *instance)
}

var builders = map[reflect.Type]builderList{}

// Register registers a type with the injection system.
func Register[T any](f func(context.Context) T) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	builders[t] = append(builders[t], func(ctx context.Context, m map[reflect.Type]*instance) {
		inst := newInstance(ctx, f)

		// Check for special registrations. These are types that have their
		// own registration logic (e.g., Group).
		var defaultVal T
		if r, ok := any(defaultVal).(registrar); ok {
			r.register(ctx, m, t, inst)
			return
		}

		// Normal registration.
		if _, ok := m[t]; ok {
			return
		}

		m[t] = inst
	})
}

// WithOverride returns a context with the given type overridden with the
// given value.
func WithOverride[T any](ctx context.Context, f func(context.Context) T) context.Context {
	x := ctx.Value(initResolveKey{})
	if x == nil {
		panic("WithOverride can only be called within the initialization context")
	}

	// We want to create a branch, so we have to create a new map so that we
	// don't manipulate the original.
	newM := map[reflect.Type]*instance{}
	ctx = context.WithValue(ctx, initResolveKey{}, initResolve{
		m:    newM,
		objs: map[any]any{},
	})

	t := reflect.TypeOf((*T)(nil)).Elem()
	instance := newInstance(ctx, f)
	return context.WithValue(ctx, t, instance)
}

type initResolveKey struct{}

type initResolve struct {
	m    map[reflect.Type]*instance
	objs map[any]any
}

type instance struct {
	once sync.Once
	ctx  context.Context
	f    func(context.Context) any
	val  any
}

func newInstance[T any](ctx context.Context, f func(context.Context) T) *instance {
	return &instance{
		ctx: ctx,
		f:   func(ctx context.Context) any { return f(ctx) },
	}
}

func (inst *instance) Val() any {
	inst.once.Do(func() {
		inst.val = inst.f(inst.ctx)
	})
	return inst.val
}

// WithInjections returns a context with the given injections.
func WithInjection(ctx context.Context) context.Context {
	m := map[reflect.Type]*instance{}
	initCtx := context.WithValue(ctx, initResolveKey{}, initResolve{
		m:    m,
		objs: map[any]any{},
	})

	for _, bs := range builders {
		bs.build(initCtx, m)
	}

	for t, val := range m {
		ctx = context.WithValue(ctx, t, val)
	}

	return ctx
}

// TryResolve returns the instance of the given type from the context if it
// exists. It returns the default value and false otherwise.
func TryResolve[T any](ctx context.Context) (T, bool) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	val, ok := ctx.Value(t).(*instance)
	if !ok {
		if x := ctx.Value(initResolveKey{}); x != nil {
			// This is within the initialization context. Therefore, we should
			// try to resolve the value.
			// NOTE: This is necessary because we can't guarantee the order
			// things are added to the map and therefore a type might rely on
			// a type that hasn't been resolved yet.
			m := x.(initResolve).m
			if instance, ok := m[t]; ok {
				// We already know about this type, simply return it.
				return instance.Val().(T), true
			} else if bs, ok := builders[t]; ok {
				// We haven't built this one yet, build and return it.
				bs.build(ctx, m)
				return m[t].Val().(T), true
			} else {
				// We don't know about this type.
				var defaultVal T
				return defaultVal, false
			}
		} else {
			// Normal resolution context, meaning we don't know how to create
			// this type.
			var defaultVal T
			return defaultVal, false
		}
	}
	return val.Val().(T), true
}

// Resolve returns the instance of the given type from the context.
func Resolve[T any](ctx context.Context) T {
	if val, ok := TryResolve[T](ctx); ok {
		return val
	}
	if useUnsafeDefaulting(ctx) {
		var empty T
		return empty
	}
	t := reflect.TypeOf((*T)(nil)).Elem()
	panic(fmt.Sprintf("unable to resolve type: %s", t.String()))
}

// WithUnsafeDefaulting returns a context that will tell the injection system
// to use the default value for a type if it can't be resolved intead of
// panicking. This is not recommended for production use as systems will
// likely NOT handle the default value correctly.
func WithUnsafeDefaulting(ctx context.Context) context.Context {
	return context.WithValue(ctx, unsafeDefaultingKey{}, true)
}

type unsafeDefaultingKey struct{}

func useUnsafeDefaulting(ctx context.Context) bool {
	v, _ := ctx.Value(unsafeDefaultingKey{}).(bool)
	return v
}
