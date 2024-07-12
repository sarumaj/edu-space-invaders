package spaceship

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/bullet"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
)

// Spaceship represents the player's spaceship.
type Spaceship struct {
	Position            objects.Position // Position of the spaceship
	Speed               objects.Position // Speed of the spaceship in both directions
	Size                objects.Size     // Size of the spaceship
	Bullets             bullet.Bullets   // Bullets fired by the spaceship
	Cooldown            time.Duration    // Time between shots
	Level               *SpaceshipLevel  // Spaceship level
	State               SpaceshipState   // Spaceship state
	HighScore           int              // HighScore is the high score of the spaceship.
	lastFired           time.Time        // Last time the spaceship fired
	lastStateTransition time.Time        // Last time the spaceship changed state
	lastThrottledLog    time.Time        // Last time the spaceship throttled log messages
}

// ChangeState changes the state of the spaceship.
// If the state is Boosted, the spaceship's cannons are doubled
// and its size is doubled. If the number of cannons exceeds
// the maximum number of cannons, it is set to the maximum number.
func (spaceship *Spaceship) ChangeState(state SpaceshipState) {
	if state == Boosted {
		spaceship.Level.Cannons *= 2
		if spaceship.Level.Cannons > config.Config.Spaceship.MaximumCannons {
			spaceship.Level.Cannons = config.Config.Spaceship.MaximumCannons
		}
		spaceship.Size.Width = objects.Number(config.Config.Spaceship.Width * 2)
		spaceship.Position.X -= objects.Number(config.Config.Spaceship.Width / 2)
	}

	spaceship.State = state
	spaceship.lastStateTransition = time.Now()

	switch spaceship.State {
	case Boosted:
		go config.PlayAudio("spaceship_boost.wav", false)

	case Frozen:
		go config.PlayAudio("spaceship_freeze.wav", false)

	case Damaged:
		go config.PlayAudio("spaceship_crash.wav", false)

	}
}

// DetectCollision checks if the spaceship has collided with an enemy.
func (spaceship Spaceship) DetectCollision(e enemy.Enemy) bool {
	return spaceship.Position.Less(e.Position.Add(e.Size.ToVector())) &&
		spaceship.Position.Add(spaceship.Size.ToVector()).Greater(e.Position)
}

// Draw draws the spaceship on the canvas.
// The spaceship is drawn in white color if it is in the Neutral state.
// The spaceship is drawn in dark red color if it is in the Damaged state.
// The spaceship is drawn in yellow color if it is in the Boosted state.
// The spaceship is drawn in blue color if it is in the Frozen state.
func (spaceship Spaceship) Draw() {
	config.DrawSpaceship(
		spaceship.Position.Pack(),
		spaceship.Size.Pack(),
		true,
		map[SpaceshipState]string{
			Neutral: "white",
			Damaged: "darkred",
			Boosted: "yellow",
			Frozen:  "blue",
		}[spaceship.State],
	)
}

// Fire fires bullets from the spaceship.
// The number of bullets fired is equal to the number of cannons
// the spaceship has. The damage of the bullets is calculated
// based on the spaceship's level.
// The trajectory of the bullets is skewed based on the position
// of the cannon.
func (spaceship *Spaceship) Fire() {
	if spaceship.State == Frozen {
		now := time.Now()
		if now.Sub(spaceship.lastThrottledLog) >= config.Config.Spaceship.LogThrottling {
			config.SendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipStillFrozen, config.Template{
				"FreezeDuration": time.Until(spaceship.lastStateTransition.Add(config.Config.Spaceship.FreezeDuration)).
					Round(config.Config.Spaceship.LogThrottling),
			}))
			spaceship.lastThrottledLog = now
		}
		return
	}

	if time.Since(spaceship.lastFired) < spaceship.Cooldown {
		return
	}

	for i := 1; i < spaceship.Level.Cannons+1; i++ {
		spaceship.Bullets.Reload(
			spaceship.Position.Add(objects.Position{
				// X position of the bullet
				// The X position of the bullet is calculated based on the position of the cannon.
				// The X position of the bullet is the X position of the spaceship plus the width of the spaceship
				// times the position of the cannon minus the width of the bullet divided by 2.
				X: spaceship.Size.Width*objects.Number(i)/objects.Number(spaceship.Level.Cannons+1) - objects.Number(config.Config.Bullet.Width/2),
			}),
			spaceship.GetBulletDamage(),
			// Skew of the bullet
			// Skew is the skew of the bullet based on the position of the cannon.
			float64(i)/float64(spaceship.Level.Cannons+1),
		)
	}

	spaceship.lastFired = time.Now()

	go config.PlayAudio("spaceship_cannon_fire.wav", false)
}

// GetBulletDamage returns the damage of the bullets fired by the spaceship.
func (spaceship Spaceship) GetBulletDamage() int {
	base := config.Config.Bullet.InitialDamage + spaceship.Level.Progress
	modifier := (spaceship.Level.Progress/config.Config.Bullet.ModifierProgressStep + 1) * spaceship.Level.Cannons
	return base*modifier + rand.Intn(base*modifier)
}

// IsDestroyed checks if the spaceship is destroyed.
func (spaceship Spaceship) IsDestroyed() bool {
	return spaceship.Level.Progress == 0
}

// MoveDown moves the spaceship down.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is greater than the canvas height,
// it is set to the canvas height.
func (spaceship *Spaceship) MoveDown() {
	if spaceship.State == Frozen {
		now := time.Now()
		if now.Sub(spaceship.lastThrottledLog) >= config.Config.Spaceship.LogThrottling {
			config.SendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipStillFrozen, config.Template{
				"FreezeDuration": time.Until(spaceship.lastStateTransition.Add(config.Config.Spaceship.FreezeDuration)).
					Round(config.Config.Spaceship.LogThrottling),
			}))
			spaceship.lastThrottledLog = now
		}
		return
	}

	if spaceship.Speed.Y.Float() < config.Config.Spaceship.MaximumSpeed {
		spaceship.Speed.Y += objects.Number(spaceship.Level.AccelerateRate)
	}

	if spaceship.Position.Y+spaceship.Size.Height+spaceship.Speed.Y < objects.Number(config.CanvasHeight()) {
		spaceship.Position.Y += spaceship.Speed.Y
	} else {
		spaceship.Position.Y = objects.Number(config.CanvasHeight() - spaceship.Size.Height.Float())
		spaceship.Speed.Y = 0
	}

	go config.PlayAudio("spaceship_deceleration.wav", false)
}

// MoveLeft moves the spaceship to the left.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is less than 0, it is set to 0.
func (spaceship *Spaceship) MoveLeft() {
	if spaceship.State == Frozen {
		now := time.Now()
		if now.Sub(spaceship.lastThrottledLog) >= config.Config.Spaceship.LogThrottling {
			config.SendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipStillFrozen, config.Template{
				"FreezeDuration": time.Until(spaceship.lastStateTransition.Add(config.Config.Spaceship.FreezeDuration)).
					Round(config.Config.Spaceship.LogThrottling),
			}))
			spaceship.lastThrottledLog = now
		}
		return
	}

	if spaceship.Speed.X.Float() < config.Config.Spaceship.MaximumSpeed {
		spaceship.Speed.X += objects.Number(spaceship.Level.AccelerateRate)
	}

	if spaceship.Position.X-spaceship.Speed.X > 0 {
		spaceship.Position.X -= spaceship.Speed.X
	} else {
		spaceship.Position.X = 0
		spaceship.Speed.X = 0
	}

	go config.PlayAudio("spaceship_whoosh.wav", false)
}

// MoveRight moves the spaceship to the right.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is greater than the canvas width,
// it is set to the canvas width.
func (spaceship *Spaceship) MoveRight() {
	if spaceship.State == Frozen {
		now := time.Now()
		if now.Sub(spaceship.lastThrottledLog) >= config.Config.Spaceship.LogThrottling {
			config.SendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipStillFrozen, config.Template{
				"FreezeDuration": time.Until(spaceship.lastStateTransition.Add(config.Config.Spaceship.FreezeDuration)).
					Round(config.Config.Spaceship.LogThrottling),
			}))
			spaceship.lastThrottledLog = now
		}
		return
	}

	if spaceship.Speed.X.Float() < config.Config.Spaceship.MaximumSpeed {
		spaceship.Speed.X += objects.Number(spaceship.Level.AccelerateRate)
	}

	if spaceship.Position.X+spaceship.Size.Width+spaceship.Speed.X < objects.Number(config.CanvasWidth()) {
		spaceship.Position.X += spaceship.Speed.X
	} else {
		spaceship.Position.X = objects.Number(config.CanvasWidth() - spaceship.Size.Width.Float())
		spaceship.Speed.X = 0
	}

	go config.PlayAudio("spaceship_whoosh.wav", false)
}

// MoveUp moves the spaceship up.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is less than 0, it is set to 0.
func (spaceship *Spaceship) MoveUp() {
	if spaceship.State == Frozen {
		now := time.Now()
		if now.Sub(spaceship.lastThrottledLog) >= config.Config.Spaceship.LogThrottling {
			config.SendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipStillFrozen, config.Template{
				"FreezeDuration": time.Until(spaceship.lastStateTransition.Add(config.Config.Spaceship.FreezeDuration)).
					Round(config.Config.Spaceship.LogThrottling),
			}))
			spaceship.lastThrottledLog = now
		}
		return
	}

	if spaceship.Speed.Y.Float() < config.Config.Spaceship.MaximumSpeed {
		spaceship.Speed.Y += objects.Number(spaceship.Level.AccelerateRate)
	}

	if spaceship.Position.Y-spaceship.Speed.Y > 0 {
		spaceship.Position.Y -= spaceship.Speed.Y
	} else {
		spaceship.Position.Y = 0
		spaceship.Speed.Y = 0
	}

	go config.PlayAudio("spaceship_acceleration.wav", false)
}

// MoveTo moves the spaceship to the specified position.
// The spaceship's position is updated based on the delta.
// If the spaceship's position is less than 0, it is set to 0.
// If the spaceship's position is greater than the canvas width,
// it is set to the canvas width.
// If the spaceship's position is less than 0, it is set to 0.
// If the spaceship's position is greater than the canvas height,
// it is set to the canvas height.
func (spaceship *Spaceship) MoveTo(target objects.Position) {
	if spaceship.State == Frozen {
		now := time.Now()
		if now.Sub(spaceship.lastThrottledLog) >= config.Config.Spaceship.LogThrottling {
			config.SendMessage(config.Execute(config.Config.Messages.Templates.SpaceshipStillFrozen, config.Template{
				"FreezeDuration": time.Until(spaceship.lastStateTransition.Add(config.Config.Spaceship.FreezeDuration)).
					Round(config.Config.Spaceship.LogThrottling),
			}))
			spaceship.lastThrottledLog = now
		}
		return
	}

	if spaceship.Speed.Y.Float() < config.Config.Spaceship.MaximumSpeed {
		spaceship.Speed.Y += objects.Number(spaceship.Level.AccelerateRate)
	}

	if spaceship.Speed.X.Float() < config.Config.Spaceship.MaximumSpeed {
		spaceship.Speed.X += objects.Number(spaceship.Level.AccelerateRate)
	}

	delta := spaceship.Position.Sub(target).Normalize()
	delta = delta.Mul(spaceship.Speed.Magnitude())
	target = spaceship.Position.Sub(delta)

	if target.X < 0 {
		target.X = 0
	}

	if target.X > objects.Number(config.CanvasWidth())-spaceship.Size.Width {
		target.X = objects.Number(config.CanvasWidth()) - spaceship.Size.Width
	}

	if target.Y < 0 {
		target.Y = 0
	}

	if target.Y > objects.Number(config.CanvasHeight())-spaceship.Size.Height {
		target.Y = objects.Number(config.CanvasHeight()) - spaceship.Size.Height
	}

	spaceship.Position = target
}

// Penalize penalizes the spaceship by downgrading its level.
// The spaceship is downgraded by the specified number of levels.
// If the spaceship's level is less than 1, it is set to 1.
func (spaceship *Spaceship) Penalize(levels int) {
	for i := 0; i < levels && spaceship.Level.Progress > 0; i++ {
		spaceship.Level.Down()
	}
}

// String returns a string representation of the spaceship.
func (spaceship Spaceship) String() string {
	return fmt.Sprintf("Spaceship (Lvl: %d, Pos: %s, State: %s)", spaceship.Level.Progress, spaceship.Position, spaceship.State)
}

// UpdateHighScore updates the high score of the spaceship.
func (spaceship *Spaceship) UpdateHighScore() {
	if spaceship.Level.Progress > spaceship.HighScore {
		spaceship.HighScore = spaceship.Level.Progress
	}
}

// UpdateState updates the state of the spaceship.
// If the time since the last state transition is greater than
// the spaceship state duration, the spaceship's state is set to Neutral.
func (spaceship *Spaceship) UpdateState() {
	switch spaceship.State {
	case Boosted:
		if time.Since(spaceship.lastStateTransition) > config.Config.Spaceship.BoostDuration {
			spaceship.Level.Cannons /= 2
			spaceship.Size.Width = objects.Number(config.Config.Spaceship.Width)
			spaceship.Position.X += objects.Number(config.Config.Spaceship.Width / 2)

			if spaceship.Level.Cannons == 0 {
				spaceship.Level.Cannons = 1
			}

			spaceship.State = Neutral
		}

	case Frozen:
		if time.Since(spaceship.lastStateTransition) > config.Config.Spaceship.FreezeDuration {
			spaceship.State = Neutral
		}

	case Damaged:
		if time.Since(spaceship.lastStateTransition) > config.Config.Spaceship.DamageDuration {
			spaceship.State = Neutral
		}

	}
}

// Embark creates a new spaceship.
// The spaceship is created at the bottom of the canvas.
// The spaceship's position, size, cooldown, level, and state are set.
func Embark() *Spaceship {
	return &Spaceship{
		Position: objects.Position{
			X: objects.Number(config.CanvasWidth()/2 - config.Config.Spaceship.Width/2),
			Y: objects.Number(config.CanvasHeight() - config.Config.Spaceship.Height),
		},
		Size: objects.Size{
			Width:  objects.Number(config.Config.Spaceship.Width),
			Height: objects.Number(config.Config.Spaceship.Height),
		},
		Cooldown: config.Config.Spaceship.Cooldown,
		Level: &SpaceshipLevel{
			AccelerateRate: objects.Number(config.Config.Spaceship.Acceleration),
			Progress:       1,
			Cannons:        1,
		},
	}
}
