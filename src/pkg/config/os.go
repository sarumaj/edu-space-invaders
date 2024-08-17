//go:build !js || !wasm

package config

import (
	"log"
	"os"
)

// dimensions represents the dimensions of the document.
type dimensions struct {
	BoxWidth, BoxHeight                  float64
	BoxLeft, BoxTop, BoxRight, BoxBottom float64
	OriginalWidth, OriginalHeight        float64
	ScaleWidth, ScaleHeight              float64
}

// score represents a score.
type score struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

// AddEventListener is a function that adds an event listener to the document.
func AddEventListener(event string, listener any) {}

// AddEventListenerToCanvas is a function that adds an event listener to the canvas.
func AddEventListenerToCanvas(event string, listener any) {}

// CanvasBoundingBox returns the bounding box of the canvas.
func CanvasBoundingBox() dimensions {
	return dimensions{BoxWidth: 800, BoxHeight: 600, BoxRight: 800, BoxBottom: 600}
}

// ClearBackground is a function that clears the invisible canvas.
func ClearBackground() {}

// ClearCanvas is a function that clears the canvas.
func ClearCanvas() {}

// ConvertArrayToSlice is a function that converts an array to a slice.
func ConvertArrayToSlice(array any) []any { return nil }

// ConvertObjectToMap is a function that converts an object to a map.
func ConvertObjectToMap(obj any) map[string]any { return nil }

// DrawBackground is a function that draws the background of the canvas.
// The background is drawn with the specified speed.
func DrawBackground(speed float64) {}

// DrawAnomalyBlackHole is a function that draws a black hole on the document.
func DrawAnomalyBlackHole(coords [2]float64, radius float64) {}

// DrawAnomalySupernova is a function that draws a supernova on the document.
func DrawAnomalySupernova(coords [2]float64, radius float64) {}

// DrawLine is a function that draws a line on the canvas.
func DrawLine(start, end [2]float64, color string, thickness float64) {}

// DrawPlanetEarth is a function that draws Earth on the document.
func DrawPlanetEarth(coords [2]float64, radius float64) {}

// DrawPlanetJupiter is a function that draws Jupiter on the document.
func DrawPlanetJupiter(coords [2]float64, radius float64) {}

// DrawPlanetMars is a function that draws Mars on the document.
func DrawPlanetMars(coords [2]float64, radius float64) {}

// DrawPlanetMercury is a function that draws Mercury on the document.
func DrawPlanetMercury(coords [2]float64, radius float64) {}

// DrawPlanetNeptune is a function that draws Neptune on the document.
func DrawPlanetNeptune(coords [2]float64, radius float64) {}

// DrawPlanetPluto is a function that draws Pluto on the document.
func DrawPlanetPluto(coords [2]float64, radius float64) {}

// DrawPlanetSaturn is a function that draws Saturn on the document.
func DrawPlanetSaturn(coords [2]float64, radius float64) {}

// DrawPlanetUranus is a function that draws Uranus on the document.
func DrawPlanetUranus(coords [2]float64, radius float64) {}

// DrawPlanetVenus is a function that draws Venus on the document.
func DrawPlanetVenus(coords [2]float64, radius float64) {}

// DrawRect is a function that draws a rectangle on the canvas.
func DrawRect(coords [2]float64, size [2]float64, color string, cornerRadius float64) {}

// DrawSpaceship is a function that draws a spaceship on the canvas.
// The spaceship is drawn at the specified position (x, y) with the specified width and height.
// The spaceship is drawn facing the specified direction.
func DrawSpaceship(coors [2]float64, size [2]float64, faceUp bool, color, label string) {}

// DrawStar draws a star on the canvas.
// The star is drawn at the specified position (cx, cy) with the specified number of spikes.
// The outer radius and inner radius of the star are specified.
// The star is filled with the specified color.
func DrawStar(coords [2]float64, spikes int, radius, innerRadius float64, color string, brightness float64) {
}

// DrawSun is a function that draws the Sun on the document.
func DrawSun(coords [2]float64, radius float64) {}

// Getenv is a function that returns the value of the environment variable key.
func Getenv(key string) string { return os.Getenv(key) }

// GetScores is a function that returns the scores.
func GetScores(top int) (scores []score) { return }

// GlobalCall is a function that calls the global function name with the specified arguments.
func GlobalCall(name string, args ...any) any { return nil }

// GlobalGet is a function that returns the global value of key.
func GlobalGet(key string) any { return nil }

// GlobalSet is a function that sets the global value of key to value.
func GlobalSet(key string, value any) {}

// IsPlaying is a function that returns true if the audio track is playing.
func IsPlaying(name string) bool { return false }

// IsTouchDevice is a function that returns true if the device is a touch device.
func IsTouchDevice() bool { return false }

// LoadAudio is a function that loads an audio file from the specified URL.
func LoadAudio(url string) ([]byte, error) { return nil, nil }

// Log is a function that logs a message.
func Log(msg string) { log.Println(msg) }

// LogError is a function that logs an error.
func LogError(err error) {
	if err != nil {
		log.Println(err)
	}
}

// MakeObject is a function that returns a new object with the specified key-value pairs.
func MakeObject(m map[string]any) any { return m }

// NewInstance is a function that returns a new instance of the type with the specified arguments.
func NewInstance(typ string, args ...any) any { return nil }

// PlayAudio is a function that plays an audio track.
func PlayAudio(name string, loop bool) {}

// SaveScores is a function that saves the score board persistently.
func SaveScores() {}

// SendMessage sends a message to the message box.
func SendMessage(msg string, reset bool) { log.Println(msg) }

// Setenv is a function that sets the environment variable key to value.
func Setenv(key, value string) { _ = os.Setenv(key, value) }

// SetScore is a function that sets the score.
func SetScore(name string, score int) (rank int) { return }

// StopAudio is a function that stops an audio track.
func StopAudio(name string) {}

// StopAudioSources is a function that stops all audio sources that match the selector.
func StopAudioSources(selector func(name string) bool) {}

// ThrowError is a function that throws an error.
func ThrowError(err error) {
	if err != nil {
		panic(err)
	}
}

// Unsetenv is a function that unsets the environment variable key.
func Unsetenv(key string) { _ = os.Unsetenv(key) }

// UpdateFPS is a function that updates the frames per second.
func UpdateFPS(fps float64) {}

// UpdateMessage is a function that updates the last message in the message box.
func UpdateMessage(msg string) {}
