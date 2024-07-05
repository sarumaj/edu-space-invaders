//go:build js && wasm

package handler

import (
	"syscall/js"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// render is a method that renders the game.
// It draws the spaceship, bullets and enemies on the canvas.
// The spaceship is drawn in white color.
// The bullets are drawn in yellow color.
// The enemies are drawn in gray color.
// The goodie enemies are drawn in green color.
// The berserker enemies are drawn in red color.
// The annihilator enemies are drawn in dark red color.
// The spaceship is drawn in dark red color if it is damaged.
// The spaceship is drawn in yellow color if it is boosted.
// The spaceship is drawn in white color if it is normal.
// If draws objects as rectangles.
func (h *handler) render() {
	config.Ctx.Call("clearRect", 0, 0, config.CanvasWidth, config.CanvasHeight)

	// Draw spaceship
	switch h.spaceship.State {
	case objects.Damaged:
		config.Ctx.Set("fillStyle", "darkred")

	case objects.Boosted:
		config.Ctx.Set("fillStyle", "yellow")

	default:
		config.Ctx.Set("fillStyle", "white")
	}
	config.Ctx.Call("fillRect", h.spaceship.Position.X, h.spaceship.Position.Y, h.spaceship.Size.Width, h.spaceship.Size.Height)

	// Draw bullets
	config.Ctx.Set("fillStyle", "yellow")
	for _, b := range h.spaceship.Bullets {
		config.Ctx.Call("fillRect", b.Position.X, b.Position.Y, b.Size.Width, b.Size.Height)
	}

	// Draw enemies
	config.Ctx.Set("fillStyle", "red")
	for _, e := range h.enemies {
		switch e.Type {
		case objects.Goodie:
			config.Ctx.Set("fillStyle", "green")
		case objects.Normal:
			config.Ctx.Set("fillStyle", "gray")

		case objects.Berserker:
			config.Ctx.Set("fillStyle", "red")

		case objects.Annihilator:
			config.Ctx.Set("fillStyle", "darkred")

		default:
			config.Ctx.Set("fillStyle", "darkgray")

		}
		config.Ctx.Call("fillRect", e.Position.X, e.Position.Y, e.Size.Width, e.Size.Height)
	}
}

// registerKeydownEvent is a method that registers the keydown event.
func (h *handler) registerKeydownEvent() {
	h.once.Do(func() {
		config.Doc.Call("addEventListener", "keydown", js.FuncOf(func(_ js.Value, p []js.Value) any {
			key := p[0].Get("code").String()
			h.keydownEvent <- key
			return nil
		}))
	})
}

// SendMessage sends a message to the message box.
func (*handler) SendMessage(msg string) {
	config.MessageBox.Set("innerText", msg)
}
