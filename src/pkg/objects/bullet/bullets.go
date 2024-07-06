package bullet

import (
	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// Bullets represents a collection of bullets.
type Bullets []Bullet

// Reload creates a new bullet at the specified position.
// The bullet has the specified damage and skew ratio.
func (bullets *Bullets) Reload(x, y float64, damage int, ratio float64) {
	*bullets = append(*bullets, Bullet{
		Position: objects.Position{
			X: x,
			Y: y,
		},
		Size: objects.Size{
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
