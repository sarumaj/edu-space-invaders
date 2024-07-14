package enemy

import (
	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// Enemies represents a collection of enemies.
type Enemies []Enemy

// countByType counts the number of enemies of the specified type.
func (enemies Enemies) countByTypes() map[EnemyType]int {
	types := make(map[EnemyType]int)
	for _, e := range enemies {
		types[e.Type]++
	}

	return types
}

func (enemies Enemies) getHighestProperty(property func(Enemy) objects.Number) objects.Number {
	var highest objects.Number
	for _, enemy := range enemies {
		if property(enemy) > highest {
			highest = property(enemy)
		}
	}

	return highest
}

// getHighestProgress returns the highest progress of the current enemies generation.
func (enemies Enemies) getHighestProgress() int {
	progress := enemies.getHighestProperty(func(e Enemy) objects.Number { return objects.Number(e.Level.Progress) })
	if progress == 0 {
		return 1
	}

	return progress.Int()
}

// getMostFrequentType returns the most frequent enemy type given the enemies generation.
func (enemies Enemies) getMostFrequentType(types map[EnemyType]int) EnemyType {
	if types == nil {
		types = enemies.countByTypes()
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

// AppendNew appends a new enemy to the collection.
// The new enemy is created with the specified name and random Y position.
// The new enemy is placed at the highest level of the existing enemies.
// The new enemy is turned into a goodie and berserk based on the probabilities.
func (enemies *Enemies) AppendNew(name string, randomY bool) {
	stats := enemies.countByTypes()
	highestProgress := enemies.getHighestProgress()
	frequentType := enemies.getMostFrequentType(stats)

	newEnemy := Challenge(name, randomY)
	newEnemy.ToProgressLevel(highestProgress)
	newEnemy.Surprise(stats)
	newEnemy.BerserkGivenAncestor(frequentType)

	*enemies = append(*enemies, *newEnemy)
}

// Update updates the enemies.
// It moves the enemies and removes the ones that are out of the screen
// or have no health points.
// If the regenerate function is provided, it regenerates the enemies.
// The enemies are regenerated when the spaceship reaches the bottom of the screen.
// The new enemies are placed at the highest level of the existing enemies.
// The new enemies are turned into a goodie and berserk based on the probabilities.
func (enemies *Enemies) Update(spaceshipPosition objects.Position) {
	stats := enemies.countByTypes()

	var visibleEnemies Enemies
	for i := range *enemies {
		enemy := &(*enemies)[i]
		if enemy.Level.HitPoints <= 0 {
			if *config.Config.Enemy.Regenerate {
				visibleEnemies.AppendNew("", false)
			}

			continue
		}

		enemy.Move(spaceshipPosition)
		canvasDimensions := config.CanvasBoundingBox()
		if (enemy.Position.Y + enemy.Size.Height).Float() >= canvasDimensions.Height {
			newEnemy := Challenge(enemy.Name, false)
			newEnemy.ToProgressLevel(enemy.Level.Progress + 1)
			newEnemy.Surprise(stats)
			newEnemy.BerserkGivenAncestor(enemy.Type)
			*enemy = *newEnemy
		}

		visibleEnemies = append(visibleEnemies, *enemy)
	}

	*enemies = visibleEnemies
}
