package handler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// handler is the game handler.
type handler struct {
	ctx          context.Context    // ctx is an abortable context of the handler
	cancel       context.CancelFunc // cancel is the cancel function of the handler
	once         sync.Once          // once is meant to register the keydown event only once
	spaceship    *objects.Spaceship // spaceship is the player's spaceship
	enemies      objects.Enemies    // enemies is the list of enemies
	keydownEvent chan string        // keydownEvent is the channel for keydown events
}

// checkCollisions checks if the spaceship has collided with an enemy.
// If the spaceship has collided with an enemy, it applies the necessary
// penalties and upgrades.
// If the spaceship has collided with a goodie, it upgrades the spaceship.
// If the spaceship has collided with a normal enemy, it applies default penalty.
// If the spaceship has collided with a berserker, it applies the berserker penalty.
// If the spaceship has collided with an annihilator, it applies the annihilator penalty.
// If the spaceship is boosted, it destroys the enemy.
// It checks if the bullets have hit an enemy.
// If the bullets have hit an enemy, it applies the necessary damage.
// If the enemy has no health points, it upgrades the spaceship.
func (h *handler) checkCollisions() {
	for j, e := range h.enemies {
		if e.Level.HitPoints > 0 && h.spaceship.DetectCollision(e) {
			penalty := config.SpaceshipDefaultPenalty
			switch e.Type {
			case objects.Goodie:
				h.enemies[j].Level.HitPoints = 0
				h.spaceship.Level.Up()
				h.spaceship.ChangeState(objects.Boosted)
				h.SendMessage(fmt.Sprintf("You got a goodie, your spaceship has been upgraded to level %d", h.spaceship.Level.ID))
				return

			case objects.Berserker:
				penalty = config.SpaceshipBerserkPenalty

			case objects.Annihilator:
				penalty = config.SpaceshipAnnihilatorPenalty

			}

			h.enemies[j].Level.HitPoints = 0
			if h.spaceship.State == objects.Boosted {
				h.SendMessage(fmt.Sprintf("You destroyed %s", e.Name))
				return
			}

			if h.spaceship.Level.ID > 1 {
				h.spaceship.Penalize(penalty)
				h.spaceship.ChangeState(objects.Damaged)
				h.SendMessage(fmt.Sprintf("You were hit, your spaceship has been downgraded to level %d", h.spaceship.Level.ID))
				return
			}

			h.spaceship.ChangeState(objects.Damaged)
			h.SendMessage("You were killed, R.I.P.")
			h.cancel()
			return
		}

		for i, b := range h.spaceship.Bullets {
			if e.Level.HitPoints > 0 && e.Type != objects.Goodie && b.HasHit(e) {
				h.SendMessage(fmt.Sprintf("You dealt %d of damage to %q", h.enemies[j].Hit(b.Damage), e.Name))
				h.spaceship.Bullets[i].Exhaust()
				if h.enemies[j].Level.HitPoints <= 0 {
					h.SendMessage(fmt.Sprintf("You killed %q, your spaceship has been upgraded to level %d", e.Name, h.spaceship.Level.ID))
					h.spaceship.Level.Up()
				}
			}
		}
	}
}

// handlerKeydown handles the keydown event.
// It moves the spaceship to the left when the left arrow key is pressed.
// It moves the spaceship to the right when the right arrow key is pressed.
// It fires bullets when the space key is pressed.
func (h *handler) handleKeydown(key string) {
	switch key {
	case "ArrowLeft":
		h.spaceship.MoveLeft()

	case "ArrowRight":
		h.spaceship.MoveRight()

	case "Space":
		h.spaceship.Fire()
	}
}

// refresh refreshes the game state.
// It updates the bullets of the spaceship.
// It updates the enemies.
// It updates the state of the spaceship.
// It checks the collisions.
func (h *handler) refresh(watch func(e *objects.Enemies)) {
	h.spaceship.Bullets.Update()
	h.enemies.Update(h.spaceship.Position, watch)
	h.spaceship.UpdateState()
	h.checkCollisions()
}

// GenerateEnemy generates a new enemy with the specified name and random Y position.
func (h *handler) GenerateEnemy(name string, randomY bool) { h.enemies.AppendNew(name, randomY) }

// GenerateEnemies generates the specified number of enemies with random Y position.
func (h *handler) GenerateEnemies(num int, randomY bool) {
	for i := 0; i < num; i++ {
		h.enemies.AppendNew("", randomY)
	}
}

// Loop starts the game loop.
// It refreshes the game state, renders the game, and handles the keydown events.
// It should be called in a separate goroutine.
func (h *handler) Loop(watch func(e *objects.Enemies)) {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	for {
		select {
		case <-h.ctx.Done():
			return

		case <-ticker.C:
			h.refresh(watch)
			h.render()

		case key := <-h.keydownEvent:
			h.handleKeydown(key)

		}
	}
}

// Wait waits for the handler to finish.
func (h *handler) Wait() { <-h.ctx.Done() }

// New creates a new handler.
// It creates a new spaceship and registers the keydown event.
func New() *handler {
	h := &handler{
		keydownEvent: make(chan string),
		spaceship:    objects.Embark(),
	}

	h.ctx, h.cancel = context.WithCancel(context.Background())
	h.registerKeydownEvent()

	return h
}
