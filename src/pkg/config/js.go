//go:build js && wasm

package config

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"slices"
	"strings"
	"sync"
	"syscall/js"
	"time"
)

const (
	originalWidth  = 760 // Original width of the drawable canvas area (px, after considering the padding and border of the surrounding containers)
	originalHeight = 570 // Original height of the drawable canvas area (px, after considering the padding and border of the surrounding containers)
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
	fpsDiv                 = document.Call("getElementById", "fps")
	invisibleCanvas        = document.Call("createElement", "canvas")
	invisibleCtx           = invisibleCanvas.Call("getContext", "2d")
	invisibleCanvasScrollY = 0.0
	messageBox             = document.Call("getElementById", messageBoxId)
	scoreBoard             []score
	scoreBoardMutex        = sync.RWMutex{}
	window                 = GlobalGet("window")
	windowLocation         = window.Get("location")
)

// audioPlayer represents an audio player.
type audioPlayer struct {
	endedCallback js.Func
	source        js.Value
	startTime     float64
}

// dimensions represents the dimensions of the document.
type dimensions struct {
	BoxWidth, BoxHeight                  float64
	BoxLeft, BoxTop, BoxRight, BoxBottom float64
	OriginalWidth, OriginalHeight        float64
	ScaleWidth, ScaleHeight              float64
}

// score represents a score.
type score struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Score     int       `json:"score"`
}

// init is a function that initializes the game interface.
func init() {
	// Set up the game interface
	setupAudioInterface()
	setupCanvasInterface()
	setupMessageBoxInterface()
	setupRefreshInterface()
	setupScoreBoard()

	// Detach the watchdogs
	checkHealth(1)
	envCallback(1)
}

// checkHealth is a function that checks the health of the game.
// The health check is performed every 10 seconds.
// The health check is performed by fetching the health endpoint.
// If the health check fails, the next health check is scheduled with exponential backoff.
func checkHealth(exponentialBackoff float64) {
	delayInMs := 10_000 * time.Millisecond

	go func() {
		GlobalCall("fetch", "health", MakeObject(map[string]any{
			"method":  http.MethodGet,
			"headers": MakeObject(map[string]any{"Accept": "application/json"}),
		})).Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
			if response := p[0]; !response.Get("ok").Bool() {
				return response.Call("text").Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
					return GlobalGet("Promise").Call("reject", GlobalGet("Error").New(p[0].String()))
				}))
			}

			time.AfterFunc(delayInMs, func() { checkHealth(exponentialBackoff) })
			return nil

		})).Call("catch", js.FuncOf(func(_ js.Value, p []js.Value) interface{} {
			LogError(fmt.Errorf("Error checking health: %s", p[0].String()))

			time.AfterFunc(delayInMs*time.Duration(exponentialBackoff), func() {
				checkHealth(exponentialBackoff * 2)
			})
			return nil

		}))
	}()
}

// envCallback is a function that fetches the environment variables.
func envCallback(exponentialBackoff float64) {
	delayInMs := 2_500 * time.Millisecond

	go func() {
		js.Global().Call("fetch", ".env", MakeObject(map[string]any{
			"method":  http.MethodGet,
			"headers": MakeObject(map[string]any{"Accept": "application/json"}),
		})).Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
			response := p[0]
			if !response.Get("ok").Bool() {
				return response.Call("text").Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
					return GlobalGet("Promise").Call("reject", GlobalGet("Error").New(p[0].String()))
				}))
			}

			return response.Call("json")

		})).Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
			envData := ConvertObjectToMap(p[0])
			prefix, _ := envData["_prefix"].(string)
			env := make(map[string]any)
			for key, value := range envData {
				if key != "_prefix" && len(prefix) > 0 && strings.HasPrefix(key, prefix) {
					env[key] = value
				}
			}

			Log(fmt.Sprintf("Retrieved environment variables: %#v", env))
			GlobalSet(goEnv, MakeObject(env))

			time.AfterFunc(delayInMs, func() {
				envCallback(exponentialBackoff)
			})

			return nil

		})).Call("catch", js.FuncOf(func(_ js.Value, p []js.Value) interface{} {
			LogError(fmt.Errorf("Error getting env: %s", p[0].String()))

			time.AfterFunc(delayInMs*time.Duration(exponentialBackoff), func() {
				envCallback(exponentialBackoff * 2)
			})

			return nil
		}))
	}()
}

// getAudioContext is a function that returns the audio context.
func getAudioContext() js.Value {
	ctx := NewInstance("AudioContext")
	if !ctx.Truthy() {
		ctx = NewInstance("webkitAudioContext")
	}
	return ctx
}

// scoreBoardSortFunc is a function that sorts the scores.
// The scores are sorted in descending order of the score.
// If the scores have the same score, they are ordered in ascending order of their timestamps.
func scoreBoardSortFunc(i, j score) int {
	if i.Score == j.Score {
		if !i.UpdatedAt.IsZero() && !j.UpdatedAt.IsZero() {
			// Sort in ascending order
			return j.UpdatedAt.Compare(i.UpdatedAt)
		}

		// Sort in ascending order
		return j.CreatedAt.Compare(i.CreatedAt)
	}

	// Sort in descending order
	return j.Score - i.Score
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

	GlobalSet("resize", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		if GlobalGet("drawFunc").Truthy() {
			GlobalGet("drawFunc").Invoke()
		}
		return nil
	}))

	window.Call("addEventListener", "resize", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		GlobalCall("requestAnimationFrame", GlobalGet("resize"))
		return nil
	}))

	GlobalCall("requestAnimationFrame", GlobalGet("resize"))
}

// setupMessageBoxInterface is a function that sets up the message box interface.
// The message box is scrollable only if the content inside the #message element can scroll.
// The touch events are prevented from propagating to the body when the message box is touched.
func setupMessageBoxInterface() {
	if !IsTouchDevice() {
		return
	}

	messageBox.Call("addEventListener", "touchstart", js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("stopPropagation") // Stop the touch event from propagating to the body
		return nil
	}))

	messageBox.Call("addEventListener", "touchmove", js.FuncOf(func(_ js.Value, p []js.Value) any {
		// Allow touch move events only if the content inside the #message element can scroll
		if messageBox.Get("scrollHeight").Float() > messageBox.Get("clientHeight").Float() {
			p[0].Call("stopPropagation") // Prevent body scroll when inside the messageBox
		}
		return nil
	}))
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

// setupScoreBoard is a function that sets up the score board.
func setupScoreBoard() {
	scoreBoardMutex.Lock()

	GlobalCall("fetch", "scores.db", MakeObject(map[string]any{
		"method":  http.MethodGet,
		"headers": MakeObject(map[string]any{"Content-Type": "application/json"}),
	})).Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
		if !p[0].Get("ok").Bool() {
			return p[0].Call("text").Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
				return GlobalGet("Promise").Call("reject", GlobalGet("Error").New(p[0].String()))
			}))
		}

		return p[0].Call("text")

	})).Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
		defer scoreBoardMutex.Unlock()

		if err := json.Unmarshal([]byte(strings.TrimPrefix(p[0].String(), "while(1);")), &scoreBoard); err != nil {
			return GlobalGet("Promise").Call("reject", GlobalGet("Error").New(err.Error()))
		}

		slices.SortStableFunc(scoreBoard, scoreBoardSortFunc)
		Log(fmt.Sprintf("Fetched score board: %#v", scoreBoard))

		return nil

	})).Call("catch", js.FuncOf(func(_ js.Value, p []js.Value) any {
		defer scoreBoardMutex.Unlock()

		LogError(fmt.Errorf("failed to load score board: %s", p[0].String()))
		return nil

	}))
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
		BoxLeft:        box.Get("left").Float(),
		BoxTop:         box.Get("top").Float(),
		BoxRight:       box.Get("right").Float(),
		BoxBottom:      box.Get("bottom").Float(),
		BoxWidth:       box.Get("width").Float(),
		BoxHeight:      box.Get("height").Float(),
		OriginalWidth:  originalWidth,
		OriginalHeight: originalHeight,
	}

	dim.ScaleWidth = dim.BoxWidth / dim.OriginalWidth
	dim.ScaleHeight = dim.BoxHeight / dim.OriginalHeight

	return dim
}

// ClearBackground is a function that clears the invisible document.
func ClearBackground() {
	invisibleCtx.Call("clearRect", 0, 0, invisibleCanvas.Get("width"), invisibleCanvas.Get("height"))
}

// ClearCanvas is a function that clears the document.
func ClearCanvas() {
	canvasObjectContext.Call("clearRect", 0, 0, canvasObject.Get("width"), canvasObject.Get("height"))
}

// ConvertArrayToSlice is a function that converts an array to a slice.
func ConvertArrayToSlice(array js.Value) []any {
	length := array.Length()
	result := make([]any, length)
	for i := 0; i < length; i++ {
		element := array.Index(i)
		switch element.Type() {
		case js.TypeObject:
			if element.InstanceOf(js.Global().Get("Array")) {
				result[i] = ConvertArrayToSlice(element)
			} else {
				result[i] = ConvertObjectToMap(element)
			}
		case js.TypeString:
			result[i] = element.String()
		case js.TypeNumber:
			result[i] = element.Float()
		case js.TypeBoolean:
			result[i] = element.Bool()
		case js.TypeNull, js.TypeUndefined:
			result[i] = nil
		default:
			result[i] = element
		}
	}
	return result
}

// ConvertObjectToMap is a function that converts an object to a map.
func ConvertObjectToMap(obj js.Value) map[string]any {
	result := make(map[string]any)
	keys := GlobalGet("Object").Call("keys", obj)
	for i := 0; i < keys.Length(); i++ {
		key := keys.Index(i).String()
		value := obj.Get(key)

		switch value.Type() {
		case js.TypeObject:
			if value.InstanceOf(GlobalGet("Array")) {
				result[key] = ConvertArrayToSlice(value)
			} else {
				result[key] = ConvertObjectToMap(value)
			}
		case js.TypeString:
			result[key] = value.String()
		case js.TypeNumber:
			result[key] = value.Float()
		case js.TypeBoolean:
			result[key] = value.Bool()
		case js.TypeNull, js.TypeUndefined:
			result[key] = nil
		default:
			result[key] = value
		}
	}

	return result
}

// DrawBackground is a function that draws the background of the document.
// The background is drawn with the specified speed.
func DrawBackground(speed float64) {
	if !*Config.Control.BackgroundAnimationEnabled {
		canvasObjectContext.Call("drawImage", invisibleCanvas, 0, 0)
		return
	}

	canvasDimensions := CanvasBoundingBox()

	// Apply the speed
	invisibleCanvasScrollY += speed
	if invisibleCanvasScrollY/canvasDimensions.OriginalHeight > 1 {
		invisibleCanvasScrollY -= canvasDimensions.OriginalHeight
	}

	canvasObjectContext.Call("drawImage", invisibleCanvas, 0, invisibleCanvasScrollY)
	canvasObjectContext.Call("drawImage", invisibleCanvas, 0, invisibleCanvasScrollY-canvasDimensions.OriginalHeight)
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

// DrawPlanetEarth is a function that draws Earth on the document.
func DrawPlanetEarth(coords [2]float64, radius float64) {
	cx, cy := coords[0], coords[1]

	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("arc", cx, cy, radius, 0, 2*math.Pi, false)
	canvasObjectContext.Call("closePath")

	// Use a blue gradient to represent the oceans
	gradient := canvasObjectContext.Call("createRadialGradient", cx, cy, radius*0.2, cx, cy, radius)
	gradient.Call("addColorStop", 0, "#00BFFF") // Light blue at the center
	gradient.Call("addColorStop", 1, "#1E90FF") // Darker blue at the edges

	canvasObjectContext.Set("fillStyle", gradient)
	canvasObjectContext.Call("fill")

	// Add some green/brown patches for land
	landColors := []string{"#228B22", "#8B4513"} // Green for forests, brown for deserts/mountains
	landPatches := [][3]float64{
		{cx - radius*0.2, cy - radius*0.3, radius * 0.4},
		{cx + radius*0.1, cy + radius*0.2, radius * 0.3},
	}

	for i, patch := range landPatches {
		canvasObjectContext.Call("beginPath")
		canvasObjectContext.Call("arc", patch[0], patch[1], patch[2], 0, 2*math.Pi, false)
		canvasObjectContext.Call("closePath")
		canvasObjectContext.Set("fillStyle", landColors[i%len(landColors)])
		canvasObjectContext.Call("fill")
	}

	// Add white patches for clouds
	clouds := [][3]float64{
		{cx - radius*0.4, cy - radius*0.1, radius * 0.6},
		{cx + radius*0.3, cy + radius*0.2, radius * 0.5},
	}

	for _, cloud := range clouds {
		canvasObjectContext.Call("beginPath")
		canvasObjectContext.Call("arc", cloud[0], cloud[1], cloud[2], 0, 2*math.Pi, false)
		canvasObjectContext.Call("closePath")
		canvasObjectContext.Set("fillStyle", "rgba(255, 255, 255, 0.6)")
		canvasObjectContext.Call("fill")
	}
}

// DrawPlanetJupiter is a function that draws Jupiter on the document.
func DrawPlanetJupiter(coords [2]float64, radius float64) {
	cx, cy := coords[0], coords[1] // Center position

	// Draw the main body of Jupiter (a sphere)
	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("arc", cx, cy, radius, 0, 2*math.Pi, false)
	canvasObjectContext.Call("closePath")

	// Create a radial gradient to simulate the planet's lighting
	gradient := canvasObjectContext.Call("createRadialGradient", cx, cy, radius*0.3, cx, cy, radius)
	gradient.Call("addColorStop", 0, "#FFF4C3") // Lighter central color
	gradient.Call("addColorStop", 1, "#E2B56D") // Darker edge color

	canvasObjectContext.Set("fillStyle", gradient)
	canvasObjectContext.Call("fill")

	// Clip the drawing area to the circle of the planet
	canvasObjectContext.Call("save") // Save the current drawing state
	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("arc", cx, cy, radius, 0, 2*math.Pi, false)
	canvasObjectContext.Call("closePath")
	canvasObjectContext.Call("clip") // Clip to the planet's circle

	// Add bands to simulate Jupiter's gas bands
	bandColors := []string{
		"#F3D29E", "#EAB277", "#E5AA66", "#DF9A55", "#D98A44",
		"#D07A33", "#C96922", "#C25811", "#BB4700",
	}

	numBands := len(bandColors)
	bandHeight := (radius * 2) / float64(numBands)

	for i, color := range bandColors {
		y := cy - radius + float64(i)*bandHeight

		canvasObjectContext.Call("beginPath")
		canvasObjectContext.Call("rect", cx-radius, y, radius*2, bandHeight)
		canvasObjectContext.Set("fillStyle", color)
		canvasObjectContext.Call("fill")
		canvasObjectContext.Call("closePath")
	}

	canvasObjectContext.Call("restore") // Restore the drawing state to remove the clipping

	// Add the Great Red Spot (simply a circle here)
	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("arc", cx+radius*0.5, cy+radius*0.4, radius*0.3, 0, 2*math.Pi, false)
	canvasObjectContext.Call("closePath")
	canvasObjectContext.Set("fillStyle", "#D66B41")
	canvasObjectContext.Call("fill")
}

// DrawPlanetMars is a function that draws Mars on the document.
func DrawPlanetMars(coords [2]float64, radius float64) {
	cx, cy := coords[0], coords[1]

	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("arc", cx, cy, radius, 0, 2*math.Pi, false)
	canvasObjectContext.Call("closePath")

	// Use a reddish color to represent Mars
	gradient := canvasObjectContext.Call("createRadialGradient", cx, cy, radius*0.2, cx, cy, radius)
	gradient.Call("addColorStop", 0, "#FF6347") // Lighter red at the center
	gradient.Call("addColorStop", 1, "#B22222") // Darker red at the edges

	canvasObjectContext.Set("fillStyle", gradient)
	canvasObjectContext.Call("fill")

	// Darker patches representing regions like Syrtis Major
	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("arc", cx-radius*0.2, cy-radius*0.1, radius*0.3, 0, 2*math.Pi, false)
	canvasObjectContext.Call("closePath")
	canvasObjectContext.Set("fillStyle", "#8B0000")
	canvasObjectContext.Call("fill")

	// Draw crater-like features on Mars
	radius *= 0.8
	for _, crater := range [][3]float64{
		{cx - radius*0.3, cy - radius*0.3, radius * 0.1},
		{cx + radius*0.2, cy - radius*0.1, radius * 0.15},
		{cx, cy + radius*0.3, radius * 0.05},
	} {
		canvasObjectContext.Call("beginPath")
		canvasObjectContext.Call("arc", crater[0], crater[1], crater[2], 0, 2*math.Pi, false)
		canvasObjectContext.Call("closePath")
		canvasObjectContext.Set("fillStyle", "#A0A0A0")
		canvasObjectContext.Call("fill")
	}
}

// DrawPlanetMercury is a function that draws Mercury on the document.
func DrawPlanetMercury(coords [2]float64, radius float64) {
	cx, cy := coords[0], coords[1]

	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("arc", cx, cy, radius, 0, 2*math.Pi, false)
	canvasObjectContext.Call("closePath")

	// Use a simple gray color to represent Mercury
	canvasObjectContext.Set("fillStyle", "#B1B1B1")
	canvasObjectContext.Call("fill")

	// Draw crater-like features on Mercury
	for _, crater := range [][3]float64{
		{cx - radius*0.3, cy - radius*0.3, radius * 0.1},
		{cx + radius*0.2, cy - radius*0.1, radius * 0.15},
		{cx, cy + radius*0.3, radius * 0.05},
	} {
		canvasObjectContext.Call("beginPath")
		canvasObjectContext.Call("arc", crater[0], crater[1], crater[2], 0, 2*math.Pi, false)
		canvasObjectContext.Call("closePath")
		canvasObjectContext.Set("fillStyle", "#A0A0A0")
		canvasObjectContext.Call("fill")
	}
}

// DrawPlanetNeptune is a function that draws Neptune on the document.
func DrawPlanetNeptune(coords [2]float64, radius float64) {
	cx, cy := coords[0], coords[1]

	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("arc", cx, cy, radius, 0, 2*math.Pi, false)
	canvasObjectContext.Call("closePath")

	// Deep blue color for Neptune
	gradient := canvasObjectContext.Call("createRadialGradient", cx, cy, radius*0.3, cx, cy, radius)
	gradient.Call("addColorStop", 0, "#4169E1")
	gradient.Call("addColorStop", 1, "#00008B")

	canvasObjectContext.Set("fillStyle", gradient)
	canvasObjectContext.Call("fill")
}

// DrawPlanetSaturn is a function that draws Saturn on the document.
func DrawPlanetSaturn(coords [2]float64, radius float64) {
	cx, cy := coords[0], coords[1]

	// Draw Saturn's body
	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("arc", cx, cy, radius, 0, 2*math.Pi, false)
	canvasObjectContext.Call("closePath")

	// Pale yellow color for Saturn
	gradient := canvasObjectContext.Call("createRadialGradient", cx, cy, radius*0.3, cx, cy, radius)
	gradient.Call("addColorStop", 0, "#F5DEB3")
	gradient.Call("addColorStop", 1, "#DAA520")

	canvasObjectContext.Set("fillStyle", gradient)
	canvasObjectContext.Call("fill")

	// Draw Saturn's rings
	innerRadius := radius * 1.2
	outerRadius := radius * 2

	// Draw the rings as ellipses
	for i := 0; i < 3; i++ {
		canvasObjectContext.Call("beginPath")
		canvasObjectContext.Call("ellipse", cx, cy, outerRadius, innerRadius*0.6, 0, 0, 2*math.Pi)
		canvasObjectContext.Call("closePath")
		canvasObjectContext.Set("fillStyle", "rgba(210, 180, 140, 0.6)")
		canvasObjectContext.Call("fill")

		innerRadius += radius * 0.1
		outerRadius += radius * 0.2
	}
}

// DrawPlanetUranus is a function that draws Uranus on the document.
func DrawPlanetUranus(coords [2]float64, radius float64) {
	cx, cy := coords[0], coords[1]

	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("arc", cx, cy, radius, 0, 2*math.Pi, false)
	canvasObjectContext.Call("closePath")

	// Cyan color for Uranus
	gradient := canvasObjectContext.Call("createRadialGradient", cx, cy, radius*0.3, cx, cy, radius)
	gradient.Call("addColorStop", 0, "#AFEEEE")
	gradient.Call("addColorStop", 1, "#5F9EA0")

	canvasObjectContext.Set("fillStyle", gradient)
	canvasObjectContext.Call("fill")
}

// DrawPlanetVenus is a function that draws Venus on the document.
func DrawPlanetVenus(coords [2]float64, radius float64) {
	cx, cy := coords[0], coords[1]

	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("arc", cx, cy, radius, 0, 2*math.Pi, false)
	canvasObjectContext.Call("closePath")

	// Use a gradient to simulate the thick atmosphere
	gradient := canvasObjectContext.Call("createRadialGradient", cx, cy, radius*0.2, cx, cy, radius)
	gradient.Call("addColorStop", 0, "#F0E68C") // Lighter yellow at the center
	gradient.Call("addColorStop", 1, "#E3C16F") // Darker yellow at the edges

	canvasObjectContext.Set("fillStyle", gradient)
	canvasObjectContext.Call("fill")

	// Add some cloud patterns or swirls
	canvasObjectContext.Call("beginPath")
	canvasObjectContext.Call("arc", cx-radius*0.2, cy-radius*0.1, radius*0.6, 0, math.Pi*2, false)
	canvasObjectContext.Call("arc", cx+radius*0.2, cy+radius*0.1, radius*0.7, 0, math.Pi*2, false)
	canvasObjectContext.Call("closePath")
	canvasObjectContext.Set("fillStyle", "rgba(255, 255, 255, 0.1)")
	canvasObjectContext.Call("fill")
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
func DrawSpaceship(coors [2]float64, size [2]float64, faceUp bool, color, label string) {
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

	// Draw the label above or below the spaceship
	if label != "" {
		canvasObjectContext.Set("font", "16px Arial") // Set font

		// Shorten the label if it is too long
		if len(label) > Config.Spaceship.MaximumLabelLength {
			label = fmt.Sprintf("%s...", label[:Config.Spaceship.MaximumLabelLength-3])
		}

		// Measure the width of the label text
		textMetrics := canvasObjectContext.Call("measureText", label)
		labelWidth := textMetrics.Get("width").Float()

		labelX := x + (width-labelWidth)/2 // Center the label horizontally

		var labelY float64
		if faceUp {
			labelY = y + height + 10 // Below the spaceship
		} else {
			labelY = y - 5 // Above the spaceship
		}

		// Draw the label text with a black stroke and fill
		canvasObjectContext.Set("strokeStyle", "black")
		canvasObjectContext.Call("strokeText", label, labelX, labelY)

		canvasObjectContext.Set("fillStyle", color) // Set text color
		canvasObjectContext.Call("fillText", label, labelX, labelY)
	}
}

// DrawStar draws a star on the invisible canvas to be used as a background on the visible one.
// The star is drawn at the specified position (cx, cy) with the specified number of spikes.
// The outer radius and inner radius of the star are specified.
// The star is filled with the specified color.
func DrawStar(coords [2]float64, spikes int, radius, innerRadius float64, color string, brightness float64) {
	cx, cy := coords[0], coords[1] // Center position

	// Calculate the positions of the star
	var positions [][2]float64
	for i, r := 0, 0.0; i < 2*spikes; i++ {
		if i%2 == 0 {
			r = radius
		} else {
			r = innerRadius
		}

		angle := float64(i) * math.Pi / float64(spikes)
		x := cx + math.Cos(angle)*r
		y := cy + math.Sin(angle)*r
		positions = append(positions, [2]float64{x, y})
	}

	// Draw the star
	// Darken the color based on the brightness
	for _, c := range []string{color, fmt.Sprintf("rgba(0, 0, 0, %.2f)", 1-brightness)} {
		invisibleCtx.Call("beginPath")
		invisibleCtx.Set("fillStyle", c)
		invisibleCtx.Call("moveTo", positions[0][0], positions[0][1])
		for i := 1; i < len(positions); i++ {
			invisibleCtx.Call("lineTo", positions[i][0], positions[i][1])
		}
		invisibleCtx.Call("lineTo", positions[0][0], positions[0][1]) // Close the star
		invisibleCtx.Call("closePath")
		invisibleCtx.Call("fill")
	}
}

// Getenv is a function that returns the value of the environment variable key.
func Getenv(key string) string {
	got := GlobalGet(goEnv).Get(key)
	if !got.Truthy() {
		return ""
	}

	return got.String()
}

// GetScores is a function that returns the scores.
func GetScores(top int) (scores []score) {
	scoreBoardMutex.RLock()
	defer scoreBoardMutex.RUnlock()

	for i := 0; i < top && i < len(scoreBoard); i++ {
		scores = append(scores, scoreBoard[i])
	}

	return
}

// GlobalCall is a function that calls the global function name with the specified arguments.
func GlobalCall(name string, args ...any) js.Value {
	return js.Global().Call(name, args...)
}

// GlobalGet is a function that returns the global value of key.
func GlobalGet(key string) js.Value {
	return js.Global().Get(key)
}

// GlobalSet is a function that sets the global value of key to value.
func GlobalSet(key string, value any) {
	switch value := value.(type) {
	case js.Value:
		js.Global().Set(key, value)
	default:
		js.Global().Set(key, js.ValueOf(value))
	}
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

	// Reinitialize the audio context if it is not initialized
	if !audioCtx.Truthy() {
		audioCtx = getAudioContext()
		if !audioCtx.Truthy() {
			LogError(fmt.Errorf("failed to initialize audio context"))
			return
		}
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

// SaveScores is a function that saves the score board persistently.
func SaveScores() {
	scoreBoardMutex.RLock()
	serialized, err := json.Marshal(scoreBoard)
	scoreBoardMutex.RUnlock()

	if err != nil {
		LogError(fmt.Errorf("failed to serialize score board: %v", err))
		return
	}

	// Save the score board
	SendMessage(Config.MessageBox.Messages.WaitForScoreBoardUpdate, false)
	scoreBoardMutex.Lock()

	GlobalCall("fetch", "scores.db", MakeObject(map[string]any{
		"method":  http.MethodPut,
		"headers": MakeObject(map[string]any{"Content-Type": "application/json"}),
		"body":    string(serialized),
	})).Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
		defer scoreBoardMutex.Unlock()

		if !p[0].Get("ok").Bool() {
			return p[0].Call("text").Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
				LogError(fmt.Errorf("server responded with error: %s", p[0].String()))
				return nil
			}))
		}

		// Send success message
		SendMessage(Config.MessageBox.Messages.ScoreBoardUpdated, false)
		return nil
	})).Call("catch", js.FuncOf(func(_ js.Value, p []js.Value) any {
		defer scoreBoardMutex.Unlock()

		LogError(fmt.Errorf("failed to save score board: %s", p[0].String()))
		return nil
	}))
}

// SendMessage sends a message to the message box.
func SendMessage(msg string, reset bool) {
	lines := []string{msg}
	if !reset {
		content := messageBox.Get("innerHTML").String()
		lines = append(strings.Split(content, "<br>"), lines...)
		if len(lines) > Config.MessageBox.BufferSize {
			lines = lines[len(lines)-Config.MessageBox.BufferSize:]
		}
	}

	messageBox.Set("innerHTML", strings.Join(lines, "<br>"))
	messageBox.Set("scrollTop", messageBox.Get("scrollHeight"))
}

// Setenv is a function that sets the environment variable key to value.
func Setenv(key, value string) {
	environ := GlobalGet(goEnv)
	environ.Set(key, value)
	GlobalSet(goEnv, environ)
}

// SetScore is a function that sets the score.
func SetScore(name string, newScore int) (rank int) {
	scoreBoardMutex.Lock()

	// Update the score board
	var exists bool
	for i, s := range scoreBoard {
		if s.Name == name {
			if newScore <= s.Score {
				scoreBoardMutex.Unlock()
				return len(scoreBoard) + 1
			}

			scoreBoard[i].Score = newScore
			exists = true
			break
		}
	}

	// Add the score if it does not exist
	if !exists {
		scoreBoard = append(scoreBoard, score{Name: name, Score: newScore})
	}

	// Sort the score board
	slices.SortStableFunc(scoreBoard, scoreBoardSortFunc)
	scoreBoardMutex.Unlock()

	// Calculate the rank of the new score
	scoreBoardMutex.RLock()
	for i, s := range scoreBoard {
		if s.Name == name && s.Score == newScore {
			rank = i + 1
			break
		}
	}
	scoreBoardMutex.RUnlock()

	// Save the score board asynchronously
	go SaveScores()

	return
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
	environ := GlobalGet(goEnv)
	environ.Delete(key)
	GlobalSet(goEnv, environ)
}

// UpdateFPS is a function that updates the frames per second.
func UpdateFPS(fps float64) {
	fpsDiv.Set("innerHTML", fmt.Sprintf(fpsDiv.Call("getAttribute", "data-format").String(), fps))
}

// UpdateMessage is a function that updates the last message in the message box.
func UpdateMessage(msg string) {
	lines := strings.Split(messageBox.Get("innerHTML").String(), "<br>")
	lines[len(lines)-1] = msg
	messageBox.Set("innerHTML", strings.Join(lines, "<br>"))
}
