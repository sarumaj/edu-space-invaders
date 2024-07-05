//go:build js && wasm

package config

import (
	"fmt"
	"syscall/js"
)

var (
	Canvas     = Doc.Call("getElementById", "gameCanvas")
	Console    = js.Global().Get("console")
	Ctx        = Canvas.Call("getContext", "2d")
	Doc        = js.Global().Get("document")
	Env        = js.Global().Get("go_env")
	Kubeconfig = js.Global().Get("go_kubeconfig")
	MessageBox = Doc.Call("getElementById", "message")
	Window     = js.Global().Get("window")
)

// Getenv is a function that returns the value of the environment variable key.
func Getenv(key string) string {
	got := Env.Get(key)
	if got.IsUndefined() {
		return ""
	}

	return got.String()
}

// Log is a function that logs a message.
func Log(msg string) {
	Console.Call("log", msg)
}

// LogError is a function that logs an error.
func LogError(err error) {
	if err != nil {
		Console.Call("error", err.Error())
	}
}

// ThrowError is a function that throws an error.
func ThrowError(err error) {
	if err != nil {
		js.Global().Call("eval", fmt.Sprintf("throw new Error('%s')", err.Error()))
	}
}
