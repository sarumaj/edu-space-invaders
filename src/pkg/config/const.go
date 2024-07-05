package config

import "time"

const (
	BulletSpeed                 = 7                       // BulletSpeed is the speed of the bullet (in px per frame).
	BulletWidth                 = 4                       // BulletWidth is the width of the bullet (in px).
	BulletHeight                = 10                      // BulletHeight is the height of the bullet (in px).
	CanvasWidth                 = 800                     // CanvasWidth is the width of the canvas (in px).
	CanvasHeight                = 600                     // CanvasHeight is the height of the canvas (in px).
	GoodieLikeliness            = 0.1                     // GoodieLikeliness is the likeliness of an enemy to become a goodie.
	EnemyWidth                  = 40                      // EnemyWidth is the width of the enemy (in px).
	EnemyHeight                 = 40                      // EnemyHeight is the height of the enemy (in px).
	EnemyMargin                 = 10                      // EnemyMargin is the margin between enemies (in px).
	EnemyInitialSpeed           = 1                       // EnemyInitialSpeed is the initial speed of the enemy (in px per frame).
	EnemyMaximumSpeed           = 5                       // EnemyMaximumSpeed is the maximum speed of the enemy (in px per frame).
	EnemyBerserkLikeliness      = 0.01                    // EnemyBerserkLikeliness is the likeliness of an enemy to become a berserker.
	EnemiesCount                = 10                      // EnemiesCount is the number of enemies on the canvas.
	SpaceshipDefaultPenalty     = 3                       // SpaceshipDefaultPenalty is the default penalty of the spaceship.
	SpaceshipBerserkPenalty     = 9                       // SpaceshipBerserkPenalty is the penalty of the spaceship when it collides with a berserker.
	SpaceshipAnnihilatorPenalty = 81                      // SpaceshipAnnihilatorPenalty is the penalty of the spaceship when it collides with an annihilator.
	SpaceshipWidth              = 40                      // SpaceshipWidth is the width of the spaceship (in px).
	SpaceshipHeight             = 40                      // SpaceshipHeight is the height of the spaceship (in px).
	SpaceshipInitialSpeed       = 25                      // SpaceshipInitialSpeed is the initial speed of the spaceship (in px per frame).
	SpaceshipMaximumSpeed       = 50                      // SpaceshipMaximumSpeed is the maximum speed of the spaceship (in px per frame).
	SpaceshipMaximumCannons     = 10                      // SpaceshipMaximumCannons is the maximum number of cannons the spaceship can have.
	SpaceshipCooldown           = 10 * time.Millisecond   // SpaceshipCooldown is the cooldown of the spaceship (in ms).
	SpaceshipStateDuration      = 1500 * time.Millisecond // SpaceshipStateDuration is the duration of the spaceship state (in ms).
)
