//go:build !js || !wasm

package handler

// ask is a method that asks the user for input.
func (h *handler) ask() {}

// monitor is a method that watches the FPS rate of the game.
func (*handler) monitor() {}

// registerEventHandlers is a method that registers the event listeners.
func (h *handler) registerEventHandlers() func() {
	h.once.Do(func() {})
	return func() {}
}
