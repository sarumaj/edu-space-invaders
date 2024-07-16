package bullet

import (
	"fmt"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
)

// Bullet represents a bullet shot by the spaceship.
type Bullet struct {
	Position     objects.Position // Position of the bullet
	Size         objects.Size     // Size of the bullet
	CurrentScale objects.Position // Scale of the bullet
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

// HasHit returns true if the bullet has hit the enemy.
func (b Bullet) HasHit(e enemy.Enemy) bool {
	return b.Position.Less(e.Position.Add(e.Size.ToVector())) &&
		b.Position.Add(e.Size.ToVector()).Greater(e.Position)
}

// Move moves the bullet.
// The bullet moves upwards and slightly to the left or right.
// The skew of the bullet is based on the position of the cannon.
func (b *Bullet) Move() {
	b.Position = b.Position.Add(objects.Position{
		Y: -objects.Number(b.Speed),
		X: objects.Number(b.skew * b.Speed),
	})
}

// Scale scales the bullet.
func (bullet *Bullet) Scale() {
	canvasDimensions := config.CanvasBoundingBox()
	scale := objects.Position{
		X: objects.Number(canvasDimensions.ScaleX),
		Y: objects.Number(canvasDimensions.ScaleY),
	}
	_ = objects.
		Measure(bullet.Position, bullet.Size).
		Scale(objects.Ones().DivX(bullet.CurrentScale)).
		Scale(scale).
		ApplyPosition(&bullet.Position).
		ApplySize(&bullet.Size)

	bullet.CurrentScale = scale
}

// String returns the string representation of the bullet.
func (bullet Bullet) String() string {
	return fmt.Sprintf("Bullet (Pos: %s, Speed: %g, Damage: %d)", bullet.Position, bullet.Speed, bullet.Damage)
}

// Craft creates a new bullet at the specified position.
func Craft(position objects.Position, damage int, ratio, speedBoost float64) *Bullet {
	bullet := Bullet{
		Position: position,
		Size: objects.Size{
			Width:  objects.Number(config.Config.Bullet.Width),
			Height: objects.Number(config.Config.Bullet.Height),
		},
		CurrentScale: objects.Ones(),
		Speed:        config.Config.Bullet.Speed + speedBoost,
		Damage:       damage,
		skew:         ratio - 0.5,
	}

	bullet.Scale()
	return &bullet
}
