//go:build js && wasm

package handler

import (
	"syscall/js"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// registerEventHandlers is a method that registers the event listeners.
func (h *handler) registerEventHandlers() {
	h.once.Do(func() {
		registeredKeys := map[keyBinding]bool{
			ArrowLeft:  true,
			ArrowRight: true,
			Space:      true,
		}
		config.AddEventListener("keydown", js.FuncOf(func(_ js.Value, p []js.Value) any {
			key := keyBinding(p[0].Get("code").String())
			if !registeredKeys[key] {
				return nil
			}
			p[0].Call("preventDefault")
			h.keydownEvent <- key
			return nil
		}))

		config.AddEventListener("keyup", js.FuncOf(func(_ js.Value, p []js.Value) any {
			key := keyBinding(p[0].Get("code").String())
			if !registeredKeys[key] {
				return nil
			}
			p[0].Call("preventDefault")
			h.keyupEvent <- key
			return nil
		}))

		var globalEvent touchEvent
		config.AddEventListener("touchstart", js.FuncOf(func(_ js.Value, p []js.Value) any {
			p[0].Call("preventDefault")
			globalEvent.Position.X = objects.Number(p[0].Get("changedTouches").Index(0).Get("clientX").Float())
			globalEvent.Position.Y = objects.Number(p[0].Get("changedTouches").Index(0).Get("clientY").Float())
			globalEvent.Delta = objects.Position{} // Reset the delta to prevent accidental movement of the spaceship
			return nil
		}))

		var lastFired time.Time
		config.AddEventListener("touchmove", js.FuncOf(func(_ js.Value, p []js.Value) any {
			p[0].Call("preventDefault")

			// Prevent rapid movement of the spaceship
			if time.Since(lastFired) <= config.Config.Control.SwipeCooldown {
				return nil
			}

			x := p[0].Get("changedTouches").Index(0).Get("clientX").Float()
			y := p[0].Get("changedTouches").Index(0).Get("clientY").Float()
			globalEvent.CalculateDelta(x, y)
			lastFired = time.Now()
			h.touchEvent <- globalEvent
			return nil
		}))

		config.AddEventListener("touchend", js.FuncOf(func(_ js.Value, p []js.Value) any {
			p[0].Call("preventDefault")
			x := p[0].Get("changedTouches").Index(0).Get("clientX").Float()
			y := p[0].Get("changedTouches").Index(0).Get("clientY").Float()
			globalEvent.CalculateDelta(x, y)
			h.touchEvent <- globalEvent
			return nil
		}))
	})
}
