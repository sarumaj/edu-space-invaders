package bullet

import "github.com/sarumaj/edu-space-invaders/src/pkg/numeric"

// Bullets represents a collection of bullets.
type Bullets []Bullet

// Reload creates a new bullet at the specified position.
// The bullet has the specified damage and skew ratio.
func (bullets *Bullets) Reload(position numeric.Position, damage int, ratio, speedBoost float64) {
	*bullets = append(*bullets, *Craft(position, damage, ratio, speedBoost))
}

// Update updates the bullets.
// It moves the bullets and removes the ones that are out of the screen.
func (bullets *Bullets) Update() {
	var visibleBullets []Bullet
	for i := range *bullets {
		bullet := &(*bullets)[i]

		if bullet.Exhausted {
			continue
		}

		bullet.Move()
		if bullet.Position.Y < 0 {
			continue
		}

		visibleBullets = append(visibleBullets, *bullet)
	}

	*bullets = visibleBullets
}
