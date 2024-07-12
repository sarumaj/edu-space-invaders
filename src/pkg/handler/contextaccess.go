package handler

import (
	"context"
)

const (
	isFirstTime contextAccess[bool] = "isFirstTime" // isFirstTime is a context key that indicates whether the game is played for the first time.
	running     contextAccess[bool] = "running"     // running is a context key that indicates whether the game is running.
	suspended   contextAccess[bool] = "suspended"   // suspended is a context key that indicates whether the game is suspended.
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
