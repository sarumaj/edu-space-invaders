package objects

const (
	Normal      EnemyType = iota // Normal is the default enemy type
	Goodie                       // Goodie is the enemy type that can increase the player's spaceship level
	Berserker                    // Berserker is the enemy type that can harm the player's spaceship more than the normal enemy
	Annihilator                  // Annihilator is the enemy type that can harm the player's spaceship more than the berserker enemy
)

// EnemyType represents the type of the enemy (Normal, Goodie, Berserker, Annihilator)
type EnemyType int

// String returns the string representation of the enemy type.
func (enemyType EnemyType) String() string {
	return [...]string{"Normal", "Goodie", "Berserker", "Annihilator"}[enemyType]
}
