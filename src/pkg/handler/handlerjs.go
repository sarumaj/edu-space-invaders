//go:build js && wasm

package handler

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// monitor is a method that watches the FPS rate of the game.
func (h *handler) monitor() {
	frameCount := 0
	lastFrameTime := time.Now()

	var watchdog func(js.Value, []js.Value) any
	watchdog = func(_ js.Value, _ []js.Value) any {
		frameCount++
		now := time.Now()

		if elapsed := now.Sub(lastFrameTime).Seconds(); elapsed >= 1 {
			if fps := float64(frameCount) / elapsed; fps <= float64(config.Config.Control.CriticalFramesPerSecondRate) {
				h.sendMessage(fmt.Sprintf("Low FPS detected: %.2f! Reducing resource consumption.", fps))
				h.stars = nil
			}
			frameCount, lastFrameTime = 0, now
		}

		// Schedule the next frame
		js.Global().Call("requestAnimationFrame", js.FuncOf(watchdog))
		return nil
	}

	// Schedule the first frame
	js.Global().Call("requestAnimationFrame", js.FuncOf(watchdog))
}

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
			// Prevent rapid movement of the spaceship
			if time.Since(lastFired) <= config.Config.Control.SwipeCooldown {
				return nil
			}

			x := p[0].Get("changedTouches").Index(0).Get("clientX").Float()
			y := p[0].Get("changedTouches").Index(0).Get("clientY").Float()
			globalEvent.CalculateDelta(x, y)

			// Prevent only horizontal swipes from being handled in default manner
			if globalEvent.Delta.Y.Abs() < globalEvent.Delta.X.Abs() {
				p[0].Call("preventDefault")
			}

			lastFired = time.Now()
			h.touchEvent <- globalEvent
			return nil
		}))

		config.AddEventListener("touchend", js.FuncOf(func(_ js.Value, p []js.Value) any {
			x := p[0].Get("changedTouches").Index(0).Get("clientX").Float()
			y := p[0].Get("changedTouches").Index(0).Get("clientY").Float()
			globalEvent.CalculateDelta(x, y)

			// Prevent only horizontal swipes from being handled in default manner
			if globalEvent.Delta.Y.Abs() < globalEvent.Delta.X.Abs() {
				p[0].Call("preventDefault")
			}

			h.touchEvent <- globalEvent
			return nil
		}))
	})
}
