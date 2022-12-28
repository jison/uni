package core

import (
	"context"
	"github.com/jison/uni/core/model"
	"github.com/jison/uni/internal/errors"
)

func FuncOf(c Container, function interface{}, opts ...model.FuncConsumerOption) (interface{}, error) {
	if c == nil {
		return nil, errors.Newf("container is nil")
	}

	opts = append(opts, model.UpdateCallLocation())
	exe := c.FuncOf(function, opts...)
	return exe.Execute()
}

func StructOf(c Container, t model.TypeVal, opts ...model.StructConsumerOption) (interface{}, error) {
	if c == nil {
		return nil, errors.Newf("container is nil")
	}

	opts = append(opts, model.UpdateCallLocation())
	exe := c.StructOf(t, opts...)
	return exe.Execute()
}

func ValueOf(c Container, t model.TypeVal, opts ...model.ValueConsumerOption) (interface{}, error) {
	if c == nil {
		return nil, errors.Newf("container is nil")
	}

	opts = append(opts, model.UpdateCallLocation())
	exe := c.ValueOf(t, opts...)
	return exe.Execute()
}

func EnterScope(c Container, scope model.Scope) (Container, error) {
	if c == nil {
		return nil, errors.Newf("container is nil")
	}

	return c.EnterScope(scope)
}

func LeaveScope(c Container) Container {
	if c == nil {
		return nil
	}

	return c.LeaveScope()
}

var containerInContextKey = model.NewSymbol("container-in-context")

func ContainerOfCtx(ctx context.Context) Container {
	val := ctx.Value(containerInContextKey)
	c, ok := val.(Container)
	if ok {
		return c
	} else {
		return nil
	}
}

func WithContainerCtx(ctx context.Context, container Container) context.Context {
	return context.WithValue(ctx, containerInContextKey, container)
}

func FuncOfCtx(ctx context.Context, function interface{}, opts ...model.FuncConsumerOption) (interface{}, error) {
	c := ContainerOfCtx(ctx)
	opts = append(opts, model.UpdateCallLocation())
	return FuncOf(c, function, opts...)
}

func StructOfCtx(ctx context.Context, t model.TypeVal, opts ...model.StructConsumerOption) (interface{}, error) {
	c := ContainerOfCtx(ctx)
	opts = append(opts, model.UpdateCallLocation())
	return StructOf(c, t, opts...)
}

func ValueOfCtx(ctx context.Context, t model.TypeVal, opts ...model.ValueConsumerOption) (interface{}, error) {
	c := ContainerOfCtx(ctx)
	opts = append(opts, model.UpdateCallLocation())
	return ValueOf(c, t, opts...)
}

func EnterScopeCtx(ctx context.Context, scope model.Scope) (context.Context, error) {
	c := ContainerOfCtx(ctx)
	c2, err := EnterScope(c, scope)
	if err != nil {
		return nil, err
	}
	return WithContainerCtx(ctx, c2), nil
}

func LeaveScopeCtx(ctx context.Context) context.Context {
	c := ContainerOfCtx(ctx)
	if c == nil {
		return ctx
	}
	c2 := LeaveScope(c)
	return WithContainerCtx(ctx, c2)
}
