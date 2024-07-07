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
	game.GenerateEnemies(config.Config.Enemy.Count, true)

	config.Log("Starting game loop")

	go game.Loop(func(e *enemy.Enemies) {
		for len(*e) < config.Config.Enemy.Count {
			e.AppendNew("", false)
		}
	})

	switch {

	case game.IsRunning():
		if config.IsTouchDevice() {
			game.SendMessage(config.Config.Messages.GameStartedTouchDevice)
		} else {
			game.SendMessage(config.Config.Messages.GameStartedNoTouchDevice)
		}

	default:
		if config.IsTouchDevice() {
			game.SendMessage(config.Config.Messages.HowToStartTouchDevice)
		} else {
			game.SendMessage(config.Config.Messages.HowToStartNoTouchDevice)
		}

	}

	game.Wait()
}
