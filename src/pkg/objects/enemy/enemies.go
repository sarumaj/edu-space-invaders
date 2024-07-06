package enemy

import (
	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// Enemies represents a collection of enemies.
type Enemies []Enemy

// getHighestLevel returns the highest level of the enemies.
func (enemies Enemies) getHighestLevel() int {
	highestLevel := 1
	for _, e := range enemies {
		if e.Level.ID > highestLevel {
			highestLevel = e.Level.ID
		}
	}

	return highestLevel
}

// isOverlapping checks if the new enemy is overlapping with any of the existing enemies.
func (enemies Enemies) isOverlapping(newEnemy Enemy) bool {
	for _, e := range enemies {
		if newEnemy.Position.X < e.Position.X+e.Size.Width+config.EnemyMargin &&
			newEnemy.Position.X+newEnemy.Size.Width+config.EnemyMargin > e.Position.X &&
			newEnemy.Position.Y < e.Position.Y+e.Size.Height+config.EnemyMargin &&
			newEnemy.Position.Y+newEnemy.Size.Height+config.EnemyMargin > e.Position.Y {

			return true
		}
	}

	return false
}

// AppendNew appends a new enemy to the collection.
// The new enemy is created with the specified name and random Y position.
// The new enemy is placed at the highest level of the existing enemies.
// The new enemy is turned into a goodie and berserk based on the probabilities.
func (enemies *Enemies) AppendNew(name string, randomY bool) {
	for {
		newEnemy := Challenge(name, randomY)
		if !enemies.isOverlapping(*newEnemy) {
			newEnemy.ToLevel(enemies.getHighestLevel())
			newEnemy.Surprise()
			newEnemy.Berserk()

			*enemies = append(*enemies, *newEnemy)

			break
		}
	}
}

// Update updates the enemies.
// It moves the enemies and removes the ones that are out of the screen
// or have no health points.
// If the regenerate function is provided, it regenerates the enemies.
// The enemies are regenerated when the spaceship reaches the bottom of the screen.
// The new enemies are placed at the highest level of the existing enemies.
// The new enemies are turned into a goodie and berserk based on the probabilities.
func (enemies *Enemies) Update(spaceshipPosition objects.Position, regenerate func(*Enemies)) {
	var visibleEnemies Enemies
	for i := range *enemies {
		enemy := &(*enemies)[i]
		if enemy.Level.HitPoints <= 0 {
			continue
		}

		enemy.Move(spaceshipPosition)
		if enemy.Position.Y+enemy.Size.Height >= config.CanvasHeight() {
			newEnemy := Challenge(enemy.Name, false)
			newEnemy.ToLevel(enemy.Level.ID + 1)
			newEnemy.Surprise()

			switch enemy.Type {
			case Annihilator:
				newEnemy.Berserk()
				fallthrough

			case Berserker:
				newEnemy.Berserk()
				fallthrough

			case Goodie, Freezer, Normal:
				newEnemy.Berserk()

			}

			*enemy = *newEnemy
		}

		visibleEnemies = append(visibleEnemies, *enemy)
	}

	*enemies = visibleEnemies
	if regenerate != nil {
		regenerate(enemies)
	}
}
