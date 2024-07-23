package handler

import (
	"context"
	"sync"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/bullet"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/spaceship"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/star"
)

// handler is the game handler.
type handler struct {
	ctx        context.Context      // ctx is an abortable context of the handler
	cancel     context.CancelFunc   // cancel is the cancel function of the handler
	enemies    enemy.Enemies        // enemies is the list of enemies
	keyEvent   chan keyEvent        // keyupEvent is the channel for key events
	keysHeld   map[keyBinding]bool  // keysHeld is the map of keys held
	mouseEvent chan mouseEvent      // mouseEvent is the channel for mouse events
	once       sync.Once            // once is meant to register the keydown event only once
	spaceship  *spaceship.Spaceship // spaceship is the player's spaceship
	stars      star.Stars           // stars is the list of stars
	touchEvent chan touchEvent      // touchEvent is the channel for touch events
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
	var collisionDetector func(enemy.Enemy) bool
	var bulletHitDetector func(bullet.Bullet) func(enemy.Enemy) bool

	// Choose the collision detection version.
	// Default is the version 2 if the version is not set.
	switch config.Config.Control.CollisionDetectionVersion.GetWithFallback(2) {
	case 1:
		collisionDetector = h.spaceship.DetectCollisionV1
		bulletHitDetector = func(b bullet.Bullet) func(enemy.Enemy) bool { return b.HasHitV1 }

	default:
		collisionDetector = h.spaceship.DetectCollisionV2
		bulletHitDetector = func(b bullet.Bullet) func(enemy.Enemy) bool { return b.HasHitV2 }

	}

	for j, e := range h.enemies {
		if e.Level.HitPoints > 0 && collisionDetector(e) {
			// If the spaceship is boosted, destroy the enemy.
			if h.spaceship.State == spaceship.Boosted {
				h.enemies[j].Destroy()
				if h.spaceship.Level.GainExperience(e) {
					h.spaceship.UpdateHighScore()
					config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.SpaceshipUpgradedByEnemyKill, config.Template{
						"EnemyName":      e.Name,
						"EnemyType":      e.Type,
						"SpaceshipLevel": h.spaceship.Level.Progress,
					}))
				}
				return
			}

			// Handle collisions with normal, berserker and annihilator enemies.
			penalty := config.Config.Spaceship.DefaultPenalty
			switch e.Type {
			case enemy.Berserker:
				penalty = config.Config.Spaceship.BerserkPenalty

			case enemy.Annihilator:
				penalty = config.Config.Spaceship.AnnihilatorPenalty

			case enemy.Freezer:
				penalty = config.Config.Spaceship.FreezerPenalty

			}

			if h.spaceship.State == spaceship.Frozen {
				h.enemies[j].Destroy()
				h.spaceship.Penalize(penalty)
				config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.SpaceshipDowngradedByEnemy, config.Template{
					"SpaceshipLevel": h.spaceship.Level.Progress,
				}))
				if h.spaceship.IsDestroyed() {
					config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.GameOver, config.Template{
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
				config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.SpaceshipUpgradedByGoodie, config.Template{
					"SpaceshipLevel": h.spaceship.Level.Progress,
					"BoostDuration":  config.Config.Spaceship.BoostDuration,
				}))
				return

			case enemy.Freezer: // If the spaceship has collided with a freezer, freeze the spaceship.
				h.enemies[j].Destroy()
				h.spaceship.ChangeState(spaceship.Frozen)
				h.spaceship.Penalize(config.Config.Spaceship.FreezerPenalty)
				config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.SpaceshipFrozen, config.Template{
					"SpaceshipLevel": h.spaceship.Level.Progress,
					"FreezeDuration": config.Config.Spaceship.FreezeDuration,
				}))
				return

			}

			h.enemies[j].Destroy()
			h.spaceship.Penalize(penalty)
			h.spaceship.ChangeState(spaceship.Damaged)
			config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.SpaceshipDowngradedByEnemy, config.Template{
				"SpaceshipLevel": h.spaceship.Level.Progress,
			}))
			if h.spaceship.IsDestroyed() {
				config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.GameOver, config.Template{
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
				!bulletHitDetector(b)(e):

			default: // The bullet has hit the enemy.
				h.spaceship.Bullets[i].Exhaust()
				h.enemies[j].Hit(b.Damage)

				// If the enemy has no health points, upgrade the spaceship.
				if h.enemies[j].IsDestroyed() && h.spaceship.Level.GainExperience(e) {
					h.spaceship.UpdateHighScore()
					config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.SpaceshipUpgradedByEnemyKill, config.Template{
						"EnemyName":      e.Name,
						"EnemyType":      e.Type,
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

// draw draws the game objects on the canvas.
// It clears the canvas and the background.
// It draws the stars on the background.
// It draws the background.
// It draws the spaceship.
// It draws the enemies.
// It draws the bullets.
func (h *handler) draw() {
	config.ClearCanvas()
	config.ClearBackground()

	// Draw stars on the background
	for _, s := range h.stars {
		s.Draw()
		s.Exhaust()
	}

	// Draw background
	config.DrawBackground(h.spaceship.Level.AccelerateRate.Float() * config.Config.Star.SpeedRatio)

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

// handleKeyEvent handles the key event.
// It sets the running state to true when the key event is triggered.
// It moves the spaceship to the left when the arrow left key is pressed.
// It moves the spaceship to the right when the arrow right key is pressed.
// It moves the spaceship up when the arrow up key is pressed.
// It moves the spaceship down when the arrow down key is pressed.
// It fires bullets when the space key is pressed.
// It pauses the game when the pause key is pressed.
// It removes the key from the keysHeld map when the key is released.
func (h *handler) handleKeyEvent(key keyEvent) {
	if !key.Pressed {
		switch key.Key {
		case ArrowDown, ArrowLeft, ArrowRight, ArrowUp, Space:
			delete(h.keysHeld, key.Key)

		}

		return
	}

	if h.start() {
		return
	}

	switch key.Key {
	case ArrowDown, ArrowLeft, ArrowRight, ArrowUp, Space:
		h.keysHeld[key.Key] = true

	case Pause:
		h.pause()

	}
}

// handleKeyHold handles the key hold event.
// It fires bullets when the space key is held.
func (h *handler) handleKeyhold() {
	for key, ok := range h.keysHeld {
		if !ok {
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

// handleMouse handles the mouse event.
// It sets the running state to true when the mouse event is triggered.
// It moves the spaceship to the left when the delta X is negative.
// It moves the spaceship to the right when the delta X is positive.
// It fires bullets when the mouse event is triggered.
// It pauses the game when the auxiliary or secondary button is pressed.
func (h *handler) handleMouse(event mouseEvent) {
	select {
	case <-h.ctx.Done():
		return

	default:
		switch event.Type {
		case MouseEventTypeDown: // Ignore the mouse down event.
			return

		}

		switch event.Button {
		case MouseButtonPrimary: // pass through

		case MouseButtonAuxiliary, MouseButtonSecondary: // If the auxiliary or secondary button is pressed, pause the game.
			h.pause()
			return

		default: // Do nothing for any other button.
			return

		}

		if !event.Pressed { // If the mouse button is released, do nothing.
			return
		}

		if h.start() { // If the game is just started, do nothing.
			return
		}

		canvasDimensions := config.CanvasBoundingBox()
		positionCorrection := numeric.Position{
			X: numeric.Number(canvasDimensions.ScaleWidth),
			Y: numeric.Number(canvasDimensions.ScaleHeight),
		}

		switch {
		case !event.CurrentPosition.IsZero():
			h.spaceship.MoveTo(event.CurrentPosition.DivX(positionCorrection))

		case !event.StartPosition.IsZero():
			h.spaceship.MoveTo(event.StartPosition.DivX(positionCorrection))

		}

		h.spaceship.Fire()
	}
}

// handleTouch handles the touch event.
// It sets the running state to true when the touch event is triggered.
// It moves the spaceship to the left when the delta X is negative.
// It moves the spaceship to the right when the delta X is positive.
// It fires bullets when the touch event is triggered.
// It pauses the game when the multi-tap event is triggered.
func (h *handler) handleTouch(event touchEvent) {
	switch event.Type {
	case TouchTypeStart: // Ignore the touch start event.
		return

	}

	if h.start() { // If the game is just started, do nothing.
		return
	}

	// If there are correlated touch events, pause the game.
	if event.MultiTap {
		h.pause()
		return
	}

	canvasDimensions := config.CanvasBoundingBox()
	positionCorrection := numeric.Position{
		X: numeric.Number(canvasDimensions.ScaleWidth),
		Y: numeric.Number(canvasDimensions.ScaleHeight),
	}

	switch {
	case !event.CurrentPosition.IsZero():
		h.spaceship.MoveTo(event.CurrentPosition.DivX(positionCorrection))

	case !event.StartPosition.IsZero():
		h.spaceship.MoveTo(event.StartPosition.DivX(positionCorrection))

	}

	h.spaceship.Fire()
}

// pause pauses the game.
func (h *handler) pause() {
	if !running.Get(h.ctx) { // If the game is not running, do nothing.
		return
	}

	paused.Set(&h.ctx, true)     // signal that the game is paused
	running.Set(&h.ctx, false)   // signal that the game is not running
	suspended.Set(&h.ctx, false) // signal that the game is not suspended

	if config.IsTouchDevice() {
		config.SendMessage(config.Config.MessageBox.Messages.GamePausedTouchDevice)
	} else {
		config.SendMessage(config.Config.MessageBox.Messages.GamePausedNoTouchDevice)
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
		!running.Get(h.ctx) && !isFirstTime.Get(h.ctx), // If the game is not running and not the first time, do nothing.
		suspended.Get(h.ctx):                           // If the game is suspended, do nothing.

		return
	}

	h.draw()
}

// refresh refreshes the game state.
// It updates the bullets of the spaceship.
// It updates the enemies.
// It updates the state of the spaceship.
// It checks the collisions.
func (h *handler) refresh() {
	if !running.Get(h.ctx) { // If the game is not running, do nothing.
		return
	}

	h.enemies.Update(h.spaceship.Position)
	h.spaceship.UpdateState()
	h.spaceship.Bullets.Update()
	h.checkCollisions()
}

// start starts the game if not already started.
func (h *handler) start() bool {
	if suspended.Get(h.ctx) { // If the game is suspended, do nothing.
		return true
	}

	if !running.Get(h.ctx) { // If the game is not running, start the game.
		running.Set(&h.ctx, true) // signal that the game is running
		paused.Set(&h.ctx, false) // signal that the game is not paused

		if isFirstTime.Get(h.ctx) {
			if config.IsTouchDevice() {
				config.SendMessage(config.Config.MessageBox.Messages.GameStartedTouchDevice)
			} else {
				config.SendMessage(config.Config.MessageBox.Messages.GameStartedNoTouchDevice)
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
				config.SendMessage(config.Config.MessageBox.Messages.HowToStartTouchDevice)
			} else {
				config.SendMessage(config.Config.MessageBox.Messages.HowToStartNoTouchDevice)
			}
		} else {
			if config.IsTouchDevice() {
				config.SendMessage(config.Config.MessageBox.Messages.HowToRestartTouchDevice)
			} else {
				config.SendMessage(config.Config.MessageBox.Messages.HowToRestartNoTouchDevice)
			}
		}
	}

	// Wait for the initial user input.
	for !running.Get(h.ctx) {
		h.render()
		select {
		case <-h.ctx.Done():
			return

		case key := <-h.keyEvent:
			h.handleKeyEvent(key)

		case event := <-h.mouseEvent:
			h.handleMouse(event)

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

		case key := <-h.keyEvent:
			h.handleKeyEvent(key)

		case event := <-h.mouseEvent:
			h.handleMouse(event)

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
		keyEvent:   make(chan keyEvent),
		keysHeld:   make(map[keyBinding]bool),
		mouseEvent: make(chan mouseEvent),
		touchEvent: make(chan touchEvent),
		spaceship:  spaceship.Embark(),
		stars:      star.Explode(config.Config.Star.Count),
	}

	h.ctx, h.cancel = context.WithCancel(context.Background())
	running.Set(&h.ctx, false)
	isFirstTime.Set(&h.ctx, true)
	h.registerEventHandlers()

	return h
}
