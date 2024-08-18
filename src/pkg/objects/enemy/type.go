package enemy

const (
	Normal      EnemyType = iota // Normal is the default enemy type
	Goodie                       // Goodie is the enemy type that can increase the player's spaceship level
	Freezer                      // Freezer is the enemy type that can freeze the player's spaceship
	Berserker                    // Berserker is the enemy type that can harm the player's spaceship more than the normal enemy
	Annihilator                  // Annihilator is the enemy type that can harm the player's spaceship more than the berserker enemy
)

// EnemyType represents the type of the enemy (Normal, Goodie, Freezer, Berserker, Annihilator)
type EnemyType int

// AnyOf returns true if the enemy type is any of the given types.
func (enemyType EnemyType) AnyOf(types ...EnemyType) bool {
	for _, typ := range types {
		if enemyType == typ {
			return true
		}
	}

	return false
}

// String returns the string representation of the enemy type.
func (enemyType EnemyType) String() string {
	return [...]string{"Normal", "Goodie", "Freezer", "Berserker", "Annihilator"}[enemyType]
}
