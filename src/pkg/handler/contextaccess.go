package handler

import (
	"context"
	"sync"
)

// contextMutex is a global mutex that protects the context from concurrent access.
var contextMutex = sync.RWMutex{}

const (
	isFirstTime contextAccess[bool] = "isFirstTime" // isFirstTime is a context key that indicates whether the game is played for the first time.
	offline     contextAccess[bool] = "offline"     // offline is a context key that indicates whether the game is offline.
	paused      contextAccess[bool] = "paused"      // paused is a context key that indicates whether the game is paused.
	running     contextAccess[bool] = "running"     // running is a context key that indicates whether the game is running.
	suspended   contextAccess[bool] = "suspended"   // suspended is a context key that indicates whether the game is suspended.
)

// contextAccess is a type that allows access to a context value.
type contextAccess[T any] string

// Get returns the value of the context key.
func (k contextAccess[T]) Get(ctx context.Context) T {
	contextMutex.RLock()
	v, _ := ctx.Value(k).(T)
	contextMutex.RUnlock()
	return v
}

// Set sets the value of the context key.
func (k contextAccess[T]) Set(ctx *context.Context, v T) {
	contextMutex.Lock()
	*ctx = context.WithValue(*ctx, k, v)
	contextMutex.Unlock()
}
