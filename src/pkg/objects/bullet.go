package objects

import (
	"fmt"
)

// Bullet represents a bullet shot by the spaceship.
type Bullet struct {
	Position      Position // Position of the bullet
	Size          Size     // Size of the bullet
	Speed, Damage int      // Speed and damage of the bullet, Damage is the amount of health points the bullet takes from the enemy
	skew          float64  // Skew of the bullet
	Exhausted     bool     // Exhausted is true if the bullet is out of the screen or has hit an enemy
}

// Exhaust sets the bullet as exhausted.
func (b *Bullet) Exhaust() {
	b.Exhausted = true
}

// HasHit returns true if the bullet has hit the enemy.
func (b Bullet) HasHit(e Enemy) bool {
	return b.Position.X < e.Position.X+e.Size.Width &&
		b.Position.X+b.Size.Width > e.Position.X &&
		b.Position.Y < e.Position.Y+e.Size.Height &&
		b.Position.Y+b.Size.Height > e.Position.Y
}

// Move moves the bullet.
// The bullet moves upwards and slightly to the left or right.
// The skew of the bullet is based on the position of the cannon.
func (b *Bullet) Move() {
	b.Position.Y -= b.Speed
	b.Position.X += int(b.skew * float64(b.Speed))
}

// String returns the string representation of the bullet.
func (bullet Bullet) String() string {
	return fmt.Sprintf("Bullet (Pos: %s, Speed: %d, Damage: %d)", bullet.Position, bullet.Speed, bullet.Damage)
}
