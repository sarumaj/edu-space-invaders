//go:build js && wasm

package main

import (
	config "github.com/sarumaj/edu-space-invaders/src/pkg/config"
	handler "github.com/sarumaj/edu-space-invaders/src/pkg/handler"
)

// ApiKey is the API key for the game.
// It is set by the build script.
// It is used to communicate with the game server.
var ApiKey string

// main is the entry point of the game.
func main() {
	if ApiKey == "" {
		panic("AppKey is not set")
	}

	config.GlobalSet("apiKey", ApiKey)

	for game := handler.New(); ; {
		config.Log("Generating enemies")
		game.GenerateEnemies(config.Config.Enemy.Count, true)

		config.Log("Starting the game loop")
		go game.Loop()

		config.Log("Awaiting the end of the game")
		game.Await()

		config.Log("Restarting the game")
		game.Restart()
	}
}
