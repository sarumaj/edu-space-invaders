//go:build js && wasm

package handler

import (
	"sync"
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
						config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.PerformanceDropped, config.Template{
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
							config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.PerformanceImproved, config.Template{
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
	h.once.Do(func() {
		config.RenderFunc = h.render

		if config.IsTouchDevice() {
			globalTouchEvent := &touchEvent{mutex: &sync.Mutex{}}
			js.Global().Set("touchstart", globalTouchEvent.touchStart())
			js.Global().Set("touchmove", globalTouchEvent.touchMove(h.touchEvent))
			js.Global().Set("touchend", globalTouchEvent.touchEnd(h.touchEvent))
			config.AddEventListenerToCanvas("touchstart", js.Global().Get("touchstart"))
			config.AddEventListenerToCanvas("touchmove", js.Global().Get("touchmove"))
			config.AddEventListenerToCanvas("touchend", js.Global().Get("touchend"))

		} else {
			globalKeyMap := registeredKeys{
				ArrowDown:  true,
				ArrowLeft:  true,
				ArrowRight: true,
				ArrowUp:    true,
				Pause:      true,
				Space:      true,
			}
			js.Global().Set("keydown", globalKeyMap.keyDown(h.keyEvent))
			js.Global().Set("keyup", globalKeyMap.keyUp(h.keyEvent))
			config.AddEventListener("keydown", js.Global().Get("keydown"))
			config.AddEventListener("keyup", js.Global().Get("keyup"))

			globalMouseEvent := &mouseEvent{mutex: &sync.Mutex{}}
			js.Global().Set("mousedown", globalMouseEvent.mouseDown())
			js.Global().Set("mousemove", globalMouseEvent.mouseMove(h.mouseEvent))
			js.Global().Set("mouseup", globalMouseEvent.mouseUp(h.mouseEvent))
			config.AddEventListenerToCanvas("contextmenu", js.Global().Get("mousedown"))
			config.AddEventListenerToCanvas("mousedown", js.Global().Get("mousedown"))
			config.AddEventListenerToCanvas("mousemove", js.Global().Get("mousemove"))
			config.AddEventListenerToCanvas("mouseup", js.Global().Get("mouseup"))
		}
	})
}

// registeredKeys represents a map of registered keys which are meant to be listened to.
type registeredKeys map[keyBinding]bool

// keyDown is a method that listens to the keydown event.
func (known registeredKeys) keyDown(rcv chan<- keyEvent) js.Func {
	return js.FuncOf(func(_ js.Value, p []js.Value) any {
		key := keyBinding(p[0].Get("code").String())
		if !known[key] {
			return nil
		}

		p[0].Call("preventDefault")
		rcv <- keyEvent{
			Key:     key,
			Pressed: true,
		}

		return nil
	})
}

// keyUp is a method that listens to the keyup event.
func (known registeredKeys) keyUp(rcv chan<- keyEvent) js.Func {
	return js.FuncOf(func(_ js.Value, p []js.Value) any {
		key := keyBinding(p[0].Get("code").String())
		if !known[key] {
			return nil
		}

		p[0].Call("preventDefault")
		rcv <- keyEvent{
			Key:     key,
			Pressed: false,
		}

		return nil
	})
}

// mouseDown is a method that listens to the mousedown or contextmenu event.
func (event *mouseEvent) mouseDown() js.Func {
	return js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("preventDefault")
		canvasDimensions := config.CanvasBoundingBox()
		_ = event.
			Reset().
			SetStartPosition(objects.Position{
				X: objects.Number(p[0].Get("clientX").Float() - canvasDimensions.Left),
				Y: objects.Number(p[0].Get("clientY").Float() - canvasDimensions.Top),
			}).
			SetButton(mouseButton(p[0].Get("button").Int())).
			SetStartTime(time.Now()).
			SetType(MouseEventTypeDown).
			SetPressed(true)

		return nil
	})
}

// mouseMove is a method that listens to the mousemove event.
func (event *mouseEvent) mouseMove(rcv chan<- mouseEvent) js.Func {
	return js.FuncOf(func(_ js.Value, p []js.Value) any {
		if !event.Pressed {
			return nil
		}

		p[0].Call("preventDefault")
		canvasDimensions := config.CanvasBoundingBox()
		event.
			SetCurrentPosition(objects.Position{
				X: objects.Number(p[0].Get("clientX").Float() - canvasDimensions.Left),
				Y: objects.Number(p[0].Get("clientY").Float() - canvasDimensions.Top),
			}).
			SetButton(mouseButton(p[0].Get("button").Int())).
			SetType(MouseEventTypeMove).
			Send(rcv)

		return nil
	})
}

// mouseUp is a method that listens to the mouseup event.
func (event *mouseEvent) mouseUp(rcv chan<- mouseEvent) js.Func {
	return js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("preventDefault")
		canvasDimensions := config.CanvasBoundingBox()
		event.
			SetEndPosition(objects.Position{
				X: objects.Number(p[0].Get("clientX").Float() - canvasDimensions.Left),
				Y: objects.Number(p[0].Get("clientY").Float() - canvasDimensions.Top),
			}).
			SetButton(mouseButton(p[0].Get("button").Int())).
			SetEndTime(time.Now()).
			SetPressed(false).
			SetType(MouseEventTypeUp).
			Send(rcv)

		return nil
	})
}

// touchEnd is a method that listens to the touchend event.
func (event *touchEvent) touchEnd(rcv chan<- touchEvent) js.Func {
	return js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("preventDefault")
		changedTouches := p[0].Get("changedTouches")
		canvasDimensions := config.CanvasBoundingBox()
		event.
			SetEndPosition(objects.Position{
				X: objects.Number(changedTouches.Index(0).Get("clientX").Float() - canvasDimensions.Left),
				Y: objects.Number(changedTouches.Index(0).Get("clientY").Float() - canvasDimensions.Top),
			}).
			SetEndTime(time.Now()).
			SetMultiTap(changedTouches.Length() > 1).
			SetType(TouchTypeEnd).
			Send(rcv)

		return nil
	})
}

// touchMove is a method that listens to the touchmove event.
func (event *touchEvent) touchMove(rcv chan<- touchEvent) js.Func {
	return js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("preventDefault")
		changedTouches := p[0].Get("changedTouches")
		canvasDimensions := config.CanvasBoundingBox()
		event.
			SetCurrentPosition(objects.Position{
				X: objects.Number(changedTouches.Index(0).Get("clientX").Float() - canvasDimensions.Left),
				Y: objects.Number(changedTouches.Index(0).Get("clientY").Float() - canvasDimensions.Top),
			}).
			SetMultiTap(changedTouches.Length() > 1).
			SetType(TouchTypeMove).
			Send(rcv)

		return nil
	})
}

// touchStart is a method that listens to the touchstart event.
func (event *touchEvent) touchStart() js.Func {
	return js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("preventDefault")
		changedTouches := p[0].Get("changedTouches")
		canvasDimensions := config.CanvasBoundingBox()
		_ = event.
			Reset().
			SetStartPosition(objects.Position{
				X: objects.Number(changedTouches.Index(0).Get("clientX").Float() - canvasDimensions.Left),
				Y: objects.Number(changedTouches.Index(0).Get("clientY").Float() - canvasDimensions.Top),
			}).
			SetStartTime(time.Now()).
			SetMultiTap(changedTouches.Length() > 1).
			SetType(TouchTypeStart)

		return nil
	})
}
