//go:build js && wasm

package handler

import (
	"fmt"
	"sync"
	"syscall/js"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

// ask is a method that asks the user for input.
func (h *handler) ask() {
	if commandant := config.GlobalCall(
		"prompt",
		config.Execute(config.Config.MessageBox.Messages.Prompt),
		h.spaceship.Commandant,
	); commandant.Truthy() && commandant.String() != "" {

		h.spaceship.Commandant = commandant.String()
	}
}

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
			fps := float64(frameCount) / elapsed
			config.UpdateFPS(fps)

			switch {
			case fps <= config.Config.Control.CriticalFramesPerSecondRate:
				if running.Get(h.ctx) { // If the game is running
					if config.Config.Control.Debug.Get() {
						config.Log(fmt.Sprintf("Performance dropped to %f FPS", fps))
					}

					running.Set(&h.ctx, false)  // Pause the game
					suspended.Set(&h.ctx, true) // Set the suspended state
				}

			case fps >= (config.Config.Control.CriticalFramesPerSecondRate+config.Config.Control.DesiredFramesPerSecondRate)/2 &&
				!running.Get(h.ctx):

				if suspendedFrameCount < config.Config.Control.SuspensionFrames {
					// Increase the suspended frame count
					suspendedFrameCount++

				} else {
					// Reset the suspended frame count
					suspendedFrameCount = 0

					if !paused.Get(h.ctx) { // Do not resume the game if it is paused
						if config.Config.Control.Debug.Get() {
							config.Log(fmt.Sprintf("Performance improved to %f FPS", fps))
						}

						running.Set(&h.ctx, true)    // Resume the game
						suspended.Set(&h.ctx, false) // Reset the suspended state
					}

				}

			}
			frameCount, lastFrameTime = 0, now
		}

		// Schedule the next frame
		config.GlobalCall("requestAnimationFrame", js.FuncOf(watchdog))
		return nil
	}

	// Schedule the first frame
	config.GlobalCall("requestAnimationFrame", js.FuncOf(watchdog))
}

// registerEventHandlers is a method that registers the event listeners.
func (h *handler) registerEventHandlers() {
	h.once.Do(func() {
		config.GlobalSet("drawFunc", js.FuncOf(func(_ js.Value, _ []js.Value) any {
			h.draw()
			return nil
		}))

		config.GlobalSet("onlineFunc", js.FuncOf(func(_ js.Value, _ []js.Value) any {
			offline.Set(&h.ctx, false)
			return nil
		}))

		config.GlobalSet("offlineFunc", js.FuncOf(func(_ js.Value, _ []js.Value) any {
			offline.Set(&h.ctx, true)
			return nil
		}))

		if config.IsTouchDevice() {
			globalTouchEvent := &touchEvent{mutex: &sync.Mutex{}}
			config.GlobalSet("touchstart", globalTouchEvent.touchStart(h.touchEvent))
			config.GlobalSet("touchmove", globalTouchEvent.touchMove(h.touchEvent))
			config.GlobalSet("touchend", globalTouchEvent.touchEnd(h.touchEvent))
			config.AddEventListenerToCanvas("touchstart", config.GlobalGet("touchstart"))
			config.AddEventListenerToCanvas("touchmove", config.GlobalGet("touchmove"))
			config.AddEventListenerToCanvas("touchend", config.GlobalGet("touchend"))

		} else {
			globalKeyMap := registeredKeys{
				ArrowDown:  true,
				ArrowLeft:  true,
				ArrowRight: true,
				ArrowUp:    true,
				Pause:      true,
				Space:      true,
			}
			config.GlobalSet("keydown", globalKeyMap.keyDown(h.keyEvent))
			config.GlobalSet("keyup", globalKeyMap.keyUp(h.keyEvent))
			config.AddEventListener("keydown", config.GlobalGet("keydown"))
			config.AddEventListener("keyup", config.GlobalGet("keyup"))

			globalMouseEvent := &mouseEvent{mutex: &sync.Mutex{}}
			config.GlobalSet("mousedown", globalMouseEvent.mouseDown(h.mouseEvent))
			config.GlobalSet("mousemove", globalMouseEvent.mouseMove(h.mouseEvent))
			config.GlobalSet("mouseup", globalMouseEvent.mouseUp(h.mouseEvent))
			config.AddEventListenerToCanvas("contextmenu", config.GlobalGet("mousedown"))
			config.AddEventListenerToCanvas("mousedown", config.GlobalGet("mousedown"))
			config.AddEventListenerToCanvas("mousemove", config.GlobalGet("mousemove"))
			config.AddEventListenerToCanvas("mouseup", config.GlobalGet("mouseup"))
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
func (event *mouseEvent) mouseDown(rcv chan<- mouseEvent) js.Func {
	return js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("preventDefault")

		canvasDimensions := config.CanvasBoundingBox()
		event.
			Reset().
			SetStartPosition(numeric.Position{
				X: numeric.Number(p[0].Get("clientX").Float() - canvasDimensions.BoxLeft),
				Y: numeric.Number(p[0].Get("clientY").Float() - canvasDimensions.BoxTop),
			}).
			SetButton(mouseButton(p[0].Get("button").Int())).
			SetStartTime(time.Now()).
			SetType(MouseEventTypeDown).
			SetPressed(true).
			Send(rcv)

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
		_ = event.
			SetCurrentPosition(numeric.Position{
				X: numeric.Number(p[0].Get("clientX").Float() - canvasDimensions.BoxLeft),
				Y: numeric.Number(p[0].Get("clientY").Float() - canvasDimensions.BoxTop),
			}).
			SetType(MouseEventTypeMove)

		// Check which buttons are pressed
		switch buttons, btnType := p[0].Get("buttons").Int(), mouseButton(p[0].Get("button").Int()); {
		case buttons&1 != 0 && btnType == MouseButtonPrimary:
			_ = event.SetPressed(true).SetButton(MouseButtonPrimary)

		case buttons&2 != 0 && btnType == MouseButtonSecondary:
			_ = event.SetPressed(true).SetButton(MouseButtonSecondary)

		case buttons&4 != 0 && btnType == MouseButtonAuxiliary:
			_ = event.SetPressed(true).SetButton(MouseButtonAuxiliary)

		default:
			_ = event.SetPressed(false).SetButton(btnType) // No buttons pressed
		}

		event.Send(rcv)
		return nil
	})
}

// mouseUp is a method that listens to the mouseup event.
func (event *mouseEvent) mouseUp(rcv chan<- mouseEvent) js.Func {
	return js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("preventDefault")
		canvasDimensions := config.CanvasBoundingBox()
		event.
			SetEndPosition(numeric.Position{
				X: numeric.Number(p[0].Get("clientX").Float() - canvasDimensions.BoxLeft),
				Y: numeric.Number(p[0].Get("clientY").Float() - canvasDimensions.BoxTop),
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
			SetEndPosition(numeric.Position{
				X: numeric.Number(changedTouches.Index(0).Get("clientX").Float() - canvasDimensions.BoxLeft),
				Y: numeric.Number(changedTouches.Index(0).Get("clientY").Float() - canvasDimensions.BoxTop),
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
			SetCurrentPosition(numeric.Position{
				X: numeric.Number(changedTouches.Index(0).Get("clientX").Float() - canvasDimensions.BoxLeft),
				Y: numeric.Number(changedTouches.Index(0).Get("clientY").Float() - canvasDimensions.BoxTop),
			}).
			SetMultiTap(changedTouches.Length() > 1).
			SetType(TouchTypeMove).
			Send(rcv)

		return nil
	})
}

// touchStart is a method that listens to the touchstart event.
func (event *touchEvent) touchStart(rcv chan<- touchEvent) js.Func {
	return js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("preventDefault")
		changedTouches := p[0].Get("changedTouches")
		canvasDimensions := config.CanvasBoundingBox()
		event.
			Reset().
			SetStartPosition(numeric.Position{
				X: numeric.Number(changedTouches.Index(0).Get("clientX").Float() - canvasDimensions.BoxLeft),
				Y: numeric.Number(changedTouches.Index(0).Get("clientY").Float() - canvasDimensions.BoxTop),
			}).
			SetStartTime(time.Now()).
			SetMultiTap(changedTouches.Length() > 1).
			SetType(TouchTypeStart).
			Send(rcv)

		return nil
	})
}
