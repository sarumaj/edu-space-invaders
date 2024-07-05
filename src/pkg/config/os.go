//go:build !js || !wasm

package config

import (
	"log"
	"os"
)

// Getenv is a function that returns the value of the environment variable key.
func Getenv(key string) string {
	return os.Getenv(key)
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

// ThrowError is a function that throws an error.
func ThrowError(err error) {
	if err != nil {
		panic(err)
	}
}
