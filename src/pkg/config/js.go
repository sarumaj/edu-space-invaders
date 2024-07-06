//go:build js && wasm

package config

import (
	"fmt"
	"syscall/js"
)

var (
	canvas     = doc.Call("getElementById", "gameCanvas")
	console    = js.Global().Get("console")
	ctx        = canvas.Call("getContext", "2d")
	doc        = js.Global().Get("document")
	env        = js.Global().Get("go_env")
	messageBox = doc.Call("getElementById", "message")
	navigator  = js.Global().Get("navigator")
	window     = js.Global().Get("window")
)

// AddEventListener is a function that adds an event listener to the document.
func AddEventListener(event string, listener any) {
	doc.Call("addEventListener", event, listener)
}

// CanvasWidth returns the width of the canvas (in px).
func CanvasWidth() float64 { return canvas.Get("width").Float() }

// CanvasHeight returns the height of the canvas (in px).
func CanvasHeight() float64 { return canvas.Get("height").Float() }

// ClearCanvas is a function that clears the canvas.
func ClearCanvas() {
	ctx.Call("clearRect", 0, 0, canvas.Get("width").Float(), canvas.Get("height").Float())
}

// DrawRect is a function that draws a rectangle on the canvas.
func DrawRect(x, y, width, height float64, color string) {
	ctx.Set("fillStyle", color)
	ctx.Call("fillRect", x, y, width, height)
}

// DrawSpaceship is a function that draws a spaceship on the canvas.
// The spaceship is drawn at the specified position (x, y) with the specified width and height.
// The spaceship is drawn facing the specified direction.
func DrawSpaceship(x, y, width, height float64, faceUp bool, color string) {
	ctx.Set("fillStyle", color)
	ctx.Set("strokeStyle", "black")

	// Draw the body of the spaceship
	ctx.Call("fillRect", x+width*0.4, y+height*0.2, width*0.2, height*0.6)
	ctx.Call("strokeRect", x+width*0.4, y+height*0.2, width*0.2, height*0.6)

	// Draw the wings
	ctx.Call("beginPath")
	ctx.Call("moveTo", x+width*0.4, y+height*0.2) // Left point of left wing
	if faceUp {
		ctx.Call("lineTo", x, y+height*0.75) // Bottom point of left wing
	} else {
		ctx.Call("lineTo", x, y+height*0.25) // Bottom point of left wing
	}
	ctx.Call("lineTo", x+width*0.4, y+height*0.8) // Right point of left wing
	ctx.Call("closePath")
	ctx.Call("fill")
	ctx.Call("stroke")

	ctx.Call("beginPath")
	ctx.Call("moveTo", x+width*0.6, y+height*0.2) // Right point of right wing
	if faceUp {
		ctx.Call("lineTo", x+width, y+height*0.75) // Bottom point of right wing
	} else {
		ctx.Call("lineTo", x+width, y+height*0.25) // Bottom point of right wing
	}
	ctx.Call("lineTo", x+width*0.6, y+height*0.8) // Left point of right wing
	ctx.Call("closePath")
	ctx.Call("fill")
	ctx.Call("stroke")

	// Draw the tip of the spaceship
	ctx.Call("beginPath")
	if faceUp {
		ctx.Call("moveTo", x+width*0.4, y+height*0.2) // Left point of the tip
		ctx.Call("lineTo", x+width*0.5, y)            // Top point of the tip
		ctx.Call("lineTo", x+width*0.6, y+height*0.2) // Right point of the tip
	} else {
		ctx.Call("moveTo", x+width*0.4, y+height*0.8) // Left point of the tip
		ctx.Call("lineTo", x+width*0.5, y+height)     // Bottom point of the tip
		ctx.Call("lineTo", x+width*0.6, y+height*0.8) // Right point of the tip
	}
	ctx.Call("closePath")
	ctx.Call("fill")
	ctx.Call("stroke")
}

// Getenv is a function that returns the value of the environment variable key.
func Getenv(key string) string {
	got := env.Get(key)
	if !got.Truthy() {
		return ""
	}

	return got.String()
}

// IsTouchDevice is a function that returns true if the device is a touch device.
func IsTouchDevice() bool {
	switch {
	case window.Call("hasOwnProperty", "ontouchstart").Bool():
		return true

	case navigator.Truthy():
		if maxTouchPoints := navigator.Get("maxTouchPoints"); maxTouchPoints.Truthy() && maxTouchPoints.Int() > 0 {
			return true
		}

		if msMaxTouchPoints := navigator.Get("msMaxTouchPoints"); msMaxTouchPoints.Truthy() && msMaxTouchPoints.Int() > 0 {
			return true
		}

	}

	return false
}

// Log is a function that logs a message.
func Log(msg string) {
	console.Call("log", msg)
}

// LogError is a function that logs an error.
func LogError(err error) {
	if err != nil {
		console.Call("error", err.Error())
	}
}

// SendMessage sends a message to the message box.
func SendMessage(msg string) {
	messageBox.Set("innerText", msg)
}

// ThrowError is a function that throws an error.
func ThrowError(err error) {
	if err != nil {
		js.Global().Call("eval", fmt.Sprintf("throw new Error('%s')", err.Error()))
	}
}
