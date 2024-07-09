//go:build js && wasm

package handler

import (
	"strings"
	"syscall/js"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// monitor is a method that watches the FPS rate of the game.
func (h *handler) monitor() {
	if !running.Get(h.ctx) {
		return
	}

	frameCount := 0
	lastFrameTime := time.Now()

	var watchdog func(js.Value, []js.Value) any
	watchdog = func(_ js.Value, _ []js.Value) any {
		frameCount++
		now := time.Now()

		precision := 1.0 // every second
		if config.Config.Control.CriticalFramesPerSecondRate > 10 {
			precision = 0.1 // every 100ms
		}

		if elapsed := now.Sub(lastFrameTime).Seconds(); elapsed >= precision {
			if fps := float64(frameCount) / elapsed; fps <= config.Config.Control.CriticalFramesPerSecondRate {
				h.sendMessage(config.Execute(config.Sprintf(`Low FPS detected: {{ "%.2f" | color "red" }}!`, fps)))

				// Stop all audio sources except the theme
				go config.StopAudioSources(func(name string) bool {
					return !strings.HasPrefix(name, "theme_")
				})
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
	var keydown js.Func
	var keyup js.Func
	var touchstart js.Func
	var touchmove js.Func
	var touchend js.Func

	h.once.Do(func() {
		registeredKeys := map[keyBinding]bool{
			ArrowDown:  true,
			ArrowLeft:  true,
			ArrowRight: true,
			ArrowUp:    true,
			Pause:      true,
			Space:      true,
		}
		keydown = js.FuncOf(func(_ js.Value, p []js.Value) any {
			key := keyBinding(p[0].Get("code").String())
			if !registeredKeys[key] {
				return nil
			}
			p[0].Call("preventDefault")
			h.keydownEvent <- key
			return nil
		})
		keyup = js.FuncOf(func(_ js.Value, p []js.Value) any {
			key := keyBinding(p[0].Get("code").String())
			if !registeredKeys[key] {
				return nil
			}
			p[0].Call("preventDefault")
			h.keyupEvent <- key
			return nil
		})

		config.AddEventListener("keydown", keydown)
		config.AddEventListener("keyup", keyup)

		var globalEvent touchEvent
		touchstart = js.FuncOf(func(_ js.Value, p []js.Value) any {
			p[0].Call("preventDefault")
			globalEvent.Position.X = objects.Number(p[0].Get("changedTouches").Index(0).Get("clientX").Float())
			globalEvent.Position.Y = objects.Number(p[0].Get("changedTouches").Index(0).Get("clientY").Float())
			globalEvent.Delta = objects.Position{} // Reset the delta to prevent accidental movement of the spaceship
			return nil
		})
		touchmove = js.FuncOf(func(_ js.Value, p []js.Value) any {
			p[0].Call("preventDefault")
			x := p[0].Get("changedTouches").Index(0).Get("clientX").Float()
			y := p[0].Get("changedTouches").Index(0).Get("clientY").Float()
			globalEvent.CalculateDelta(x, y)
			h.touchEvent <- globalEvent
			return nil
		})
		var lastTap time.Time
		touchend = js.FuncOf(func(_ js.Value, p []js.Value) any {
			p[0].Call("preventDefault")
			x := p[0].Get("changedTouches").Index(0).Get("clientX").Float()
			y := p[0].Get("changedTouches").Index(0).Get("clientY").Float()
			globalEvent.CalculateDelta(x, y)
			globalEvent.IsDoubleTap, lastTap = time.Since(lastTap) < config.Config.Control.DoubleTapDuration, time.Now()
			h.touchEvent <- globalEvent
			return nil
		})

		config.AddEventListener("touchstart", touchstart)
		config.AddEventListener("touchmove", touchmove)
		config.AddEventListener("touchend", touchend)
	})
}
