package config

import "time"

const (
	BulletSpeed                 = 7.0                     // BulletSpeed is the speed of the bullet (in px per frame).
	BulletWidth                 = 4.0                     // BulletWidth is the width of the bullet (in px).
	BulletHeight                = 10.0                    // BulletHeight is the height of the bullet (in px).
	EnemySpecialtyLikeliness    = 0.1                     // EnemySpecialtyLikeliness is the likeliness of an enemy to become a goodie or a freezer.
	EnemyWidth                  = 40.0                    // EnemyWidth is the width of the enemy (in px).
	EnemyHeight                 = 40.0                    // EnemyHeight is the height of the enemy (in px).
	EnemyMargin                 = 10.0                    // EnemyMargin is the margin between enemies (in px).
	EnemyInitialSpeed           = 1.0                     // EnemyInitialSpeed is the initial speed of the enemy (in px per frame).
	EnemyMaximumSpeed           = 5.0                     // EnemyMaximumSpeed is the maximum speed of the enemy (in px per frame).
	EnemyBerserkLikeliness      = 0.01                    // EnemyBerserkLikeliness is the likeliness of an enemy to become a berserker.
	EnemiesCount                = 10                      // EnemiesCount is the number of enemies on the canvas.
	SpaceshipDefaultPenalty     = 3                       // SpaceshipDefaultPenalty is the default penalty of the spaceship.
	SpaceshipBerserkPenalty     = 9                       // SpaceshipBerserkPenalty is the penalty of the spaceship when it collides with a berserker.
	SpaceshipAnnihilatorPenalty = 81                      // SpaceshipAnnihilatorPenalty is the penalty of the spaceship when it collides with an annihilator.
	SpaceshipWidth              = 40.0                    // SpaceshipWidth is the width of the spaceship (in px).
	SpaceshipHeight             = 40.0                    // SpaceshipHeight is the height of the spaceship (in px).
	SpaceshipInitialSpeed       = 25.0                    // SpaceshipInitialSpeed is the initial speed of the spaceship (in px per frame).
	SpaceshipMaximumSpeed       = 50.0                    // SpaceshipMaximumSpeed is the maximum speed of the spaceship (in px per frame).
	SpaceshipCannonProgress     = 10                      // SpaceshipCannonProgress is the amount of spaceship progress to unlock a new cannon.
	SpaceshipMaximumCannons     = 20                      // SpaceshipMaximumCannons is the maximum number of cannons the spaceship can have.
	SpaceshipCooldown           = 100 * time.Millisecond  // SpaceshipCooldown is the cooldown of the spaceship (in ms).
	SpaceshipStateDuration      = 1500 * time.Millisecond // SpaceshipStateDuration is the duration of the spaceship state (in ms).
)

const (
	MessageGameStartedNoTouchDevice = "Game started! Use ARROW KEYS (<, >) to move and SPACE to shoot."
	MessageHowToStartNoTouchDevice  = "Let's begin! Press any key to start."
	MessageGameStartedTouchDevice   = "Game started! Swipe left or right to move and tap to shoot."
	MessageHowToStartTouchDevice    = "Let's begin! Tap to start."
)
