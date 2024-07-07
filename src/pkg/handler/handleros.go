//go:build !js || !wasm

package handler

// render is a method that renders the game.
func (*handler) render() {}

// registerEventHandlers is a method that registers the event listeners.
func (h *handler) registerEventHandlers() {
	h.once.Do(func() {})
}

// sendMessage is a method that sends a message.
func (*handler) sendMessage(string) {}
