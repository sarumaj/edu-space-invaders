package bullet

import (
	"fmt"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
)

// Bullet represents a bullet shot by the spaceship.
type Bullet struct {
	Position  numeric.Position // Position of the bullet
	Size      numeric.Size     // Size of the bullet
	Speed     numeric.Number   // Speed and damage of the bullet
	Damage    int              // Damage is the amount of health points the bullet takes from the enemy
	skew      numeric.Number   // Skew of the bullet
	Exhausted bool             // Exhausted is true if the bullet is out of the screen or has hit an enemy
}

// Draw draws the bullet.
// The bullet is drawn as a yellow rectangle.
func (bullet Bullet) Draw() {
	var color string
	switch {
	case bullet.Damage > 25_000:
		color = "DarkRed" // Very high damage, intense color
	case bullet.Damage > 12_000:
		color = "Crimson" // High damage, strong but less intense
	case bullet.Damage > 5_000:
		color = "DarkOrange" // Moderate damage, warm color
	case bullet.Damage > 1_000:
		color = "Orange" // Lower damage, vibrant but softer
	case bullet.Damage > 100:
		color = "Gold" // Low damage, bright and shiny
	default:
		color = "Lavender" // Minimal damage, soft and neutral
	}

	//config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), color, 0)
	config.DrawLine(
		bullet.Position.Pack(), // Start position
		bullet.Position.Add(numeric.Locate(-bullet.skew*bullet.Size.Height, bullet.Size.Height)).Pack(), // End position
		color,                     // Color
		bullet.Size.Width.Float(), // Width
	)
}

// Exhaust sets the bullet as exhausted.
func (b *Bullet) Exhaust() {
	b.Exhausted = true
}

// HasHit returns true if the bullet has hit the enemy.
// It uses the Separating Axis Theorem to check for collision.
// The Separating Axis Theorem states that if two convex shapes do not overlap on any axis, then they do not intersect.
// The axes to test are the normals to the edges of the spaceship polygon and the bullet rectangle.
// If there is a separating axis, there is no collision.
// It assumes that the bullet is a rectangle and the enemy is a spaceship polygon.
func (b Bullet) HasHit(e enemy.Enemy) bool {
	if e.Type == enemy.Goodie {
		return false
	}

	switch config.Config.Control.CollisionDetectionVersion.Get() {
	case 1:
		return !numeric.GetRectangularVertices(b.Position, b.Size, false).
			Vertices().
			HasSeparatingAxis(numeric.GetRectangularVertices(e.Position, e.Size, true).
				Vertices())

	case 2:
		return !numeric.GetRectangularVertices(b.Position, b.Size, false).
			Vertices().
			HasSeparatingAxis(numeric.GetSpaceshipVerticesV1(e.Position, e.Size, e.Type == enemy.Goodie).
				Vertices())

	case 3:
		return !numeric.
			GetRectangularVertices(b.Position, b.Size, false).
			Vertices().
			HasSeparatingAxis(numeric.GetSpaceshipVerticesV2(e.Position, e.Size, e.Type == enemy.Goodie).
				Vertices())

	}

	return false
}

// Move moves the bullet.
// The bullet moves upwards and slightly to the left or right.
// The skew of the bullet is based on the position of the cannon.
func (b *Bullet) Move() {
	b.Position = b.Position.Add(numeric.Locate(b.skew*b.Speed, -b.Speed))
}

// String returns the string representation of the bullet.
func (bullet Bullet) String() string {
	return fmt.Sprintf("Bullet (Pos: %s, Speed: %g, Damage: %d)", bullet.Position, bullet.Speed, bullet.Damage)
}

// Craft creates a new bullet at the specified position.
func Craft(position numeric.Position, damage int, ratio, speedBoost numeric.Number) *Bullet {
	bullet := Bullet{
		Position: position,
		Size:     numeric.Locate(config.Config.Bullet.Width, config.Config.Bullet.Height).ToBox(),
		Speed:    numeric.Number(config.Config.Bullet.Speed) + speedBoost,
		Damage:   damage,
		skew:     ratio - 0.5,
	}

	return &bullet
}
