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

const (
	originalWidth  = 760
	originalHeight = 570
)

const (
	audioIconId      = "audioIcon"
	audioIconMuted   = "fa-volume-mute"
	audioIconUnmuted = "fa-volume-up"
	audioToggleBtnId = "audioToggle"
	canvasId         = "gameCanvas"
	goEnv            = "go_env"
	messageBoxId     = "message"
	refreshButtonId  = "refreshButton"
)

var (
	audioCtx               = getAudioContext()
	audioPlayers           = make(map[string]audioPlayer)
	audioPlayersMutex      = sync.RWMutex{}
	audioTracks            = make(map[string][]byte)
	audioTracksMutex       = sync.RWMutex{}
	canvasObject           = document.Call("getElementById", canvasId)
	canvasObjectContext    = canvasObject.Call("getContext", "2d", MakeObject(map[string]any{"willReadFrequently": true}))
	console                = GlobalGet("console")
	document               = GlobalGet("document")
	environ                = GlobalGet(goEnv)
	invisibleCanvas        = document.Call("createElement", "canvas")
	invisibleCtx           = invisibleCanvas.Call("getContext", "2d")
	invisibleCanvasScrollY = 0.0
	messageBox             = document.Call("getElementById", messageBoxId)
	window                 = GlobalGet("window")
	windowLocation         = window.Get("location")
)

type audioPlayer struct {
	endedCallback js.Func
	source        js.Value
	startTime     float64
}

type dimensions struct {
	Width, Height            float64
	Left, Top, Right, Bottom float64
	ScaleX, ScaleY           float64
}

func init() {
	setupAudioInterface()
	setupRefreshInterface()
	setupCanvasInterface()
}

// getAudioContext is a function that returns the audio context.
func getAudioContext() js.Value {
	ctx := NewInstance("AudioContext")
	if !ctx.Truthy() {
		ctx = NewInstance("webkitAudioContext")
	}
	return ctx
}

// setupAudioInterface is a function that sets up the audio interface.
// The audio interface includes the audio icon and the audio toggle button.
// The audio icon is updated based on the audio state.
// The audio toggle button toggles the audio state.
// The audio toggle button is updated based on the audio state.
func setupAudioInterface() {
	audioIcon := document.Call("getElementById", audioIconId)

	if *Config.Control.AudioEnabled {
		audioIcon.Get("classList").Call("remove", audioIconMuted)
		audioIcon.Get("classList").Call("add", audioIconUnmuted)
	} else {
		audioIcon.Get("classList").Call("remove", audioIconUnmuted)
		audioIcon.Get("classList").Call("add", audioIconMuted)
	}

	audioToggle := func() {
		*Config.Control.AudioEnabled = !*Config.Control.AudioEnabled

		if *Config.Control.AudioEnabled {
			audioIcon.Get("classList").Call("remove", audioIconMuted)
			audioIcon.Get("classList").Call("add", audioIconUnmuted)

			go PlayAudio("theme_heroic.wav", true)
		} else {
			audioIcon.Get("classList").Call("remove", audioIconUnmuted)
			audioIcon.Get("classList").Call("add", audioIconMuted)

			go StopAudioSources(func(string) bool { return true })
		}
	}

	GlobalSet("toggleAudioClick", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		audioToggle()
		return nil
	}))

	GlobalSet("toggleAudioTouchEnd", js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("preventDefault")
		audioToggle()
		return nil
	}))

	audioToggleBtn := document.Call("getElementById", audioToggleBtnId)
	audioToggleBtn.Call("addEventListener", "click", GlobalGet("toggleAudioClick"))
	audioToggleBtn.Call("addEventListener", "touchend", GlobalGet("toggleAudioTouchEnd"))
}

// setupCanvasInterface is a function that sets up the canvas interface.
// The canvas interface includes the invisible canvas and the resize event listener.
// The invisible canvas is used to draw the background of the visible canvas.
// The resize event listener is used to redraw the document when the window is resized.
// The draw function is called when the document is resized.
func setupCanvasInterface() {
	invisibleCanvas.Set("width", canvasObject.Get("width").Float())
	invisibleCanvas.Set("height", canvasObject.Get("height").Float())

	GlobalSet("resize", js.FuncOf(func(_ js.Value, p []js.Value) any {
		canvasWidth := canvasObject.Get("width")
		canvasHeight := canvasObject.Get("height")

		data := canvasObjectContext.Call("getImageData", 0, 0, canvasWidth, canvasHeight)

		innerWidth := canvasObject.Get("clientWidth")
		innerHeight := canvasObject.Get("clientHeight")

		canvasObject.Set("width", innerWidth)
		canvasObject.Set("height", innerHeight)

		invisibleCanvas.Set("width", canvasObject.Get("width"))
		invisibleCanvas.Set("height", canvasObject.Get("height"))

		canvasObjectContext.Call("putImageData", data, 0, 0)

		if GlobalGet("drawFunc").Truthy() {
			GlobalGet("drawFunc").Invoke()
		}

		return nil
	}))

	window.Call("addEventListener", "resize", GlobalGet("resize"))
}

// setupRefreshInterface is a function that sets up the refresh interface.
// The refresh interface includes the refresh button.
// The refresh button is animated when clicked or touched.
// The document is reloaded when the refresh button is clicked or touched.
func setupRefreshInterface() {
	refreshButton := document.Call("getElementById", refreshButtonId)
	GlobalSet("animateRefreshButton", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return NewInstance("Promise", js.FuncOf(func(_ js.Value, p []js.Value) any {
			refreshButton.Get("classList").Call("add", "animated-click")
			resolve := p[0]

			transitionEndCallback := js.FuncOf(func(_ js.Value, _ []js.Value) any {
				refreshButton.Get("classList").Call("remove", "animated-click")
				refreshButton.Get("classList").Call("add", "animated-click-end")
				resolve.Invoke()

				return nil
			})

			refreshButton.Call("addEventListener", "transitionend", transitionEndCallback, MakeObject(map[string]any{"once": true}))
			return nil
		}))
	}))

	then := js.FuncOf(func(_ js.Value, _ []js.Value) any {
		windowLocation.Call("reload")
		return nil
	})

	GlobalSet("refreshButtonClick", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		GlobalGet("animateRefreshButton").Invoke().Call("then", then)
		return nil
	}))

	GlobalSet("refreshButtonTouchEnd", js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("preventDefault")
		GlobalGet("animateRefreshButton").Invoke().Call("then", then)
		return nil
	}))

	refreshButton.Call("addEventListener", "click", GlobalGet("refreshButtonClick"))
	refreshButton.Call("addEventListener", "touchend", GlobalGet("refreshButtonTouchEnd"))
}

// AddEventListener is a function that adds an event listener to the document.
func AddEventListener(event string, listener any) {
	document.Call("addEventListener", event, listener)
}

// AddEventListenerToCanvas is a function that adds an event listener to the document.
func AddEventListenerToCanvas(event string, listener any) {
	canvasObject.Call("addEventListener", event, listener)
}

// CanvasBoundingBox returns the bounding box of the document.
func CanvasBoundingBox() dimensions {
	box := canvasObject.Call("getBoundingClientRect")
	dim := dimensions{
		Left:   box.Get("left").Float(),
		Top:    box.Get("top").Float(),
		Right:  box.Get("right").Float(),
		Bottom: box.Get("bottom").Float(),
		Width:  box.Get("width").Float(),
		Height: box.Get("height").Float(),
	}

	dim.ScaleX = dim.Width / originalWidth
	dim.ScaleY = dim.Height / originalHeight

	return dim
}

// ClearBackground is a function that clears the invisible document.
func ClearBackground() {
	invisibleCtx.Call("clearRect", 0, 0, invisibleCanvas.Get("width").Float(), invisibleCanvas.Get("height").Float())
}

// ClearCanvas is a function that clears the document.
func ClearCanvas() {
	canvasObjectContext.Call("clearRect", 0, 0, canvasObject.Get("width").Float(), canvasObject.Get("height").Float())
}

// DrawBackground is a function that draws the background of the document.
// The background is drawn with the specified speed.
func DrawBackground(speed float64) {
	canvasDimensions := CanvasBoundingBox()

	// Apply the speed
	invisibleCanvasScrollY += speed
	if invisibleCanvasScrollY >= canvasDimensions.Height {
		invisibleCanvasScrollY = 0
	}

	if *Config.Control.BackgroundAnimationEnabled {
		canvasObjectContext.Call("drawImage", invisibleCanvas, 0, invisibleCanvasScrollY)
		canvasObjectContext.Call("drawImage", invisibleCanvas, 0, invisibleCanvasScrollY-canvasDimensions.Height)
	} else {
		canvasObjectContext.Call("drawImage", invisibleCanvas, 0, 0)
	}
}

// DrawLine is a function that draws a line on the document.
func DrawLine(start, end [2]float64, color string, thickness float64) {
	defaultLineWidth := canvasObjectContext.Get("lineWidth")
	defer canvasObjectContext.Set("lineWidth", defaultLineWidth)

	canvasObjectContext.Set("strokeStyle", color)
	canvasObjectContext.Set("lineWidth", thickness)
	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("moveTo", start[0], start[1])
	canvasObjectContext.Call("lineTo", end[0], end[1])
	canvasObjectContext.Call("stroke")
}

// DrawRect is a function that draws a rectangle on the document.
func DrawRect(coords [2]float64, size [2]float64, color string, cornerRadius float64) {
	x, y := coords[0], coords[1]
	width, height := size[0], size[1]

	if cornerRadius == 0 {
		canvasObjectContext.Set("fillStyle", color)
		canvasObjectContext.Call("fillRect", x, y, width, height)
		return
	}

	canvasObjectContext.Set("fillStyle", color)
	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("moveTo", x+cornerRadius, y)
	canvasObjectContext.Call("lineTo", x+width-cornerRadius, y)
	canvasObjectContext.Call("quadraticCurveTo", x+width, y, x+width, y+cornerRadius)
	canvasObjectContext.Call("lineTo", x+width, y+height-cornerRadius)
	canvasObjectContext.Call("quadraticCurveTo", x+width, y+height, x+width-cornerRadius, y+height)
	canvasObjectContext.Call("lineTo", x+cornerRadius, y+height)
	canvasObjectContext.Call("quadraticCurveTo", x, y+height, x, y+height-cornerRadius)
	canvasObjectContext.Call("lineTo", x, y+cornerRadius)
	canvasObjectContext.Call("quadraticCurveTo", x, y, x+cornerRadius, y)
	canvasObjectContext.Call("fill")
}

// DrawSpaceship is a function that draws a spaceship on the document.
// The spaceship is drawn at the specified position (x, y) with the specified width and height.
// The spaceship is drawn facing the specified direction.
func DrawSpaceship(coors [2]float64, size [2]float64, faceUp bool, color string) {
	x, y := coors[0], coors[1]
	width, height := size[0], size[1]

	canvasObjectContext.Set("fillStyle", color)
	canvasObjectContext.Set("strokeStyle", "black")

	// Draw the body of the spaceship
	canvasObjectContext.Call("fillRect", x+width*0.4, y+height*0.2, width*0.2, height*0.6)
	canvasObjectContext.Call("strokeRect", x+width*0.4, y+height*0.2, width*0.2, height*0.6)

	// Draw the wings
	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("moveTo", x+width*0.4, y+height*0.2) // Left point of left wing
	if faceUp {
		canvasObjectContext.Call("lineTo", x, y+height*0.75) // Bottom point of left wing
	} else {
		canvasObjectContext.Call("lineTo", x, y+height*0.25) // Bottom point of left wing
	}
	canvasObjectContext.Call("lineTo", x+width*0.4, y+height*0.8) // Right point of left wing
	canvasObjectContext.Call("closePath")
	canvasObjectContext.Call("fill")
	canvasObjectContext.Call("stroke")

	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("moveTo", x+width*0.6, y+height*0.2) // Right point of right wing
	if faceUp {
		canvasObjectContext.Call("lineTo", x+width, y+height*0.75) // Bottom point of right wing
	} else {
		canvasObjectContext.Call("lineTo", x+width, y+height*0.25) // Bottom point of right wing
	}
	canvasObjectContext.Call("lineTo", x+width*0.6, y+height*0.8) // Left point of right wing
	canvasObjectContext.Call("closePath")
	canvasObjectContext.Call("fill")
	canvasObjectContext.Call("stroke")

	// Draw the tip of the spaceship
	canvasObjectContext.Call("beginPath")
	if faceUp {
		canvasObjectContext.Call("moveTo", x+width*0.4, y+height*0.2) // Left point of the tip
		canvasObjectContext.Call("lineTo", x+width*0.5, y)            // Top point of the tip
		canvasObjectContext.Call("lineTo", x+width*0.6, y+height*0.2) // Right point of the tip
	} else {
		canvasObjectContext.Call("moveTo", x+width*0.4, y+height*0.8) // Left point of the tip
		canvasObjectContext.Call("lineTo", x+width*0.5, y+height)     // Bottom point of the tip
		canvasObjectContext.Call("lineTo", x+width*0.6, y+height*0.8) // Right point of the tip
	}
	canvasObjectContext.Call("closePath")
	canvasObjectContext.Call("fill")
	canvasObjectContext.Call("stroke")
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
	got := environ.Get(key)
	if !got.Truthy() {
		return ""
	}

	return got.String()
}

// GlobalCall is a function that calls the global function name with the specified arguments.
func GlobalCall(name string, args ...any) any {
	return js.Global().Call(name, args...)
}

// GlobalGet is a function that returns the global value of key.
func GlobalGet(key string) js.Value {
	return js.Global().Get(key)
}

// GlobalSet is a function that sets the global value of key to value.
func GlobalSet(key string, value any) {
	js.Global().Set(key, value)
}

// IsPlaying is a function that returns true if the audio track is playing.
func IsPlaying(name string) bool {
	audioPlayersMutex.RLock()
	player, playerOk := audioPlayers[name]
	audioPlayersMutex.RUnlock()

	if playerOk && player.source.Truthy() {
		return true
	}

	return false
}

// IsTouchDevice is a function that returns true if the device is a touch device.
func IsTouchDevice() bool {
	navigator := GlobalGet("navigator")
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

	url := fmt.Sprintf("%s//%s:%s/audio/%s", protocol, hostname, port, name)
	if port == "" {
		url = fmt.Sprintf("%s//%s/audio/%s", protocol, hostname, name)
	}

	response, err := http.Get(url)
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

// MakeObject is a function that returns a new object with the specified key-value pairs.
func MakeObject(m map[string]any) js.Value {
	obj := NewInstance("Object")
	for key, value := range m {
		obj.Set(key, value)
	}
	return obj
}

// NewInstance is a function that returns a new instance of the type with the specified arguments.
func NewInstance(typ string, args ...any) js.Value {
	return GlobalGet(typ).New(args...)
}

// PlayAudio is a function that plays an audio track.
func PlayAudio(name string, loop bool) {
	if !*Config.Control.AudioEnabled {
		return
	}

	audioPlayersMutex.RLock()
	player, playerOk := audioPlayers[name]
	audioPlayersMutex.RUnlock()

	if playerOk && player.source.Truthy() {
		if Config.Control.Debug.Get() {
			Log(fmt.Sprintf("Audio source already playing: %s", name))
		}
		return
	}

	audioTracksMutex.RLock()
	track, trackOk := audioTracks[name]
	audioTracksMutex.RUnlock()
	if !trackOk {
		raw, err := LoadAudio(name)
		if err != nil {
			LogError(err)
			return
		}

		audioTracksMutex.Lock()
		audioTracks[name], track = raw, raw
		audioTracksMutex.Unlock()
	}

	buffer := NewInstance("Uint8Array", len(track))
	js.CopyBytesToJS(buffer, track)

	audioBufferPromise := audioCtx.Call("decodeAudioData", buffer.Get("buffer"))
	then := js.FuncOf(func(_ js.Value, p []js.Value) any {
		player.source = audioCtx.Call("createBufferSource")
		player.source.Set("buffer", p[0])
		player.source.Call("connect", audioCtx.Get("destination"))

		player.endedCallback = js.FuncOf(func(_ js.Value, _ []js.Value) any {
			audioPlayersMutex.Lock()
			audioPlayers[name] = audioPlayer{
				endedCallback: player.endedCallback,
				source:        js.Null(),
				startTime:     0,
			}
			audioPlayersMutex.Unlock()

			if loop {
				defer PlayAudio(name, loop)
			}

			return nil
		})
		player.source.Call("addEventListener", "ended", player.endedCallback)

		audioPlayersMutex.Lock()
		audioPlayers[name] = player
		audioPlayersMutex.Unlock()

		if Config.Control.Debug.Get() {
			Log(fmt.Sprintf("Playing audio source: %s", name))
		}

		player.source.Call("start", js.ValueOf(0), js.ValueOf(player.startTime))
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

// RegisterDrawFunc is a function that registers a draw function.
// The draw function is called when the document is resized.
func RegisterDrawFunc(f func()) {
	GlobalSet("drawFunc", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		f()
		return nil
	}))
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
	environ.Set(key, value)
}

// StopAudio is a function that stops an audio track.
func StopAudio(name string) {
	audioPlayersMutex.RLock()
	player, playerOk := audioPlayers[name]
	audioPlayersMutex.RUnlock()

	// Recursive function to stop the audio source.
	// Recursion might be necessary if the audio source is still playing
	// when the stop function is called
	// and the event listener is at end of the audio source
	// fires after the audio source has been stopped
	var stop func(recursive int)
	stop = func(recursive int) {
		if playerOk && player.source.Truthy() {
			if Config.Control.Debug.Get() {
				Log(fmt.Sprintf("Stopping audio source: %s", name))
			}

			player.startTime = audioCtx.Get("currentTime").Float()

			player.source.Call("removeEventListener", "ended", player.endedCallback)
			player.source.Call("stop")
			player.source = js.Null()

			audioPlayersMutex.Lock()
			audioPlayers[name] = player
			audioPlayersMutex.Unlock()
		}

		// Stop the audio source if it is still playing
		// and the recursive limit has not been reached
		if IsPlaying(name) && recursive > 0 {
			stop(recursive - 1)
		}
	}

	// Stop the audio source (recursive limit: 10)
	stop(10)
}

// StopAudioSources is a function that stops all audio sources that match the selector.
func StopAudioSources(selector func(name string) bool) {
	audioPlayersMutex.RLock()

	var stopped []string
	for name, player := range audioPlayers {
		if selector(name) && player.source.Truthy() {
			stopped = append(stopped, name)
		}
	}

	audioPlayersMutex.RUnlock()

	for _, name := range stopped {
		StopAudio(name)
	}

	if Config.Control.Debug.Get() {
		Log(fmt.Sprintf("Stopped audio sources: %v", stopped))
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
	environ.Delete(key)
}
