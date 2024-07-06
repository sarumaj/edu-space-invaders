package enemy

import "github.com/sarumaj/edu-space-invaders/src/pkg/config"

// EnemyLevel represents the enemy level.
type EnemyLevel struct {
	ID                 int     // ID is the level of the enemy.
	Speed              float64 // Speed is the speed of the enemy
	HitPoints, Defense int     // HitPoints is the health points of the enemy, Defense is the defense of the enemy.
	BerserkLikeliness  float64 // BerserkLikeliness is the likeliness of the enemy to become berserk or to become an annihilator.
}

// Down decreases the enemy level.
// If the enemy level is already at the minimum level, it does nothing.
// If the enemy speed is greater than the initial speed, it decreases the speed by 1.
// If the enemy berserk likeliness is greater than the initial berserk likeliness, it decreases the berserk likeliness by 0.01.
// If the enemy hit points are greater than 100, it decreases the hit points by 10.
// If the enemy defense is greater than 0, it decreases the defense by 10.
func (lvl *EnemyLevel) Down() {
	if lvl.ID == 1 {
		return
	}

	if lvl.Speed > config.EnemyInitialSpeed {
		lvl.Speed -= 1
	}

	if lvl.BerserkLikeliness > config.EnemyBerserkLikeliness {
		lvl.BerserkLikeliness -= 0.01
	}

	if lvl.HitPoints > 100 {
		lvl.HitPoints -= 10
	}

	if lvl.Defense > 0 {
		lvl.Defense -= 10
	}

	lvl.ID--
}

// Up increases the enemy level.
// If the enemy speed is less than the maximum speed, it increases the speed by 1.
// It increases the berserk likeliness by 0.01, the hit points by 10 and the defense by 10.
func (lvl *EnemyLevel) Up() {
	if lvl.Speed < config.EnemyMaximumSpeed {
		lvl.Speed += 1
	}

	lvl.BerserkLikeliness += 0.01
	lvl.HitPoints += 10
	lvl.Defense += 10
	lvl.ID++
}
