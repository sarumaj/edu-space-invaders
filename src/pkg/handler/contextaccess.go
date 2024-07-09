package handler

import (
	"context"
)

const (
	isFirstTime contextAccess[bool] = "isFirstTime"
	running     contextAccess[bool] = "running"
)

// contextAccess is a type that allows access to a context value.
type contextAccess[T any] string

// Get returns the value of the context key.
func (k contextAccess[T]) Get(ctx context.Context) T {
	v, _ := ctx.Value(k).(T)
	return v
}

// Set sets the value of the context key.
func (k contextAccess[T]) Set(ctx *context.Context, v T) {
	*ctx = context.WithValue(*ctx, k, v)
}
