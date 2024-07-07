package spaceship

import "github.com/sarumaj/edu-space-invaders/src/pkg/config"

// SpaceshipLevel represents the spaceship level.
type SpaceshipLevel struct {
	Progress int     // Progress is the progress level of the spaceship.
	Speed    float64 // Speed is the speed of the spaceship
	Cannons  int     // Cannons is the number of cannons the spaceship has.
}

// Down decreases the spaceship level.
// If the spaceship level is already at the minimum level, it does nothing.
// If the spaceship speed is greater than the initial speed, it decreases the speed by 1.
// If the number of cannons is greater than 1, it decreases the number of cannons by 1.
func (lvl *SpaceshipLevel) Down() {
	if lvl.Progress == 1 {
		return
	}

	if lvl.Speed > config.Config.Spaceship.InitialSpeed {
		lvl.Speed -= 1
	}

	if lvl.Cannons > 1 && (lvl.Progress-1)%config.Config.Spaceship.CannonProgress == 0 {
		lvl.Cannons -= 1
	}

	lvl.Progress--
}

// Up increases the spaceship level.
// If the spaceship speed is less than the maximum speed, it increases the speed by 1.
// If the number of cannons is less than the maximum number of cannons, it increases the number of cannons by 1.
func (lvl *SpaceshipLevel) Up() {
	if lvl.Speed < config.Config.Spaceship.MaximumSpeed {
		lvl.Speed += 1
	}

	if lvl.Cannons < config.Config.Spaceship.MaximumCannons && (lvl.Progress+1)%config.Config.Spaceship.CannonProgress == 0 {
		lvl.Cannons += 1
	}

	lvl.Progress++
}
