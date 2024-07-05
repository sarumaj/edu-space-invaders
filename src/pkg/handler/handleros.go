//go:build !js || !wasm

package handler

// render is a method that renders the game.
func (*handler) render() {}

// registerKeydownEvent is a method that registers the keydown event.
func (h *handler) registerKeydownEvent() {
	h.once.Do(func() {})
}

// SendMessage is a method that sends a message.
func (*handler) SendMessage(string) {}
