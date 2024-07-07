package enemy

import (
	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// Enemies represents a collection of enemies.
type Enemies []Enemy

// getHighestProgress returns the highest progress of the current enemies generation.
func (enemies Enemies) getHighestProgress() int {
	highestLevel := 1
	for _, e := range enemies {
		if e.Level.Progress > highestLevel {
			highestLevel = e.Level.Progress
		}
	}

	return highestLevel
}

// getMostFrequentType returns the most frequent enemy type given the enemies generation.
func (enemies Enemies) getMostFrequentType() EnemyType {
	types := make(map[EnemyType]int)
	for _, e := range enemies {
		types[e.Type]++
	}

	var max int
	var mostFrequentType EnemyType
	for t, c := range types {
		if c > max {
			max = c
			mostFrequentType = t
		}
	}

	return mostFrequentType
}

// isOverlapping checks if the new enemy is overlapping with any of the existing enemies.
func (enemies Enemies) isOverlapping(newEnemy Enemy) bool {
	for _, e := range enemies {
		if newEnemy.Position.X.Float() < e.Position.X.Float()+e.Size.Width.Float()+config.Config.Enemy.Margin &&
			newEnemy.Position.X.Float()+newEnemy.Size.Width.Float()+config.Config.Enemy.Margin > e.Position.X.Float() &&
			newEnemy.Position.Y.Float() < e.Position.Y.Float()+e.Size.Height.Float()+config.Config.Enemy.Margin &&
			newEnemy.Position.Y.Float()+newEnemy.Size.Height.Float()+config.Config.Enemy.Margin > e.Position.Y.Float() {

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
			newEnemy.ToProgressLevel(enemies.getHighestProgress())
			newEnemy.Surprise()
			newEnemy.BerserkGivenAncestor(enemies.getMostFrequentType())

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
		if enemy.Position.Y.Float()+enemy.Size.Height.Float() >= config.CanvasHeight() {
			newEnemy := Challenge(enemy.Name, false)
			newEnemy.ToProgressLevel(enemy.Level.Progress + 1)
			newEnemy.Surprise()
			newEnemy.BerserkGivenAncestor(enemy.Type)
			*enemy = *newEnemy
		}

		visibleEnemies = append(visibleEnemies, *enemy)
	}

	*enemies = visibleEnemies
	if regenerate != nil {
		regenerate(enemies)
	}
}
