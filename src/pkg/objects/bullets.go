package objects

import "github.com/sarumaj/edu-space-invaders/src/pkg/config"

// Bullets represents a collection of bullets.
type Bullets []Bullet

// Reload creates a new bullet at the specified position.
// The bullet has the specified damage and skew ratio.
func (bullets *Bullets) Reload(x, y, damage int, ratio float64) {
	*bullets = append(*bullets, Bullet{
		Position: Position{
			X: x,
			Y: y,
		},
		Size: Size{
			Width:  config.BulletWidth,
			Height: config.BulletHeight,
		},
		Speed:  config.BulletSpeed,
		Damage: damage,
		skew:   ratio - 0.5,
	})
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
