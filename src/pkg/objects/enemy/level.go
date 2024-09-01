package enemy

import (
	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

// EnemyLevel represents the enemy level.
type EnemyLevel struct {
	Progress           int            // Progress is the level of the enemy.
	Speed              numeric.Number // Speed is the speed of the enemy
	HitPoints, Defense int            // HitPoints is the health points of the enemy, Defense is the defense of the enemy.
	HitPointsLoss      int            // HitPointsLoss is the loss of hit points of the enemy.
	BerserkLikeliness  numeric.Number // BerserkLikeliness is the likeliness of the enemy to become berserk or to become an annihilator.
}

// Down decreases the enemy level.
// If the enemy level is already at the minimum level, it does nothing.
// If the enemy speed is greater than the initial speed, it decreases the speed by 1.
// If the enemy berserk likeliness is greater than the initial berserk likeliness, it decreases the berserk likeliness by 0.01.
// If the enemy hit points are greater than 100, it decreases the hit points by 10.
// If the enemy defense is greater than 0, it decreases the defense by 10.
func (lvl *EnemyLevel) Down() {
	if lvl.Progress == 0 {
		return
	}

	lvl.Speed = (lvl.Speed - numeric.Number(config.Config.Enemy.AccelerationProgress)).
		Max(numeric.Number(config.Config.Enemy.InitialSpeed))

	lvl.BerserkLikeliness = (lvl.BerserkLikeliness - numeric.Number(config.Config.Enemy.BerserkLikelinessProgress)).
		Max(numeric.Number(config.Config.Enemy.BerserkLikelinessProgress))

	lvl.HitPoints = numeric.Number(lvl.HitPoints - config.Config.Enemy.HitpointProgress).
		Max(numeric.Number(config.Config.Enemy.InitialHitpoints)).Int()

	lvl.Defense = numeric.Number(lvl.Defense - config.Config.Enemy.DefenseProgress).
		Max(0).Int()

	lvl.Progress -= 1
}

// Up increases the enemy level.
// If the enemy speed is less than the maximum speed, it increases the speed by 1.
// It increases the berserk likeliness by 0.01, the hit points by 10 and the defense by 10.
func (lvl *EnemyLevel) Up() {
	lvl.Speed = (lvl.Speed + numeric.Number(config.Config.Enemy.AccelerationProgress)).
		Min(numeric.Number(config.Config.Enemy.MaximumSpeed))

	lvl.BerserkLikeliness = (lvl.BerserkLikeliness + numeric.Number(config.Config.Enemy.BerserkLikelinessProgress)).
		Min(1)

	lvl.HitPoints += config.Config.Enemy.HitpointProgress
	lvl.Defense += config.Config.Enemy.DefenseProgress
	lvl.Progress += 1
}
