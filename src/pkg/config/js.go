//go:build js && wasm

package config

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"syscall/js"
)

var (
	audioCtx = func() js.Value {
		ctx := js.Global().Get("AudioContext").New()
		if !ctx.Truthy() {
			ctx = js.Global().Get("webkitAudioContext").New()
		}
		return ctx
	}()

	audioPlayers = make(map[string]struct {
		source    js.Value
		startTime float64
		playing   bool
	})
	audioPlayersMutex = sync.Mutex{}
	audioTracks       = make(map[string][]byte)
	audioTracksMutex  = sync.Mutex{}
	canvas            = doc.Call("getElementById", "gameCanvas")
	console           = js.Global().Get("console")
	ctx               = func() js.Value {
		contextOpts := js.Global().Get("Object").New()
		contextOpts.Set("willReadFrequently", true)
		return canvas.Call("getContext", "2d", contextOpts)
	}()
	doc                    = js.Global().Get("document")
	env                    = js.Global().Get("go_env")
	invisibleCanvas        = doc.Call("createElement", "canvas")
	invisibleCtx           = invisibleCanvas.Call("getContext", "2d")
	invisibleCanvasScrollY = 0.0
	messageBox             = doc.Call("getElementById", "message")
	navigator              = js.Global().Get("navigator")
	window                 = js.Global().Get("window")
	windowLocation         = window.Get("location")
)

func init() {
	invisibleCanvas.Set("width", canvas.Get("width"))
	invisibleCanvas.Set("height", canvas.Get("height"))
	window.Call("addEventListener", "resize", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		invisibleCanvas.Set("width", canvas.Get("width"))
		invisibleCanvas.Set("height", canvas.Get("height"))
		return nil
	}))
}

// AddEventListener is a function that adds an event listener to the document.
func AddEventListener(event string, listener any) {
	doc.Call("addEventListener", event, listener)
}

// CanvasWidth returns the width of the canvas (in px).
func CanvasWidth() float64 { return canvas.Get("width").Float() }

// CanvasHeight returns the height of the canvas (in px).
func CanvasHeight() float64 { return canvas.Get("height").Float() }

// ClearBackground is a function that clears the invisible canvas.
func ClearBackground() {
	invisibleCtx.Call("clearRect", 0, 0, invisibleCanvas.Get("width").Float(), invisibleCanvas.Get("height").Float())
}

// ClearCanvas is a function that clears the canvas.
func ClearCanvas() {
	ctx.Call("clearRect", 0, 0, canvas.Get("width").Float(), canvas.Get("height").Float())
}

// DrawBackground is a function that draws the background of the canvas.
// The background is drawn with the specified speed.
func DrawBackground(speed float64) {
	invisibleCanvasScrollY += speed
	if invisibleCanvasScrollY >= CanvasHeight() {
		invisibleCanvasScrollY = 0
	}

	ctx.Call("drawImage", invisibleCanvas, 0, invisibleCanvasScrollY)
	ctx.Call("drawImage", invisibleCanvas, 0, invisibleCanvasScrollY-CanvasHeight())
}

// DrawLine is a function that draws a line on the canvas.
func DrawLine(start, end [2]float64, color string, thickness float64) {
	defaultLineWidth := ctx.Get("lineWidth")
	defer ctx.Set("lineWidth", defaultLineWidth)

	ctx.Set("strokeStyle", color)
	ctx.Set("lineWidth", thickness)
	ctx.Call("beginPath")
	ctx.Call("moveTo", start[0], start[1])
	ctx.Call("lineTo", end[0], end[1])
	ctx.Call("stroke")
}

// DrawRect is a function that draws a rectangle on the canvas.
func DrawRect(coords [2]float64, size [2]float64, color string, cornerRadius float64) {
	x, y := coords[0], coords[1]
	width, height := size[0], size[1]

	if cornerRadius == 0 {
		ctx.Set("fillStyle", color)
		ctx.Call("fillRect", x, y, width, height)
		return
	}

	ctx.Set("fillStyle", color)
	ctx.Call("beginPath")
	ctx.Call("moveTo", x+cornerRadius, y)
	ctx.Call("lineTo", x+width-cornerRadius, y)
	ctx.Call("quadraticCurveTo", x+width, y, x+width, y+cornerRadius)
	ctx.Call("lineTo", x+width, y+height-cornerRadius)
	ctx.Call("quadraticCurveTo", x+width, y+height, x+width-cornerRadius, y+height)
	ctx.Call("lineTo", x+cornerRadius, y+height)
	ctx.Call("quadraticCurveTo", x, y+height, x, y+height-cornerRadius)
	ctx.Call("lineTo", x, y+cornerRadius)
	ctx.Call("quadraticCurveTo", x, y, x+cornerRadius, y)
	ctx.Call("fill")
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

// DrawStar draws a star on the invisible canvas to be used as a background on the visible one.
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

	// Draw the star
	// Darken the color based on the brightness
	for _, c := range []string{color, fmt.Sprintf("rgba(0, 0, 0, %.2f)", 1-brightness)} {
		invisibleCtx.Call("beginPath")
		invisibleCtx.Set("fillStyle", c)
		invisibleCtx.Call("moveTo", cx, cy-radius)
		for i := 1; i < len(positions)-1; i++ {
			invisibleCtx.Call("lineTo", positions[i][0], positions[i][1])
		}
		invisibleCtx.Call("lineTo", cx, cy-radius)
		invisibleCtx.Call("closePath")
		invisibleCtx.Call("fill")
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

// LoadAudio is a function that loads an audio file.
func LoadAudio(name string) ([]byte, error) {
	protocol := windowLocation.Get("protocol").String()
	hostname := windowLocation.Get("hostname").String()
	port := windowLocation.Get("port").String()

	response, err := http.Get(fmt.Sprintf("%s//%s:%s/audio/%s", protocol, hostname, port, name))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return raw, nil
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

// PauseAudio is a function that pauses an audio track.
func PauseAudio(name string) {
	audioPlayersMutex.Lock()
	defer audioPlayersMutex.Unlock()

	if player, ok := audioPlayers[name]; ok && player.playing {
		player.startTime = audioCtx.Get("currentTime").Float()
		player.playing = false
		player.source.Call("stop")
		audioPlayers[name] = player
	}
}

// PauseAudioSources is a function that pauses all audio sources that match the selector.
func PauseAudioSources(selector func(name string) bool) {
	audioPlayersMutex.Lock()
	defer audioPlayersMutex.Unlock()

	for name, player := range audioPlayers {
		if selector(name) && player.playing {
			player.startTime = audioCtx.Get("currentTime").Float()
			player.playing = false
			player.source.Call("stop")
			audioPlayers[name] = player
		}
	}
}

// PlayAudio is a function that plays an audio track.
func PlayAudio(name string, loop bool) {
	if !*Config.Control.AudioEnabled {
		return
	}

	audioPlayersMutex.Lock()
	if _, ok := audioPlayers[name]; ok {
		audioPlayersMutex.Unlock()
		return
	}
	audioPlayersMutex.Unlock()

	audioTracksMutex.Lock()
	track, ok := audioTracks[name]
	audioTracksMutex.Unlock()
	if !ok {
		raw, err := LoadAudio(name)
		if err != nil {
			LogError(err)
			return
		}

		audioTracksMutex.Lock()
		audioTracks[name], track = raw, raw
		audioTracksMutex.Unlock()
	}

	buffer := js.Global().Get("Uint8Array").New(len(track))
	js.CopyBytesToJS(buffer, track)

	audioBufferPromise := audioCtx.Call("decodeAudioData", buffer.Get("buffer"))
	then := js.FuncOf(func(_ js.Value, p []js.Value) any {
		source := audioCtx.Call("createBufferSource")
		source.Set("buffer", p[0])
		source.Call("connect", audioCtx.Get("destination"))

		source.Call("addEventListener", "ended", js.FuncOf(func(_ js.Value, _ []js.Value) any {
			audioPlayersMutex.Lock()
			delete(audioPlayers, name)
			audioPlayersMutex.Unlock()

			if loop {
				defer PlayAudio(name, loop)
			}

			return nil
		}))

		audioPlayersMutex.Lock()
		audioPlayers[name] = struct {
			source    js.Value
			startTime float64
			playing   bool
		}{source, 0, true}
		audioPlayersMutex.Unlock()

		source.Call("start", js.ValueOf(0), js.ValueOf(0))
		return nil
	})
	catch := js.FuncOf(func(_ js.Value, p []js.Value) any {
		message := p[0].Get("message").String()
		stack := p[0].Get("stack").String()
		LogError(fmt.Errorf("failed to decode audio data: %s\n%s\n", message, stack))
		return nil
	})
	audioBufferPromise.Call("then", then).Call("catch", catch)
}

// RemoveEventListener is a function that removes an event listener from the document.
func RemoveEventListener(event string, listener any) {
	doc.Call("removeEventListener", event, listener)
}

// ResumeAudio is a function that resumes an audio track.
func ResumeAudio(name string) {
	audioPlayersMutex.Lock()
	defer audioPlayersMutex.Unlock()

	if player, ok := audioPlayers[name]; ok && !player.playing {
		player.startTime = 0
		player.playing = true
		player.source.Call("start", js.ValueOf(0), js.ValueOf(player.startTime))
		audioPlayers[name] = player
	}
}

// ResumeAudioSources is a function that resumes all audio sources that match the selector.
func ResumeAudioSources(selector func(name string) bool) {
	audioPlayersMutex.Lock()
	defer audioPlayersMutex.Unlock()

	for name, player := range audioPlayers {
		if selector(name) && !player.playing {
			player.startTime = 0
			player.playing = true
			player.source.Call("start", js.ValueOf(0), js.ValueOf(player.startTime))
			audioPlayers[name] = player
		}
	}
}

// SendMessage sends a message to the message box.
func SendMessage(msg string) {
	content := messageBox.Get("innerHTML").String()
	lines := append(strings.Split(content, "<br>"), msg)
	if len(lines) > Config.MessageBox.BufferSize {
		lines = lines[len(lines)-Config.MessageBox.BufferSize:]
	}

	messageBox.Set("innerHTML", strings.Join(lines, "<br>"))
	messageBox.Set("scrollTop", messageBox.Get("scrollHeight"))
}

// Setenv is a function that sets the environment variable key to value.
func Setenv(key, value string) {
	env.Set(key, value)
}

// StopAudio is a function that stops an audio track.
func StopAudio(name string) {
	audioPlayersMutex.Lock()
	defer audioPlayersMutex.Unlock()

	if player, ok := audioPlayers[name]; ok {
		player.source.Call("stop")
		delete(audioPlayers, name)
	}
}

// StopAudioSources is a function that stops all audio sources that match the selector.
func StopAudioSources(selector func(name string) bool) {
	audioPlayersMutex.Lock()
	defer audioPlayersMutex.Unlock()

	for name, player := range audioPlayers {
		if selector(name) {
			player.source.Call("stop")
			delete(audioPlayers, name)
		}
	}
}

// ThrowError is a function that throws an error.
func ThrowError(err error) {
	if err != nil {
		js.Global().Call("eval", fmt.Sprintf("throw new Error('%s')", err.Error()))
	}
}

// Unsetenv is a function that unsets the environment variable key.
func Unsetenv(key string) {
	env.Delete(key)
}
