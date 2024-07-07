//go:build js && wasm

package config

import (
	"fmt"
	"math"
	"strings"
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
func DrawRect(coors [2]float64, size [2]float64, color string) {
	x, y := coors[0], coors[1]
	width, height := size[0], size[1]

	ctx.Set("fillStyle", color)
	ctx.Call("fillRect", x, y, width, height)
}

// DrawSpaceship is a function that draws a spaceship on the canvas.
// The spaceship is drawn at the specified position (x, y) with the specified width and height.
// The spaceship is drawn facing the specified direction.
func DrawSpaceship(coors [2]float64, size [2]float64, faceUp bool, color string) {
	x, y := coors[0], coors[1]
	width, height := size[0], size[1]

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

// DrawStar draws a star on the canvas.
// The star is drawn at the specified position (cx, cy) with the specified number of spikes.
// The outer radius and inner radius of the star are specified.
// The star is filled with the specified color.
func DrawStar(coords [2]float64, spikes, radius float64, color string, brightness float64) {
	rot := math.Pi / 2 * 3         // Starting rotation angle (90 degrees)
	cx, cy := coords[0], coords[1] // Center position
	x, y := cx, cy                 // Starting position
	step := math.Pi / spikes       // Angle between each spike

	// Calculate the positions of the star
	var positions [][2]float64
	positions = append(positions, [2]float64{cx, cy - radius})
	for i := 0; i < int(spikes); i++ {
		x = cx + math.Cos(rot)*radius
		y = cy + math.Sin(rot)*radius
		positions = append(positions, [2]float64{x, y})
		rot += step

		// inner radius
		x = cx + math.Cos(rot)*radius/2
		y = cy + math.Sin(rot)*radius/2
		positions = append(positions, [2]float64{x, y})
		rot += step
	}
	positions = append(positions, [2]float64{cx, cy - radius})

	// Draw the star
	// Darken the color based on the brightness
	first := positions[0]
	last := positions[len(positions)-1]
	for _, c := range []string{color, fmt.Sprintf("rgba(0, 0, 0, %.2f)", 1-brightness)} {
		ctx.Call("beginPath")
		ctx.Set("fillStyle", c)
		ctx.Call("moveTo", first[0], first[1])
		for i := 1; i < len(positions)-1; i++ {
			ctx.Call("lineTo", positions[i][0], positions[i][1])
		}
		ctx.Call("lineTo", last[0], last[1])
		ctx.Call("closePath")
		ctx.Call("fill")
	}
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
	content := messageBox.Get("innerText").String()
	lines := append(strings.Split(content, "\n"), msg)
	if len(lines) > Config.MessageBox.BufferSize {
		lines = lines[len(lines)-Config.MessageBox.BufferSize:]
	}

	messageBox.Set("innerText", strings.Join(lines, "\n"))
	messageBox.Set("scrollTop", messageBox.Get("scrollHeight"))
}

// ThrowError is a function that throws an error.
func ThrowError(err error) {
	if err != nil {
		js.Global().Call("eval", fmt.Sprintf("throw new Error('%s')", err.Error()))
	}
}
