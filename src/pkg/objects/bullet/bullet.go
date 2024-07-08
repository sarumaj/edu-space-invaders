package bullet

import (
	"fmt"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
)

// Bullet represents a bullet shot by the spaceship.
type Bullet struct {
	Position  objects.Position // Position of the bullet
	Size      objects.Size     // Size of the bullet
	Speed     float64          // Speed and damage of the bullet
	Damage    int              // Damage is the amount of health points the bullet takes from the enemy
	skew      float64          // Skew of the bullet
	Exhausted bool             // Exhausted is true if the bullet is out of the screen or has hit an enemy
}

// Draw draws the bullet.
// The bullet is drawn as a yellow rectangle.
func (bullet Bullet) Draw() {
	switch {
	case bullet.Damage > 1_000_000:
		config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), "purple")
	case bullet.Damage > 100_000:
		config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), "blue")
	case bullet.Damage > 10_000:
		config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), "violet")
	case bullet.Damage > 1_000:
		config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), "red")
	case bullet.Damage > 100:
		config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), "orange")
	default:
		config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), "yellow")
	}
}

// Exhaust sets the bullet as exhausted.
func (b *Bullet) Exhaust() {
	b.Exhausted = true
}

// HasHit returns true if the bullet has hit the enemy.
func (b Bullet) HasHit(e enemy.Enemy) bool {
	return b.Position.X < e.Position.X+e.Size.Width &&
		b.Position.X+b.Size.Width > e.Position.X &&
		b.Position.Y < e.Position.Y+e.Size.Height &&
		b.Position.Y+b.Size.Height > e.Position.Y
}

// Move moves the bullet.
// The bullet moves upwards and slightly to the left or right.
// The skew of the bullet is based on the position of the cannon.
func (b *Bullet) Move() {
	b.Position.Y -= objects.Number(b.Speed)
	b.Position.X += objects.Number(b.skew * b.Speed)
}

// String returns the string representation of the bullet.
func (bullet Bullet) String() string {
	return fmt.Sprintf("Bullet (Pos: %s, Speed: %g, Damage: %d)", bullet.Position, bullet.Speed, bullet.Damage)
}
