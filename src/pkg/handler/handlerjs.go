//go:build js && wasm

package handler

import (
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
	suspendedFrameCount := 0
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
			switch fps := float64(frameCount) / elapsed; {
			case fps <= config.Config.Control.CriticalFramesPerSecondRate:
				// Notify the user about the performance drop
				config.SendMessage(config.Execute(config.Config.Messages.Templates.PerformanceDropped, config.Template{
					"FPS": fps,
				}))

				// Pause the game
				if running.Get(h.ctx) {
					running.Set(&h.ctx, false)
					suspended.Set(&h.ctx, true)
				}

			case fps >= (config.Config.Control.CriticalFramesPerSecondRate+config.Config.Control.DesiredFramesPerSecondRate)/2 && !running.Get(h.ctx):
				if suspendedFrameCount < config.Config.Control.SuspensionFrames {
					// Increase the suspended frame count
					suspendedFrameCount++

				} else {
					// Reset the suspended frame count
					suspendedFrameCount = 0

					// Notify the user about the performance boost
					config.SendMessage(config.Execute(config.Config.Messages.Templates.PerformanceImproved, config.Template{
						"FPS": fps,
					}))

					// Resume the game
					running.Set(&h.ctx, true)
					suspended.Set(&h.ctx, false)

				}

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
			changedTouches := p[0].Get("changedTouches")
			globalEvent.Clear()
			globalEvent.StartPosition = objects.Position{
				X: objects.Number(changedTouches.Index(0).Get("clientX").Float()),
				Y: objects.Number(changedTouches.Index(0).Get("clientY").Float()),
			}
			globalEvent.StartTime = time.Now()
			for i := 1; i < changedTouches.Length(); i++ {
				globalEvent.Correlations = append(globalEvent.Correlations, touchEvent{
					StartPosition: objects.Position{
						X: objects.Number(changedTouches.Index(i).Get("clientX").Float()),
						Y: objects.Number(changedTouches.Index(i).Get("clientY").Float()),
					},
					StartTime: globalEvent.StartTime,
				})
			}
			return nil
		})
		touchmove = js.FuncOf(func(_ js.Value, p []js.Value) any {
			p[0].Call("preventDefault")
			changedTouches := p[0].Get("changedTouches")
			globalEvent.CurrentPosition = objects.Position{
				X: objects.Number(changedTouches.Index(0).Get("clientX").Float()),
				Y: objects.Number(changedTouches.Index(0).Get("clientY").Float()),
			}
			for i := 1; i < changedTouches.Length(); i++ {
				if i >= len(globalEvent.Correlations) {
					break
				}

				globalEvent.Correlations[i-1].CurrentPosition = objects.Position{
					X: objects.Number(changedTouches.Index(i).Get("clientX").Float()),
					Y: objects.Number(changedTouches.Index(i).Get("clientY").Float()),
				}
			}
			h.touchEvent <- globalEvent
			return nil
		})
		touchend = js.FuncOf(func(_ js.Value, p []js.Value) any {
			p[0].Call("preventDefault")
			changedTouches := p[0].Get("changedTouches")
			globalEvent.EndPosition = objects.Position{
				X: objects.Number(changedTouches.Index(0).Get("clientX").Float()),
				Y: objects.Number(changedTouches.Index(0).Get("clientY").Float()),
			}
			globalEvent.EndTime = time.Now()
			for i := 1; i < changedTouches.Length(); i++ {
				if i >= len(globalEvent.Correlations) {
					break
				}

				globalEvent.Correlations[i-1].EndPosition = objects.Position{
					X: objects.Number(changedTouches.Index(i).Get("clientX").Float()),
					Y: objects.Number(changedTouches.Index(i).Get("clientY").Float()),
				}
				globalEvent.Correlations[i-1].EndTime = globalEvent.EndTime
			}
			h.touchEvent <- globalEvent
			return nil
		})

		config.AddEventListener("touchstart", touchstart)
		config.AddEventListener("touchmove", touchmove)
		config.AddEventListener("touchend", touchend)
	})
}
