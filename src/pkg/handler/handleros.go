//go:build !js || !wasm

package handler

// registerEventHandlers is a method that registers the event listeners.
func (h *handler) registerEventHandlers() {
	h.once.Do(func() {})
}
