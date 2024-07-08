//go:build !js || !wasm

package handler

// monitor is a method that watches the FPS rate of the game.
func (*handler) monitor() {}

// registerEventHandlers is a method that registers the event listeners.
func (h *handler) registerEventHandlers() func() {
	h.once.Do(func() {})
	return func() {}
}
