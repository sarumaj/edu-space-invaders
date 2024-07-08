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
	Position            objects.Position //Position of the spaceship
	Size                objects.Size     // Size of the spaceship
	Bullets             bullet.Bullets   // Bullets fired by the spaceship
	Cooldown            time.Duration    // Time between shots
	Level               *SpaceshipLevel  // Spaceship level
	State               SpaceshipState   // Spaceship state
	HighScore           int              // HighScore is the high score of the spaceship.
	lastFired           time.Time        // Last time the spaceship fired
	lastStateTransition time.Time        // Last time the spaceship changed state
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
}

// DetectCollision checks if the spaceship has collided with an enemy.
func (spaceship Spaceship) DetectCollision(e enemy.Enemy) bool {
	return spaceship.Position.X < e.Position.X+e.Size.Width &&
		spaceship.Position.X+spaceship.Size.Width > e.Position.X &&
		spaceship.Position.Y < e.Position.Y+e.Size.Height &&
		spaceship.Position.Y+spaceship.Size.Height > e.Position.Y
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
		return
	}

	if time.Since(spaceship.lastFired) < spaceship.Cooldown {
		return
	}

	for i := 1; i < spaceship.Level.Cannons+1; i++ {
		spaceship.Bullets.Reload(
			spaceship.Position.X.Float()+spaceship.Size.Width.Float()*float64(i)/float64(spaceship.Level.Cannons+1)-config.Config.Bullet.Width/2,
			spaceship.Position.Y.Float(),
			spaceship.GetBulletDamage(),
			float64(i)/float64(spaceship.Level.Cannons+1),
		)
	}

	spaceship.lastFired = time.Now()
}

// GetBulletDamage returns the damage of the bullets fired by the spaceship.
func (spaceship Spaceship) GetBulletDamage() int {
	base := config.Config.Bullet.InitialDamage + spaceship.Level.Progress
	modifier := (spaceship.Level.Progress/config.Config.Bullet.ModifierProgressStep + 1) * spaceship.Level.Cannons
	return base*modifier + rand.Intn(base*modifier)
}

// MoveDown moves the spaceship down.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is greater than the canvas height,
// it is set to the canvas height.
func (spaceship *Spaceship) MoveDown() {
	if spaceship.State == Frozen {
		return
	}

	if spaceship.Position.Y.Float()+spaceship.Size.Height.Float()+spaceship.Level.Speed < config.CanvasHeight() {
		spaceship.Position.Y += objects.Number(spaceship.Level.Speed)
	} else {
		spaceship.Position.Y = objects.Number(config.CanvasHeight() - spaceship.Size.Height.Float())
	}
}

// MoveLeft moves the spaceship to the left.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is less than 0, it is set to 0.
func (spaceship *Spaceship) MoveLeft() {
	if spaceship.State == Frozen {
		return
	}

	if spaceship.Position.X.Float()-spaceship.Level.Speed > 0 {
		spaceship.Position.X -= objects.Number(spaceship.Level.Speed)
	} else {
		spaceship.Position.X = 0
	}
}

// MoveRight moves the spaceship to the right.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is greater than the canvas width,
// it is set to the canvas width.
func (spaceship *Spaceship) MoveRight() {
	if spaceship.State == Frozen {
		return
	}

	if spaceship.Position.X.Float()+spaceship.Size.Width.Float()+spaceship.Level.Speed < config.CanvasWidth() {
		spaceship.Position.X += objects.Number(spaceship.Level.Speed)
	} else {
		spaceship.Position.X = objects.Number(config.CanvasWidth() - spaceship.Size.Width.Float())
	}
}

// MoveUp moves the spaceship up.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is less than 0, it is set to 0.
func (spaceship *Spaceship) MoveUp() {
	if spaceship.State == Frozen {
		return
	}

	if spaceship.Position.Y.Float()-spaceship.Level.Speed > 0 {
		spaceship.Position.Y -= objects.Number(spaceship.Level.Speed)
	} else {
		spaceship.Position.Y = 0
	}
}

// Penalize penalizes the spaceship by downgrading its level.
// The spaceship is downgraded by the specified number of levels.
// If the spaceship's level is less than 1, it is set to 1.
func (spaceship *Spaceship) Penalize(levels int) {
	for i := 0; i < levels && spaceship.Level.Progress > 1; i++ {
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
	if time.Since(spaceship.lastStateTransition) > config.Config.Spaceship.SpecialStateDuration {
		if spaceship.State == Boosted {
			spaceship.Level.Cannons /= 2
			spaceship.Size.Width = objects.Number(config.Config.Spaceship.Width)
			spaceship.Position.X += objects.Number(config.Config.Spaceship.Width / 2)
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
			Progress: 1,
			Cannons:  1,
			Speed:    config.Config.Spaceship.InitialSpeed,
		},
	}
}
