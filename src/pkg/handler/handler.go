package handler

import (
	"context"
	"sync"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/spaceship"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/star"
)

// handler is the game handler.
type handler struct {
	ctx          context.Context      // ctx is an abortable context of the handler
	cancel       context.CancelFunc   // cancel is the cancel function of the handler
	enemies      enemy.Enemies        // enemies is the list of enemies
	keydownEvent chan keyBinding      // keydownEvent is the channel for keydown events
	keyupEvent   chan keyBinding      // keyupEvent is the channel for keyup events
	keysHeld     map[keyBinding]bool  // keysHeld is the map of keys held
	once         sync.Once            // once is meant to register the keydown event only once
	spaceship    *spaceship.Spaceship // spaceship is the player's spaceship
	stars        star.Stars           // stars is the list of stars
	touchEvent   chan touchEvent      // touchEvent is the channel for touch events
}

// checkCollisions checks if the spaceship has collided with an enemy.
// If the spaceship has collided with an enemy, it applies the necessary
// penalties and upgrades.
// If the spaceship has collided with a goodie, it upgrades the spaceship.
// If the spaceship has collided with a freezer, it freezes the spaceship.
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
			// If the spaceship is boosted, destroy the enemy.
			if h.spaceship.State == spaceship.Boosted {
				h.enemies[j].Destroy()
				return
			}

			// Handle collisions with normal, berserker and annihilator enemies.
			penalty := config.Config.Spaceship.DefaultPenalty
			switch e.Type {
			case enemy.Berserker:
				penalty = config.Config.Spaceship.BerserkPenalty

			case enemy.Annihilator:
				penalty = config.Config.Spaceship.AnnihilatorPenalty
			}

			if h.spaceship.State == spaceship.Frozen {
				h.enemies[j].Destroy()
				h.spaceship.Penalize(penalty)
				config.SendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipDowngradedByEnemy, config.Template{
					"SpaceshipLevel": h.spaceship.Level.Progress,
				}))
				if h.spaceship.IsDestroyed() {
					config.SendMessage(config.Execute(config.Config.Messages.Templates.GameOver, config.Template{
						"HighScore": h.spaceship.HighScore,
					}))
					h.cancel()
				}
				return
			}

			switch e.Type {
			case enemy.Goodie: // If the spaceship has collided with a goodie, upgrade the spaceship.
				h.enemies[j].Destroy()
				h.spaceship.Level.Up()
				h.spaceship.ChangeState(spaceship.Boosted)
				config.SendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipUpgradedByGoodie, config.Template{
					"SpaceshipLevel": h.spaceship.Level.Progress,
					"BoostDuration":  config.Config.Spaceship.BoostDuration,
				}))
				return

			case enemy.Freezer: // If the spaceship has collided with a freezer, freeze the spaceship.
				h.enemies[j].Destroy()
				h.spaceship.ChangeState(spaceship.Frozen)
				h.spaceship.Penalize(config.Config.Spaceship.FreezerPenalty)
				config.SendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipFrozen, config.Template{
					"SpaceshipLevel": h.spaceship.Level.Progress,
					"FreezeDuration": config.Config.Spaceship.FreezeDuration,
				}))
				return

			}

			h.enemies[j].Destroy()
			h.spaceship.Penalize(penalty)
			h.spaceship.ChangeState(spaceship.Damaged)
			config.SendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipDowngradedByEnemy, config.Template{
				"SpaceshipLevel": h.spaceship.Level.Progress,
			}))
			if h.spaceship.IsDestroyed() {
				config.SendMessage(config.Execute(config.Config.Messages.Templates.GameOver, config.Template{
					"HighScore": h.spaceship.HighScore,
				}))
				h.cancel()
				return
			}
		}

		for i, b := range h.spaceship.Bullets {
			switch {
			case
				e.Level.HitPoints <= 0,
				e.Type == enemy.Goodie,
				e.Type == enemy.Freezer,
				!b.HasHit(e):

			default: // The bullet has hit the enemy.
				h.spaceship.Bullets[i].Exhaust()
				h.enemies[j].Hit(b.Damage)

				// If the enemy has no health points, upgrade the spaceship.
				if h.enemies[j].IsDestroyed() {
					h.spaceship.Level.Up()
					h.spaceship.UpdateHighScore()
					config.SendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipUpgradedByEnemyKill, config.Template{
						"EnemyName":      e.Name,
						"SpaceshipLevel": h.spaceship.Level.Progress,
					}))
				}

				// If the progress is a multiple of the enemy count progress step,
				// generate a new enemy.
				if h.spaceship.Level.Progress%config.Config.Enemy.CountProgressStep == 0 &&
					len(h.enemies) < config.Config.Enemy.MaximumCount {

					h.GenerateEnemy("", false)
				}
			}

		}
	}
}

// handlerKeydown handles the keydown event.
// It sets the running state to true when any key is pressed.
// It handles the keydown event based on the key.
// It moves the spaceship to the left when the left arrow key is pressed.
// It moves the spaceship to the right when the right arrow key is pressed.
// If space key is pressed, it is observed in the keysHeld map.
func (h *handler) handleKeydown(key keyBinding) {
	if h.start() {
		return
	}

	switch key {
	case ArrowDown, ArrowLeft, ArrowRight, ArrowUp, Space:
		h.keysHeld[key] = true

	case Pause:
		h.pause()

	}
}

// handleKeyHold handles the key hold event.
// It fires bullets when the space key is held.
func (h *handler) handleKeyhold() {
	for key := range h.keysHeld {
		if !h.keysHeld[key] {
			continue
		}

		switch key {
		case ArrowDown:
			h.spaceship.MoveDown()

		case ArrowLeft:
			h.spaceship.MoveLeft()

		case ArrowRight:
			h.spaceship.MoveRight()

		case ArrowUp:
			h.spaceship.MoveUp()

		case Space:
			h.spaceship.Fire()

		}
	}
}

// handleKeyup handles the keyup event.
// It removes the key from the keysHeld map.
func (h *handler) handleKeyup(key keyBinding) {
	delete(h.keysHeld, key)
}

// handleTouch handles the touch event.
// It sets the running state to true when the touch event is triggered.
// It moves the spaceship to the left when the delta X is negative.
// It moves the spaceship to the right when the delta X is positive.
// It fires bullets when the touch event is triggered.
func (h *handler) handleTouch(event touchEvent) {
	if h.start() {
		return
	}

	// If there are correlated touch events, pause the game.
	if len(event.Correlations) > 0 {
		h.pause()
		return
	}

	sizeCorrection := h.spaceship.Size.ToVector().Div(2)
	spaceshipPosition := h.spaceship.Position.Add(h.spaceship.Size.ToVector().Div(2))

	// Check if the spaceship is in the swipe proximity range.
	switch {
	case
		event.StartPosition.Distance(spaceshipPosition).Float() <= config.Config.Control.SwipeProximityRange,
		event.CurrentPosition.Distance(spaceshipPosition).Float() <= config.Config.Control.SwipeProximityRange:

		switch {
		case !event.CurrentPosition.IsZero():
			h.spaceship.MoveTo(event.CurrentPosition.Sub(sizeCorrection))

		case !event.StartPosition.IsZero():
			h.spaceship.MoveTo(event.StartPosition.Sub(sizeCorrection))

		}
	}

	h.spaceship.Fire()
}

// pause pauses the game.
func (h *handler) pause() {
	running.Set(&h.ctx, false)

	if config.IsTouchDevice() {
		config.SendMessage(config.Config.Messages.GamePausedTouchDevice)
	} else {
		config.SendMessage(config.Config.Messages.GamePausedNoTouchDevice)
	}
}

// render is a method that renders the game.
// It draws the spaceship, bullets and enemies on the canvas.
// The spaceship is drawn in white color.
// The bullets are drawn in yellow color.
// The enemies are drawn in gray color.
// The goodie enemies are drawn in green color.
// The berserker enemies are drawn in red color.
// The annihilator enemies are drawn in dark red color.
// The spaceship is drawn in dark red color if it is damaged.
// The spaceship is drawn in yellow color if it is boosted.
// The spaceship is drawn in white color if it is normal.
// If draws objects as rectangles.
func (h *handler) render() {
	switch {
	case
		!running.Get(h.ctx) && !isFirstTime.Get(h.ctx),
		suspended.Get(h.ctx):

		return
	}

	config.ClearCanvas()
	config.ClearBackground()

	// Draw stars on the background
	for _, s := range h.stars {
		s.Draw()
	}

	// Draw background
	config.DrawBackground(h.spaceship.Speed.Magnitude().Float() * config.Config.Star.SpeedRatio)

	// Draw spaceship
	h.spaceship.Draw()

	// Draw enemies
	for _, e := range h.enemies {
		e.Draw()
	}

	// Draw bullets
	for _, b := range h.spaceship.Bullets {
		b.Draw()
	}
}

// refresh refreshes the game state.
// It updates the bullets of the spaceship.
// It updates the enemies.
// It updates the state of the spaceship.
// It checks the collisions.
func (h *handler) refresh() {
	if !running.Get(h.ctx) {
		return
	}

	h.enemies.Update(h.spaceship.Position.Add(h.spaceship.Size.ToVector().Div(2)))
	h.spaceship.UpdateState()
	h.spaceship.Bullets.Update()
	h.checkCollisions()
}

// start starts the game if not already started.
func (h *handler) start() bool {
	if suspended.Get(h.ctx) {
		return true
	}

	if !running.Get(h.ctx) {
		running.Set(&h.ctx, true)

		if isFirstTime.Get(h.ctx) {
			if config.IsTouchDevice() {
				config.SendMessage(config.Config.Messages.GameStartedTouchDevice)
			} else {
				config.SendMessage(config.Config.Messages.GameStartedNoTouchDevice)
			}
		}

		go config.PlayAudio("theme_heroic.wav", true)

		return true
	}

	return false
}

// Await waits for the handler to finish and executes the shutdown function.
func (h *handler) Await() {
	<-h.ctx.Done()
	go config.StopAudio("theme_heroic.wav")
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
func (h *handler) Loop() {
	// Notify the user about how to start the game.
	if !running.Get(h.ctx) {
		if isFirstTime.Get(h.ctx) {
			if config.IsTouchDevice() {
				config.SendMessage(config.Config.Messages.HowToStartTouchDevice)
			} else {
				config.SendMessage(config.Config.Messages.HowToStartNoTouchDevice)
			}
		} else {
			if config.IsTouchDevice() {
				config.SendMessage(config.Config.Messages.HowToRestartTouchDevice)
			} else {
				config.SendMessage(config.Config.Messages.HowToRestartNoTouchDevice)
			}
		}
	}

	// Wait for the initial user input.
	for !running.Get(h.ctx) {
		h.render()
		select {
		case <-h.ctx.Done():
			return

		case key := <-h.keydownEvent:
			h.handleKeydown(key)

		case key := <-h.keyupEvent:
			h.handleKeyup(key)

		case event := <-h.touchEvent:
			h.handleTouch(event)

		}
	}

	// Monitor the FPS rate.
	h.monitor()

	ticker := time.NewTicker(time.Second / time.Duration(config.Config.Control.DesiredFramesPerSecondRate))
	for {
		select {
		case <-h.ctx.Done():
			return

		case <-ticker.C:
			h.refresh()
			h.render()
			h.handleKeyhold()

		case key := <-h.keydownEvent:
			h.handleKeydown(key)

		case key := <-h.keyupEvent:
			h.handleKeyup(key)

		case event := <-h.touchEvent:
			h.handleTouch(event)

		}
	}
}

// Restart restarts the game.
func (h *handler) Restart() {
	h.spaceship = spaceship.Embark()
	h.enemies = nil
	h.stars = star.Explode(config.Config.Star.Count)
	h.ctx, h.cancel = context.WithCancel(context.Background())
	running.Set(&h.ctx, false)
	isFirstTime.Set(&h.ctx, false)
}

// New creates a new handler.
// It creates a new spaceship and registers the keydown event.
func New() *handler {
	h := &handler{
		keydownEvent: make(chan keyBinding),
		keyupEvent:   make(chan keyBinding),
		keysHeld:     make(map[keyBinding]bool),
		touchEvent:   make(chan touchEvent),
		spaceship:    spaceship.Embark(),
		stars:        star.Explode(config.Config.Star.Count),
	}

	h.ctx, h.cancel = context.WithCancel(context.Background())
	running.Set(&h.ctx, false)
	isFirstTime.Set(&h.ctx, true)
	h.registerEventHandlers()

	return h
}
