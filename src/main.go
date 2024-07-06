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

	switch {

	case game.IsRunning():
		if config.IsTouchDevice() {
			game.SendMessage(config.MessageGameStartedTouchDevice)
		} else {
			game.SendMessage(config.MessageGameStartedNoTouchDevice)
		}

	default:
		if config.IsTouchDevice() {
			game.SendMessage(config.MessageHowToStartTouchDevice)
		} else {
			game.SendMessage(config.MessageHowToStartNoTouchDevice)
		}

	}

	game.Wait()
}
