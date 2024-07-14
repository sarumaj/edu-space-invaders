package spaceship

import (
	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
)

// SpaceshipLevel represents the spaceship level.
type SpaceshipLevel struct {
	AccelerateRate objects.Number // AccelerateRate is the rate at which the spaceship accelerates.
	Cannons        int            // Cannons is the number of cannons the spaceship has.
	Experience     int            // Experience is the experience level of the spaceship.
	Progress       int            // Progress is the progress level of the spaceship.
}

// Down decreases the spaceship level.
// If the spaceship level is already at the minimum level, it does nothing.
// If the number of cannons is greater than 1, it decreases the number of cannons by 1.
// It decreases the accelerate rate by the acceleration value: x = -0.5 + sqrt(0.25 + x').
// It returns true if the spaceship level has decreased.
func (lvl *SpaceshipLevel) Down() bool {
	if config.Config.Control.GodMode.Get() {
		return false
	}

	if lvl.Progress == 0 {
		return false
	}

	lvl.AccelerateRate = -0.5 + (0.25 + lvl.AccelerateRate).Root()
	if lvl.AccelerateRate.Float() < config.Config.Spaceship.Acceleration {
		lvl.AccelerateRate = objects.Number(config.Config.Spaceship.Acceleration)
	}

	if lvl.Cannons > 1 && (lvl.Progress-1)%config.Config.Spaceship.CannonProgress == 0 {
		lvl.Cannons -= 1
	}

	lvl.Progress--
	return true
}

// GainExperience gains experience for the spaceship.
// It calculates the experience gain based on the enemy type and level using the penalty values.
// If the experience is greater than the required experience, it increases the spaceship level.
// The formula for the required experience is the logarithm of the spaceship progress multiplied by the experience factor.
// It returns true if the spaceship level has increased.
func (lvl *SpaceshipLevel) GainExperience(e enemy.Enemy) bool {
	// Calculate the base experience using the penalty values
	var base objects.Number
	switch e.Type {
	case enemy.Freezer:
		base = objects.Number(config.Config.Spaceship.FreezerPenalty)

	case enemy.Normal:
		base = objects.Number(config.Config.Spaceship.DefaultPenalty)

	case enemy.Berserker:
		base = objects.Number(config.Config.Spaceship.BerserkPenalty)

	case enemy.Annihilator:
		base = objects.Number(config.Config.Spaceship.AnnihilatorPenalty)

	}

	// Calculate the experience gain
	gain := (base * objects.Number(e.Level.Progress)).Int()

	// Formula for the required experience
	formula := func() int {
		return (objects.Number(lvl.Progress+1).Log() * objects.Number(config.Config.Spaceship.ExperienceFactor)).Int()
	}

	// Increase the experience
	lvl.Experience += gain

	// Check later if the spaceship level has increased
	currentLvl := lvl.Progress

	// Increase the spaceship level
	for lvl.Experience >= formula() {
		lvl.Up()
		lvl.Experience -= gain
	}

	// Return true if the spaceship level has increased
	return lvl.Progress > currentLvl
}

// Up increases the spaceship level.
// If the progress is a multiple of the cannon progress, it increases the number of cannons by 1.
// If the number of cannons is less than the maximum number of cannons, it increases the number of cannons by 1.
// It increases the accelerate rate by the acceleration value: x' = x*(1+x).
func (lvl *SpaceshipLevel) Up() {
	if lvl.Cannons < config.Config.Spaceship.MaximumCannons && (lvl.Progress+1)%config.Config.Spaceship.CannonProgress == 0 {
		lvl.Cannons += 1
	}

	lvl.AccelerateRate *= 1 + objects.Number(config.Config.Spaceship.Acceleration)
	if lvl.AccelerateRate.Float() > config.Config.Spaceship.MaximumSpeed {
		lvl.AccelerateRate = objects.Number(config.Config.Spaceship.MaximumSpeed * config.Config.Spaceship.Acceleration)
	}

	lvl.Progress++
}
