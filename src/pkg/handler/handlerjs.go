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
				if running.Get(h.ctx) { // If the game is running
					// Notify the user about the performance drop
					if config.Config.Control.Debug.Get() {
						config.SendMessage(config.Execute(config.Config.Messages.Templates.PerformanceDropped, config.Template{
							"FPS": fps,
						}))
					}

					running.Set(&h.ctx, false)  // Pause the game
					suspended.Set(&h.ctx, true) // Set the suspended state
				}

			case fps >= (config.Config.Control.CriticalFramesPerSecondRate+config.Config.Control.DesiredFramesPerSecondRate)/2 && !running.Get(h.ctx):
				if suspendedFrameCount < config.Config.Control.SuspensionFrames {
					// Increase the suspended frame count
					suspendedFrameCount++

				} else {
					// Reset the suspended frame count
					suspendedFrameCount = 0

					if !paused.Get(h.ctx) { // Do not resume the game if it is paused
						if config.Config.Control.Debug.Get() {
							// Notify the user about the performance boost
							config.SendMessage(config.Execute(config.Config.Messages.Templates.PerformanceImproved, config.Template{
								"FPS": fps,
							}))
						}

						running.Set(&h.ctx, true)    // Resume the game
						suspended.Set(&h.ctx, false) // Reset the suspended state
					}

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
			globalEvent = touchEvent{
				StartPosition: objects.Position{
					X: objects.Number(changedTouches.Index(0).Get("clientX").Float()),
					Y: objects.Number(changedTouches.Index(0).Get("clientY").Float()),
				},
				StartTime:    time.Now(),
				Correlations: make([]touchEvent, changedTouches.Length()-1),
			}
			for i := 1; i < changedTouches.Length() && i < len(globalEvent.Correlations); i++ {
				globalEvent.Correlations[i-1] = touchEvent{
					StartPosition: objects.Position{
						X: objects.Number(changedTouches.Index(i).Get("clientX").Float()),
						Y: objects.Number(changedTouches.Index(i).Get("clientY").Float()),
					},
					StartTime: globalEvent.StartTime,
				}
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
			for i := 1; i < changedTouches.Length() && i < len(globalEvent.Correlations); i++ {
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
			for i := 1; i < changedTouches.Length() && i < len(globalEvent.Correlations); i++ {
				globalEvent.Correlations[i-1].EndPosition = objects.Position{
					X: objects.Number(changedTouches.Index(i).Get("clientX").Float()),
					Y: objects.Number(changedTouches.Index(i).Get("clientY").Float()),
				}
				globalEvent.Correlations[i-1].EndTime = globalEvent.EndTime
			}
			h.touchEvent <- globalEvent
			return nil
		})

		config.AddEventListenerToCanvas("touchstart", touchstart)
		config.AddEventListenerToCanvas("touchmove", touchmove)
		config.AddEventListenerToCanvas("touchend", touchend)
	})
}
