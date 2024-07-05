package objects

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
)

// Spaceship represents the player's spaceship.
type Spaceship struct {
	Position            Position        //Position of the spaceship
	Size                Size            // Size of the spaceship
	Bullets             Bullets         // Bullets fired by the spaceship
	Cooldown            time.Duration   // Time between shots
	Level               *SpaceshipLevel // Spaceship level
	State               SpaceshipState  // Spaceship state
	lastFired           time.Time       // Last time the spaceship fired
	lastStateTransition time.Time       // Last time the spaceship changed state
}

// ChangeState changes the state of the spaceship.
// If the state is Boosted, the spaceship's cannons are doubled
// and its size is doubled. If the number of cannons exceeds
// the maximum number of cannons, it is set to the maximum number.
func (spaceship *Spaceship) ChangeState(state SpaceshipState) {
	if state == Boosted {
		spaceship.Level.Cannons *= 2
		if spaceship.Level.Cannons > config.SpaceshipMaximumCannons {
			spaceship.Level.Cannons = config.SpaceshipMaximumCannons
		}
		spaceship.Size.Width = config.SpaceshipWidth * 2
		spaceship.Position.X -= config.SpaceshipWidth / 2
	}

	spaceship.State = state
	spaceship.lastStateTransition = time.Now()
}

// DetectCollision checks if the spaceship has collided with an enemy.
func (spaceship Spaceship) DetectCollision(e Enemy) bool {
	return spaceship.Position.X < e.Position.X+e.Size.Width &&
		spaceship.Position.X+spaceship.Size.Width > e.Position.X &&
		spaceship.Position.Y < e.Position.Y+e.Size.Height &&
		spaceship.Position.Y+spaceship.Size.Height > e.Position.Y
}

// Fire fires bullets from the spaceship.
// The number of bullets fired is equal to the number of cannons
// the spaceship has. The damage of the bullets is calculated
// based on the spaceship's level.
// The trajectory of the bullets is skewed based on the position
// of the cannon.
func (spaceship *Spaceship) Fire() {
	if time.Since(spaceship.lastFired) < spaceship.Cooldown {
		return
	}

	for i := 1; i < spaceship.Level.Cannons+1; i++ {
		spaceship.Bullets.Reload(
			spaceship.Position.X+spaceship.Size.Width*i/(spaceship.Level.Cannons+1)-config.BulletWidth/2,
			spaceship.Position.Y,
			spaceship.GetBulletDamage(),
			float64(i)/float64(spaceship.Level.Cannons+1),
		)
	}

	spaceship.lastFired = time.Now()
}

// GetBulletDamage returns the damage of the bullets fired by the spaceship.
func (spaceship Spaceship) GetBulletDamage() int {
	return 30 + spaceship.Level.ID + rand.Intn(30+spaceship.Level.ID)
}

// MoveLeft moves the spaceship to the left.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is less than 0, it is set to 0.
func (spaceship *Spaceship) MoveLeft() {
	if spaceship.Position.X-spaceship.Level.Speed > 0 {
		spaceship.Position.X -= spaceship.Level.Speed
	} else {
		spaceship.Position.X = 0
	}
}

// MoveRight moves the spaceship to the right.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is greater than the canvas width,
// it is set to the canvas width.
func (spaceship *Spaceship) MoveRight() {
	if spaceship.Position.X+spaceship.Size.Width+spaceship.Level.Speed < config.CanvasWidth {
		spaceship.Position.X += spaceship.Level.Speed
	} else {
		spaceship.Position.X = config.CanvasWidth - spaceship.Size.Width
	}
}

// Penalize penalizes the spaceship by downgrading its level.
// The spaceship is downgraded by the specified number of levels.
// If the spaceship's level is less than 1, it is set to 1.
func (spaceship *Spaceship) Penalize(levels int) {
	for i := 0; i < levels && spaceship.Level.ID > 1; i++ {
		spaceship.Level.Down()
	}
}

// String returns a string representation of the spaceship.
func (spaceship Spaceship) String() string {
	return fmt.Sprintf("Spaceship (Lvl: %d, Pos: %s, State: %s)", spaceship.Level.ID, spaceship.Position, spaceship.State)
}

// UpdateState updates the state of the spaceship.
// If the time since the last state transition is greater than
// the spaceship state duration, the spaceship's state is set to Neutral.
func (spaceship *Spaceship) UpdateState() {
	if time.Since(spaceship.lastStateTransition) > config.SpaceshipStateDuration {
		if spaceship.State == Boosted {
			spaceship.Level.Cannons /= 2
			spaceship.Size.Width = config.SpaceshipWidth
			spaceship.Position.X += config.SpaceshipWidth / 2
		}

		if spaceship.Level.Cannons == 0 {
			spaceship.Level.Cannons = 1
		}

		spaceship.State = Neutral
	}
}

// Embark creates a new spaceship.
// The spaceship is created at the center of the canvas.
// The spaceship's position, size, cooldown, level, and state are set.
func Embark() *Spaceship {
	return &Spaceship{
		Position: Position{
			X: config.CanvasWidth / 2,
			Y: config.CanvasHeight - config.SpaceshipHeight,
		},
		Size: Size{
			Width:  config.SpaceshipWidth,
			Height: config.SpaceshipHeight,
		},
		Cooldown: config.SpaceshipCooldown,
		Level: &SpaceshipLevel{
			ID:      1,
			Cannons: 1,
			Speed:   config.SpaceshipInitialSpeed,
		},
	}
}
