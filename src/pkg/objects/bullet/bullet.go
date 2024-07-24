package bullet

import (
	"fmt"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
)

// Bullet represents a bullet shot by the spaceship.
type Bullet struct {
	Position     numeric.Position // Position of the bullet
	Size         numeric.Size     // Size of the bullet
	CurrentScale numeric.Position // Scale of the bullet
	Speed        float64          // Speed and damage of the bullet
	Damage       int              // Damage is the amount of health points the bullet takes from the enemy
	skew         float64          // Skew of the bullet
	Exhausted    bool             // Exhausted is true if the bullet is out of the screen or has hit an enemy
}

// Draw draws the bullet.
// The bullet is drawn as a yellow rectangle.
func (bullet Bullet) Draw() {
	switch {
	case bullet.Damage > 1_000_000:
		config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), "Red", 0)

	case bullet.Damage > 100_000:
		config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), "MediumVioletRed", 0)

	case bullet.Damage > 10_000:
		config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), "Goldenrod", 0)

	case bullet.Damage > 1_000:
		config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), "Coral", 0)

	case bullet.Damage > 100:
		config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), "Aquamarine", 0)

	default:
		config.DrawRect(bullet.Position.Pack(), bullet.Size.Pack(), "Bisque", 0)

	}
}

// Exhaust sets the bullet as exhausted.
func (b *Bullet) Exhaust() {
	b.Exhausted = true
}

// HasHitV1 returns true if the bullet has hit the enemy.
// This version is less precise than HasHitV2.
// It uses a simple bounding box collision check.
func (b Bullet) HasHitV1(e enemy.Enemy) bool {
	return b.Position.Less(e.Position.Add(e.Size.ToVector())) &&
		b.Position.Add(e.Size.ToVector()).Greater(e.Position)
}

// HasHitV2 returns true if the bullet has hit the enemy.
// This version is more precise than HasHit.
// It uses the Separating Axis Theorem to check for collision.
// The Separating Axis Theorem states that if two convex shapes do not overlap on any axis, then they do not intersect.
// The axes to test are the normals to the edges of the triangle and the rectangle.
// If there is a separating axis, there is no collision.
// It assumes that the bullet is a rectangle and the enemy is a triangle.
func (b Bullet) HasHitV2(e enemy.Enemy) bool {
	// The vertices of the bullet and the enemy
	bulletVertices := numeric.GetRectangularVertices(b.Position, b.Size)
	enemyVertices := numeric.GetSpaceshipVerticesV1(e.Position, e.Size, false)

	// Check for overlap on all axes
	// Do not sort assuming that the vertices are already sorted
	return !numeric.HaveSeparatingAxis(bulletVertices[:], enemyVertices[:], false)
}

// HasHitV3 returns true if the bullet has hit the enemy.
// This version is more precise than HasHitV2.
// It uses the Separating Axis Theorem to check for collision.
// The Separating Axis Theorem states that if two convex shapes do not overlap on any axis, then they do not intersect.
// The axes to test are the normals to the edges of the spaceship polygon and the bullet rectangle.
// If there is a separating axis, there is no collision.
// It assumes that the bullet is a rectangle and the enemy is a spaceship polygon.
func (b Bullet) HasHitV3(e enemy.Enemy) bool {
	// The vertices of the bullet and the enemy
	bulletVertices := numeric.GetRectangularVertices(b.Position, b.Size)
	enemyVertices := numeric.GetSpaceshipVerticesV2(e.Position, e.Size, false)

	// Check for overlap on all axes
	// Do not sort assuming that the vertices are already sorted
	return !numeric.HaveSeparatingAxis(bulletVertices[:], enemyVertices[:], false)
}

// Move moves the bullet.
// The bullet moves upwards and slightly to the left or right.
// The skew of the bullet is based on the position of the cannon.
func (b *Bullet) Move() {
	b.Position = b.Position.Add(numeric.Position{
		Y: -numeric.Number(b.Speed),
		X: numeric.Number(b.skew * b.Speed),
	})
}

// String returns the string representation of the bullet.
func (bullet Bullet) String() string {
	return fmt.Sprintf("Bullet (Pos: %s, Speed: %g, Damage: %d)", bullet.Position, bullet.Speed, bullet.Damage)
}

// Craft creates a new bullet at the specified position.
func Craft(position numeric.Position, damage int, ratio, speedBoost float64) *Bullet {
	bullet := Bullet{
		Position:     position,
		Size:         numeric.Locate(config.Config.Bullet.Width, config.Config.Bullet.Height).ToBox(),
		CurrentScale: numeric.Ones(),
		Speed:        config.Config.Bullet.Speed + speedBoost,
		Damage:       damage,
		skew:         ratio - 0.5,
	}

	return &bullet
}
