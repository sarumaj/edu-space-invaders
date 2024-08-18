package handler

import (
	"context"
	"sync"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/bullet"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/planet"
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
	planet     *planet.Planet       // planet is the planet to be drawn
	spaceship  *spaceship.Spaceship // spaceship is the player's spaceship
	stars      star.Stars           // stars is the list of stars
	touchEvent chan touchEvent      // touchEvent is the channel for touch events
}

// applyGravityOnEnemies applies gravity to the enemies.
// It applies gravity to the enemies, each enemy trapped in the planet's gravity is increasing the planet's mass.
// If the planet is a black hole, it pulls the enemies away, if the spaceship is not within the range of the planet.
func (h *handler) applyGravityOnEnemies() {
	// Apply gravity to the enemies, each enemy trapped in the planet's gravity is increasing the planet's mass.
	for i, e := range h.enemies {
		h.enemies[i].Position = h.planet.ApplyGravity(
			e.Position.Add(e.Size.Half().ToVector()),
			e.Size.Area(),
			true, // Increase the planet's mass
			h.planet.Type == planet.BlackHole && // Reverse gravity field, if the planet is a black hole,
				!h.planet.WithinRange(h.spaceship.Position.Add(h.spaceship.Size.Half().ToVector())) && // the spaceship is not within the range of the planet,
				!h.planet.WithinRange(e.Position.Add(e.Size.Half().ToVector())), // and the enemy is not within the range of the planet.
		).Sub(e.Size.Half().ToVector())
	}
}

// applyGravityOnSpaceship applies gravity to the spaceship.
// It applies gravity to the spaceship.
// The spaceship's mass should not increase the planet's mass.
// If the planet is a black hole or a supernova, it applies gravity to the bullets.
func (h *handler) applyGravityOnSpaceship() {
	// Apply gravity to the spaceship.
	// The spaceship's mass should not increase the planet's mass.
	h.spaceship.Position = h.planet.ApplyGravity(
		h.spaceship.Position.Add(h.spaceship.Size.Half().ToVector()),
		h.spaceship.Size.Area(),
		false, // Do not increase the planet's mass
		false, // Do not reverse the gravity
	).Sub(h.spaceship.Size.Half().ToVector())

	// Correct the spaceship's position if it is out of the canvas.
	canvasDimensions := config.CanvasBoundingBox()
	if h.spaceship.Position.Y.Float() > canvasDimensions.OriginalHeight {
		h.spaceship.Position.Y = numeric.Number(canvasDimensions.OriginalHeight)
	}

	if h.spaceship.Position.X < 0 {
		h.spaceship.Position.X = 0
	} else if h.spaceship.Position.X.Float() > canvasDimensions.OriginalWidth {
		h.spaceship.Position.X = numeric.Number(canvasDimensions.OriginalWidth)
	}

	switch h.planet.Type {
	case planet.BlackHole, planet.Supernova:
		// Apply gravity to the bullets.
		for i, bullet := range h.spaceship.Bullets {
			h.spaceship.Bullets[i].Position = h.planet.ApplyGravity(
				bullet.Position.Add(bullet.Size.Half().ToVector()),
				numeric.Number(config.Config.Bullet.GravityAmplifier)*bullet.Size.Area(),
				false, // Do not increase the planet's mass
				false, // Do not reverse the gravity
			).Sub(bullet.Size.Half().ToVector())

			if numeric.Equal(h.planet.Position, h.spaceship.Bullets[i].Position.Add(bullet.Size.Half().ToVector()), 1e-3) {
				h.spaceship.Bullets[i].Exhaust()
			}
		}
	}
}

// applyPlanetImpact applies the impact of the planet, anomaly, or the sun on the game objects.
// If the planet is Uranus, Neptune, or Pluto, it increases the specialty likeliness of the enemies for freezers.
// If the planet is Mercury, Mars, or Pluto, it increases the berserk likeliness of the enemies.
// If the planet is Jupiter or Saturn, it increases the defense and hitpoints of the enemies.
// If the planet is Venus or Earth, it slows down the spaceship and increases the specialty likeliness of the enemies for goodies.
// If the spaceship is within range of the sun, it unfreezes the spaceship.
// If a freezer is within range of the sun, it unfreezes the freezer.
// If the anomaly is a black hole, it sucks in the bullets and other objects.
// If the anomaly is a supernova, it distorts the bullets and other objects and disables the freezers.
func (h *handler) applyPlanetImpact() {
	defer h.applyGravityOnEnemies()
	defer h.applyGravityOnSpaceship()

	switch h.planet.Type {
	case planet.Sun:
		// If the spaceship is within range of the sun, unfreeze the spaceship.
		if h.spaceship.State.AnyOf(spaceship.Frozen) && h.planet.WithinRange(h.spaceship.Position.Add(h.spaceship.Size.Half().ToVector())) {
			h.spaceship.ChangeState(spaceship.Neutral)
		}

		for i, e := range h.enemies {
			// If a freezer is within range of the sun, unfreeze the freezer.
			if e.Type.AnyOf(enemy.Freezer) && h.planet.WithinRange(h.enemies[i].Position.Add(e.Size.Half().ToVector())) {
				h.enemies[i].Type = enemy.Normal
			}
		}

		h.planet.DoOnce(func() {
			config.SendMessage(config.Execute(
				config.Config.MessageBox.Messages.Templates.PlanetImpactsSystem,
				config.Template{"PlanetName": h.planet.Type.String()},
			)+" It destroys all freezers within its range and unfreezes your spaceship when it happens to be close enough.", false)
		})

	case planet.BlackHole:
		h.planet.DoOnce(func() {
			config.SendMessage(config.Execute(
				config.Config.MessageBox.Messages.Templates.PlanetImpactsSystem,
				config.Template{"PlanetName": h.planet.Type.String()},
			)+" It sucks in the bullets and other objects. Be careful and do not get trapped!", false)
		})

	case planet.Supernova:
		// Unfreeze the spaceship immediately if frozen.
		if h.spaceship.State.AnyOf(spaceship.Frozen) {
			h.spaceship.ChangeState(spaceship.Neutral)
		}

		h.planet.DoOnce(func() {
			config.SendMessage(config.Execute(
				config.Config.MessageBox.Messages.Templates.PlanetImpactsSystem,
				config.Template{"PlanetName": h.planet.Type.String()},
			)+" It distorts the bullets and other objects. Luckily, freezers are ineffective.", false)
		})

	case planet.Uranus, planet.Neptune:
		h.planet.DoOnce(func() {
			// Increases the specialty likeliness of the enemies for freezers.
			for i := range h.enemies {
				h.enemies[i].SpecialtyLikeliness *= map[planet.PlanetType]float64{
					planet.Uranus:  2,
					planet.Neptune: 4,
				}[h.planet.Type]
				h.enemies[i].Surprise(map[enemy.EnemyType]int{enemy.Goodie: 100, enemy.Freezer: 0})
			}

			config.SendMessage(config.Execute(
				config.Config.MessageBox.Messages.Templates.PlanetImpactsSystem,
				config.Template{"PlanetName": h.planet.Type.String()},
			)+" It increases the likeliness for freezers to appear.", false)
		})

	case planet.Mercury, planet.Mars:
		h.planet.DoOnce(func() {
			// Double the berserk likeliness of the enemies.
			for i := range h.enemies {
				h.enemies[i].Level.BerserkLikeliness *= map[planet.PlanetType]float64{
					planet.Mercury: 2,
					planet.Mars:    4,
				}[h.planet.Type]
				h.enemies[i].Berserk()
			}

			config.SendMessage(config.Execute(
				config.Config.MessageBox.Messages.Templates.PlanetImpactsSystem,
				config.Template{"PlanetName": h.planet.Type.String()},
			)+" It increases the berserk likeliness of the enemies.", false)
		})

	case planet.Pluto:
		h.planet.DoOnce(func() {
			// Increases the specialty likeliness of the enemies for freezers and going on a berserk.
			for i := range h.enemies {
				h.enemies[i].SpecialtyLikeliness *= 8
				h.enemies[i].Level.BerserkLikeliness *= 8
				h.enemies[i].Berserk()
				h.enemies[i].Surprise(map[enemy.EnemyType]int{enemy.Goodie: 100, enemy.Freezer: 0})
			}

			config.SendMessage(config.Execute(
				config.Config.MessageBox.Messages.Templates.PlanetImpactsSystem,
				config.Template{"PlanetName": h.planet.Type.String()},
			)+" It increases the likeliness for freezers to appear and going on a berserk.", false)
		})

	case planet.Jupiter, planet.Saturn:
		h.planet.DoOnce(func() {
			// Increases the defense and hitpoints of the enemies.
			factor := map[planet.PlanetType]int{
				planet.Jupiter: 2,
				planet.Saturn:  4,
			}[h.planet.Type]
			for i := range h.enemies {
				h.enemies[i].Level.Defense *= factor
				h.enemies[i].Level.HitPoints *= factor
			}

			config.SendMessage(config.Execute(
				config.Config.MessageBox.Messages.Templates.PlanetImpactsSystem,
				config.Template{"PlanetName": h.planet.Type.String()},
			)+" It increases the defense and hitpoints of the enemies.", false)
		})

	case planet.Venus, planet.Earth:
		h.planet.DoOnce(func() {
			// Slow down the spaceship and increase the specialty likeliness for goodies.
			factor := map[planet.PlanetType]float64{
				planet.Venus: 2,
				planet.Earth: 4,
			}[h.planet.Type]
			h.spaceship.Level.AccelerateRate /= numeric.Number(2.5 * factor)
			for i := range h.enemies {
				h.enemies[i].SpecialtyLikeliness *= factor
				h.enemies[i].Surprise(map[enemy.EnemyType]int{enemy.Goodie: 0, enemy.Freezer: 100})
			}

			config.SendMessage(config.Execute(
				config.Config.MessageBox.Messages.Templates.PlanetImpactsSystem,
				config.Template{"PlanetName": h.planet.Type.String()},
			)+" It slows down the spaceship considerably and increases the likeliness for goodies to appear.", false)
		})

	}
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
// If the enemy is a goodie, it does nothing.
// If the enemy is a freezer and the spaceship is not an admiral, it does nothing.
// If the enemy is a freezer and the planet is not the sun or a supernova, it does nothing.
// If the bullet has not hit the enemy, it does nothing.
func (h *handler) checkCollisions() {
	var collisionDetector func(enemy.Enemy) bool
	var bulletHitDetector func(bullet.Bullet) func(enemy.Enemy) bool

	// Choose the collision detection version.
	// Default is the version 2 if the version is not set.
	switch config.Config.Control.CollisionDetectionVersion.GetWithFallback(3) {
	case 1:
		collisionDetector = h.spaceship.DetectCollisionV1
		bulletHitDetector = func(b bullet.Bullet) func(enemy.Enemy) bool { return b.HasHitV1 }

	case 2:
		collisionDetector = h.spaceship.DetectCollisionV2
		bulletHitDetector = func(b bullet.Bullet) func(enemy.Enemy) bool { return b.HasHitV2 }

	default:
		collisionDetector = h.spaceship.DetectCollisionV3
		bulletHitDetector = func(b bullet.Bullet) func(enemy.Enemy) bool { return b.HasHitV3 }

	}

	h.spaceship.Discover(h.planet)

	for j, e := range h.enemies {
		if e.Level.HitPoints > 0 && collisionDetector(e) {
			// If the spaceship is boosted, destroy the enemy.
			if h.spaceship.State.AnyOf(spaceship.Boosted) {
				h.enemies[j].Destroy()
				if h.spaceship.Level.GainExperience(e) {
					h.spaceship.UpdateHighScore()
					config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.SpaceshipUpgradedByEnemyKill, config.Template{
						"EnemyName":      e.Name,
						"EnemyType":      e.Type,
						"SpaceshipLevel": h.spaceship.Level.Progress,
					}), false)
				}
				return
			}

			// Handle collisions with normal, berserker and annihilator enemies.
			penalty := map[enemy.EnemyType]int{
				enemy.Normal:      config.Config.Spaceship.DefaultPenalty,
				enemy.Berserker:   config.Config.Spaceship.BerserkPenalty,
				enemy.Annihilator: config.Config.Spaceship.AnnihilatorPenalty,
				enemy.Freezer:     config.Config.Spaceship.FreezerPenalty,
			}[e.Type]

			if h.spaceship.State.AnyOf(spaceship.Frozen) {
				h.enemies[j].Destroy()
				h.spaceship.Penalize(penalty)
				config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.SpaceshipDowngradedByEnemy, config.Template{
					"SpaceshipLevel": h.spaceship.Level.Progress,
				}), false)
				if h.spaceship.IsDestroyed() {
					config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.GameOver, config.Template{
						"HighScore": h.spaceship.HighScore,
						"Rank":      config.SetScore(h.spaceship.Commandant, h.spaceship.HighScore),
						"TopScores": config.GetScores(10),
					}), false)
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
				}), false)
				return

			case enemy.Freezer: // If the spaceship has collided with a freezer, freeze the spaceship.
				h.enemies[j].Destroy()
				h.spaceship.ChangeState(spaceship.Frozen)
				h.spaceship.Penalize(config.Config.Spaceship.FreezerPenalty)
				if h.spaceship.IsDestroyed() {
					config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.GameOver, config.Template{
						"HighScore": h.spaceship.HighScore,
						"Rank":      config.SetScore(h.spaceship.Commandant, h.spaceship.HighScore),
						"TopScores": config.GetScores(10),
					}), false)
					h.cancel()
				} else {
					config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.SpaceshipFrozen, config.Template{
						"SpaceshipLevel": h.spaceship.Level.Progress,
						"FreezeDuration": config.Config.Spaceship.FreezeDuration,
					}), false)
				}
				return

			}

			h.enemies[j].Destroy()
			h.spaceship.Penalize(penalty)
			h.spaceship.ChangeState(spaceship.Damaged)
			config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.SpaceshipDowngradedByEnemy, config.Template{
				"SpaceshipLevel": h.spaceship.Level.Progress,
			}), false)
			if h.spaceship.IsDestroyed() {
				config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.GameOver, config.Template{
					"HighScore": h.spaceship.HighScore,
					"Rank":      config.SetScore(h.spaceship.Commandant, h.spaceship.HighScore),
					"TopScores": config.GetScores(10),
				}), false)
				h.cancel()
				return
			}
		}

		for i, b := range h.spaceship.Bullets {
			switch {
			case
				e.Level.HitPoints <= 0,                                // If the enemy has no health points, do nothing.
				e.Type.AnyOf(enemy.Goodie),                            // If the enemy is a goodie, do nothing.
				e.Type.AnyOf(enemy.Freezer) && !h.spaceship.IsAdmiral, // If the enemy is a freezer and the spaceship is not an admiral, do nothing.
				e.Type.AnyOf(enemy.Freezer) && !h.planet.Type.AnyOf(planet.Sun, planet.Supernova), // If the enemy is a freezer and the planet is not the sun or a supernova, do nothing.
				!bulletHitDetector(b)(e): // If the bullet has not hit the enemy, do nothing.

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
					}), false)
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

	// Draw planet
	h.planet.Draw()

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
		config.SendMessage(config.Config.MessageBox.Messages.GamePausedTouchDevice, false)
	} else {
		config.SendMessage(config.Config.MessageBox.Messages.GamePausedNoTouchDevice, false)
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

	// Update the positions of the enemies.
	h.enemies.Update(h.spaceship.Position)

	// Update the position of the planet.
	h.planet.Update(h.spaceship.Level.AccelerateRate * numeric.Number(config.Config.Planet.SpeedRatio))

	// Update the state of the spaceship.
	h.spaceship.UpdateState()

	// Update the positions of the bullets.
	h.spaceship.Bullets.Update()

	// Apply the impact of the planet on the system.
	h.applyPlanetImpact()

	// Check the collisions.
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
				config.SendMessage(config.Config.MessageBox.Messages.GameStartedTouchDevice, false)
			} else {
				config.SendMessage(config.Config.MessageBox.Messages.GameStartedNoTouchDevice, false)
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
	if isFirstTime.Get(h.ctx) {
		config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.Greeting, config.Template{
			"Commandant": h.spaceship.Commandant,
		}), true)
	}

	// Notify the user about how to start the game.
	if !running.Get(h.ctx) {
		if isFirstTime.Get(h.ctx) {
			if config.IsTouchDevice() {
				config.SendMessage(config.Config.MessageBox.Messages.HowToStartTouchDevice, false)
			} else {
				config.SendMessage(config.Config.MessageBox.Messages.HowToStartNoTouchDevice, false)
			}
		} else {
			if config.IsTouchDevice() {
				config.SendMessage(config.Config.MessageBox.Messages.HowToRestartTouchDevice, false)
			} else {
				config.SendMessage(config.Config.MessageBox.Messages.HowToRestartNoTouchDevice, false)
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
	h.spaceship = spaceship.Embark(h.spaceship.Commandant)
	h.enemies = nil
	h.stars = star.Explode(config.Config.Star.Count)
	h.ctx, h.cancel = context.WithCancel(context.Background())
	running.Set(&h.ctx, false)
	isFirstTime.Set(&h.ctx, false)
}

// New creates a new handler.
// It creates a new spaceship and registers all event handlers.
func New() *handler {
	h := &handler{
		keyEvent:   make(chan keyEvent),
		keysHeld:   make(map[keyBinding]bool),
		mouseEvent: make(chan mouseEvent),
		touchEvent: make(chan touchEvent),
		planet:     planet.Reveal(true, true),
		spaceship:  spaceship.Embark(""),
		stars:      star.Explode(config.Config.Star.Count),
	}

	h.ctx, h.cancel = context.WithCancel(context.Background())
	running.Set(&h.ctx, false)
	isFirstTime.Set(&h.ctx, true)
	h.registerEventHandlers()
	h.ask()

	return h
}
