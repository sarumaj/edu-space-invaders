package handler

import (
	"context"
	"sync"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
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
				h.sendMessage(config.Execute(config.Config.Messages.Templates.EnemyDestroyed, config.Template{
					"Name": e.Name,
				}))
				return
			}

			switch e.Type {
			case enemy.Goodie: // If the spaceship has collided with a goodie, upgrade the spaceship.
				h.enemies[j].Destroy()
				h.spaceship.Level.Up()
				h.spaceship.ChangeState(spaceship.Boosted)
				h.sendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipUpgradedByGoodie, config.Template{
					"SpaceshipLevel": h.spaceship.Level.Progress,
				}))
				return

			case enemy.Freezer: // If the spaceship has collided with a freezer, freeze the spaceship.
				h.enemies[j].Destroy()
				h.spaceship.ChangeState(spaceship.Frozen)
				h.spaceship.Penalize(config.Config.Spaceship.FreezerPenalty)
				h.sendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipFrozen, config.Template{
					"SpaceshipLevel": h.spaceship.Level.Progress,
				}))
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

			h.enemies[j].Destroy()
			if h.spaceship.Level.Progress > 1 {
				// Apply penalty to the spaceship.
				h.spaceship.Penalize(penalty)
				h.spaceship.ChangeState(spaceship.Damaged)
				h.sendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipDowngradedByEnemy, config.Template{
					"SpaceshipLevel": h.spaceship.Level.Progress,
				}))
				return
			}

			// If the spaceship has no health points, game over.
			h.spaceship.ChangeState(spaceship.Damaged)
			h.sendMessage(config.Execute(config.Config.Messages.Templates.GameOver, config.Template{
				"HighScore": h.spaceship.HighScore,
			}))
			h.cancel()
			return
		}

		for i, b := range h.spaceship.Bullets {
			switch {
			case
				e.Level.HitPoints <= 0,
				e.Type == enemy.Goodie,
				e.Type == enemy.Freezer,
				!b.HasHit(e):

			default: // The bullet has hit the enemy.
				h.sendMessage(config.Execute(config.Config.Messages.Templates.EnemyHit, config.Template{
					"EnemyName": e.Name,
					"Damage":    h.enemies[j].Hit(b.Damage),
				}))
				h.spaceship.Bullets[i].Exhaust()

				// If the enemy has no health points, upgrade the spaceship.
				if h.enemies[j].Level.HitPoints <= 0 {
					h.sendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipUpgradedByEnemyKill, config.Template{
						"EnemyName":      e.Name,
						"SpaceshipLevel": h.spaceship.Level.Progress,
					}))
					h.spaceship.Level.Up()
					h.spaceship.UpdateHighScore()
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

	delta := event.Delta()
	deltaAbs := delta.Magnitude()
	distance := event.StartPosition.Distance(h.spaceship.Position)

	if config.Config.Control.Debug.Get() {
		h.sendMessage(config.Execute(config.Sprintf(
			"Touch (Delta: %.2f, Distance to spaceship: %.2f, Duration: %s, Correlations: %d)", deltaAbs, distance)))
	}

	// Check if the duration is greater than the hold-tap duration.
	// If the duration is greater than the hold-tap duration and the delta is less than the minimum swipe distance,
	// pause the game.
	if len(event.Correlations) > 1 {
		h.pause()
		return
	}

	// Check if the swipe distance is greater than the minimum swipe distance.
	// Check if the spaceship is in the swipe proximity range.
	if deltaAbs.Float() > config.Config.Control.MinimumSwipeDistance &&
		distance.Float() <= config.Config.Control.SwipeProximityRange {

		if delta.X.Float() < 0 {
			h.spaceship.MoveLeft()
		} else {
			h.spaceship.MoveRight()
		}

		if delta.Y.Float() < 0 {
			h.spaceship.MoveUp()
		} else {
			h.spaceship.MoveDown()
		}
	}

	h.spaceship.Fire()
}

// pause pauses the game.
func (h *handler) pause() {
	running.Set(&h.ctx, false)

	if config.IsTouchDevice() {
		h.sendMessage(config.Config.Messages.GamePausedTouchDevice)
	} else {
		h.sendMessage(config.Config.Messages.GamePausedNoTouchDevice)
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
	if !running.Get(h.ctx) && !isFirstTime.Get(h.ctx) {
		return
	}

	config.ClearCanvas()
	config.ClearBackground()

	// Draw stars on the background
	for _, s := range h.stars {
		s.Draw()
	}

	// Draw background
	config.DrawBackground(h.spaceship.Level.Speed * config.Config.Star.SpeedRatio)

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

	h.enemies.Update(objects.Position{
		X: h.spaceship.Position.X + h.spaceship.Size.Width/2,
		Y: h.spaceship.Position.Y,
	})
	h.spaceship.UpdateState()
	h.spaceship.Bullets.Update()
	h.checkCollisions()
}

// sendMessage sends a message to the message box.
func (*handler) sendMessage(msg string) { config.SendMessage(msg) }

// start starts the game if not already started.
func (h *handler) start() bool {
	if !running.Get(h.ctx) {
		running.Set(&h.ctx, true)

		if isFirstTime.Get(h.ctx) {
			if config.IsTouchDevice() {
				h.sendMessage(config.Config.Messages.GameStartedTouchDevice)
			} else {
				h.sendMessage(config.Config.Messages.GameStartedNoTouchDevice)
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
	if !running.Get(h.ctx) {
		if isFirstTime.Get(h.ctx) {
			if config.IsTouchDevice() {
				h.sendMessage(config.Config.Messages.HowToStartTouchDevice)
			} else {
				h.sendMessage(config.Config.Messages.HowToStartNoTouchDevice)
			}
		} else {
			if config.IsTouchDevice() {
				h.sendMessage(config.Config.Messages.HowToRestartTouchDevice)
			} else {
				h.sendMessage(config.Config.Messages.HowToRestartNoTouchDevice)
			}
		}
	}

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
