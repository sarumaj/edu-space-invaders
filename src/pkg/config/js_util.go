//go:build js && wasm

package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"syscall/js"
	"time"
)

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
	SendMessage(Execute(Config.MessageBox.Messages.WaitForScoreBoardUpdate), false, false)
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
		SendMessage(Execute(Config.MessageBox.Messages.ScoreBoardUpdated), false, false)
		return nil
	})).Call("catch", js.FuncOf(func(_ js.Value, p []js.Value) any {
		defer scoreBoardMutex.Unlock()

		LogError(fmt.Errorf("failed to save score board: %s", p[0].String()))
		return nil
	}))
}

// SendInfoMessage sends a message to the message box.
func SendMessage(msg string, reset, event logEvent) {
	msg = fmt.Sprintf(`<div>%s</div>`, msg)

	channel := event.Channel()
	channelBtn := event.ChannelButton()

	if reset {
		// Reset the content, keeping only the new message
		channel.Set("innerHTML", msg)
	} else {
		// Append the new message to the DOM
		channel.Call("insertAdjacentHTML", "beforeend", msg)

		// Limit the number of messages in the DOM
		for channel.Get("children").Length() > Config.MessageBox.ChannelBufferSize {
			channel.Call("removeChild", channel.Get("firstChild"))
		}
	}

	// Scroll to the beginning of the newly added message
	channel.Get("lastChild").Call("scrollIntoView", MakeObject(map[string]any{
		"block":    "start",
		"behavior": "smooth",
	}))

	if channelBtn.Get("classList").Call("contains", tabActiveClass).Bool() {
		return // Already active, nothing to do
	}

	// Flash the channel
	channelBtn.Get("classList").Call("add", tabFlashClass)
	channelBtn.Call("addEventListener", "animationend", js.FuncOf(func(this js.Value, _ []js.Value) any {
		this.Get("classList").Call("remove", tabFlashClass)
		if !event { // If not event log, activate the tab
			this.Get("classList").Call("add", tabActiveClass)
			this.Call("click")
		}
		return nil
	}), js.ValueOf(map[string]any{"once": true}))
}

// SendMessageThrottled sends a message to the message box with a cooldown.
func SendMessageThrottled(msg string, reset, event logEvent, cooldown time.Duration) {
	if !lastLogSentTime.IsZero() && time.Since(lastLogSentTime) < cooldown {
		return
	}

	SendMessage(msg, reset, event)
	lastLogSentTime = time.Now()
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
