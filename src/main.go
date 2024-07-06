//go:build js && wasm

package main

import (
	"fmt"

	config "github.com/sarumaj/edu-space-invaders/src/pkg/config"
	handler "github.com/sarumaj/edu-space-invaders/src/pkg/handler"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
)

// main is the entry point of the game.
func main() {
	game := handler.New()

	// For future use
	mode := config.Getenv("SPACE_INVADERS_MODE")
	config.Log(fmt.Sprintf("SPACE_INVADERS_MODE: %s", mode))

	config.Log("Generating enemies")
	game.GenerateEnemies(config.EnemiesCount, true)

	config.Log("Starting game loop")

	go game.Loop(func(e *enemy.Enemies) {
		for len(*e) < config.EnemiesCount {
			e.AppendNew("", false)
		}
	})

	if game.IsRunning() {
		game.SendMessage("Game started! Use ARROW KEYS (<, >) to move and SPACE to shoot.")
	} else {
		game.SendMessage("Let's begin! Press any key to start.")
	}

	game.Wait()
}
