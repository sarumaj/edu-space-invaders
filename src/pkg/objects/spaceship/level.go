package spaceship

import (
	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// SpaceshipLevel represents the spaceship level.
type SpaceshipLevel struct {
	AccelerateRate objects.Number // AccelerateRate is the rate at which the spaceship accelerates.
	Progress       int            // Progress is the progress level of the spaceship.
	Cannons        int            // Cannons is the number of cannons the spaceship has.
}

// Down decreases the spaceship level.
// If the spaceship level is already at the minimum level, it does nothing.
// If the number of cannons is greater than 1, it decreases the number of cannons by 1.
func (lvl *SpaceshipLevel) Down() {
	if lvl.Progress == 0 {
		return
	}

	lvl.AccelerateRate += objects.Number(config.Config.Spaceship.Acceleration)

	if lvl.Cannons > 1 && (lvl.Progress-1)%config.Config.Spaceship.CannonProgress == 0 {
		lvl.Cannons -= 1
	}

	lvl.Progress--
}

// Up increases the spaceship level.
// If the number of cannons is less than the maximum number of cannons, it increases the number of cannons by 1.
func (lvl *SpaceshipLevel) Up() {
	if lvl.Cannons < config.Config.Spaceship.MaximumCannons && (lvl.Progress+1)%config.Config.Spaceship.CannonProgress == 0 {
		lvl.Cannons += 1
	}

	lvl.AccelerateRate -= objects.Number(config.Config.Spaceship.Acceleration)
	if lvl.AccelerateRate.Float() < config.Config.Spaceship.Acceleration {
		lvl.AccelerateRate = objects.Number(config.Config.Spaceship.Acceleration)
	}

	lvl.Progress++
}
