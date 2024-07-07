//go:build !js || !wasm

package config

import (
	"log"
	"os"
)

// AddEventListener is a function that adds an event listener to the document.
func AddEventListener(event string, listener any) {}

// CanvasWidth returns the width of the canvas (in px).
func CanvasWidth() float64 { return 800 }

// CanvasHeight returns the height of the canvas (in px).
func CanvasHeight() float64 { return 600 }

// ClearCanvas is a function that clears the canvas.
func ClearCanvas() {}

// DrawRect is a function that draws a rectangle on the canvas.
func DrawRect(coors [2]float64, size [2]float64, color string) {}

// DrawSpaceship is a function that draws a spaceship on the canvas.
// The spaceship is drawn at the specified position (x, y) with the specified width and height.
// The spaceship is drawn facing the specified direction.
func DrawSpaceship(coors [2]float64, size [2]float64, faceUp bool, color string) {}

// DrawStar draws a star on the canvas.
// The star is drawn at the specified position (cx, cy) with the specified number of spikes.
// The outer radius and inner radius of the star are specified.
// The star is filled with the specified color.
func DrawStar(coords [2]float64, spikes, radius float64, color string, brightness float64) {}

// Getenv is a function that returns the value of the environment variable key.
func Getenv(key string) string {
	return os.Getenv(key)
}

// IsTouchDevice is a function that returns true if the device is a touch device.
func IsTouchDevice() bool {
	return false
}

// Log is a function that logs a message.
func Log(msg string) {
	log.Println(msg)
}

// LogError is a function that logs an error.
func LogError(err error) {
	if err != nil {
		log.Println(err)
	}
}

// SendMessage sends a message to the message box.
func SendMessage(msg string) {
	log.Println(msg)
}

// ThrowError is a function that throws an error.
func ThrowError(err error) {
	if err != nil {
		panic(err)
	}
}
