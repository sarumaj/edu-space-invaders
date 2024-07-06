package config

import "time"

const (
	BulletHeight                   = 10.0                    // BulletHeight is the height of the bullet (in px).
	BulletInitialDamage            = 30                      // BulletInitialDamage is the initial damage of the bullet.
	BulletModifierProgressStep     = 100                     // BulletModifierProgressStep is the progress step required to increase the bullet damage modifier.
	BulletSpeed                    = 7.0                     // BulletSpeed is the speed of the bullet (in px per frame).
	BulletWidth                    = 4.0                     // BulletWidth is the width of the bullet (in px).
	EnemiesCount                   = 10                      // EnemiesCount is the number of enemies on the canvas.
	EnemyBerserkLikeliness         = 0.01                    // EnemyBerserkLikeliness is the likeliness of an enemy to become a berserker.
	EnemyBerserkLikelinessProgress = 0.01                    // EnemyBerserkLikelinessProgress is the amount of berserk likeliness an enemy receives on progress.
	EnemyDefenseProgress           = 25                      // EnemyDefenseProgress is the amount of defense an enemy receives on progress.
	EnemyHeight                    = 40.0                    // EnemyHeight is the height of the enemy (in px).
	EnemyHitpointProgress          = 20                      // EnemyHitpointProgress is the amount of hit points an enemy receives on progress.
	EnemyInitialDefense            = 10                      // EnemyInitialDefense is the initial defense of the enemy.
	EnemyInitialHitpoints          = 100                     // EnemyInitialHitpoints is the initial hit points of the enemy.
	EnemyInitialSpeed              = 1.0                     // EnemyInitialSpeed is the initial speed of the enemy (in px per frame).
	EnemyMargin                    = 10.0                    // EnemyMargin is the margin between enemies (in px).
	EnemyMaximumSpeed              = 5.0                     // EnemyMaximumSpeed is the maximum speed of the enemy (in px per frame).
	EnemySpecialtyLikeliness       = 0.1                     // EnemySpecialtyLikeliness is the likeliness of an enemy to become a goodie or a freezer.
	EnemyWidth                     = 40.0                    // EnemyWidth is the width of the enemy (in px).
	SpaceshipAnnihilatorPenalty    = 81                      // SpaceshipAnnihilatorPenalty is the penalty of the spaceship when it collides with an annihilator.
	SpaceshipBerserkPenalty        = 9                       // SpaceshipBerserkPenalty is the penalty of the spaceship when it collides with a berserker.
	SpaceshipCannonProgress        = 10                      // SpaceshipCannonProgress is the amount of spaceship progress to unlock a new cannon.
	SpaceshipCooldown              = 100 * time.Millisecond  // SpaceshipCooldown is the cooldown of the spaceship (in ms).
	SpaceshipDefaultPenalty        = 3                       // SpaceshipDefaultPenalty is the default penalty of the spaceship.
	SpaceshipHeight                = 40.0                    // SpaceshipHeight is the height of the spaceship (in px).
	SpaceshipInitialSpeed          = 25.0                    // SpaceshipInitialSpeed is the initial speed of the spaceship (in px per frame).
	SpaceshipMaximumCannons        = 20                      // SpaceshipMaximumCannons is the maximum number of cannons the spaceship can have.
	SpaceshipMaximumSpeed          = 50.0                    // SpaceshipMaximumSpeed is the maximum speed of the spaceship (in px per frame).
	SpaceshipStateDuration         = 1500 * time.Millisecond // SpaceshipStateDuration is the duration of the spaceship state (in ms).
	SpaceshipWidth                 = 40.0                    // SpaceshipWidth is the width of the spaceship (in px).

)

const (
	MessageGameStartedNoTouchDevice = "Game started! Use ARROW KEYS (<, >) to move and SPACE to shoot."
	MessageHowToStartNoTouchDevice  = "Let's begin! Press any key to start."
	MessageGameStartedTouchDevice   = "Game started! Swipe left or right to move and tap to shoot."
	MessageHowToStartTouchDevice    = "Let's begin! Tap to start."
)
