//go:build js && wasm

package main

import (
	"fmt"

	config "github.com/sarumaj/edu-space-invaders/src/pkg/config"
	handler "github.com/sarumaj/edu-space-invaders/src/pkg/handler"
)

// main is the entry point of the game.
func main() {
	game := handler.New()

	// For future use
	mode := config.Getenv("SPACE_INVADERS_MODE")
	config.Log(fmt.Sprintf("SPACE_INVADERS_MODE: %s", mode))

	for {
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
