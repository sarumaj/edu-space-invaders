package spaceship

import (
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/graphics"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

const (
	Neutral  SpaceshipState = iota // Neutral is the default state
	Damaged                        // Damaged is the state when the spaceship is hit
	Boosted                        // Boosted is the state when the spaceship is upgraded
	Frozen                         // Frozen is the state when the spaceship is frozen
	Hijacked                       // Hijacked is the state when the spaceship is controlled by the enemy
)

// SpaceshipState represents the state of the spaceship (Neutral, Damaged, Boosted, Frozen)
type SpaceshipState int

// AnyOf returns true if the spaceship state is any of the given states.
func (state SpaceshipState) AnyOf(states ...SpaceshipState) bool {
	for _, s := range states {
		if state == s {
			return true
		}
	}

	return false
}

// GetColor returns the color of the spaceship based on its state.
func (state SpaceshipState) GetColor() graphics.Color {
	switch state {
	case Damaged:
		return graphics.Catalogue().Crimson()
	case Boosted:
		return graphics.Catalogue().Chartreuse()
	case Frozen:
		return graphics.Catalogue().DeepSkyBlue()
	case Hijacked:
		return graphics.Catalogue().OrangeRed()
	default:
		return graphics.Catalogue().Lavender()
	}
}

// GetDuration returns the duration of the spaceship state.
func (state SpaceshipState) GetDuration() time.Duration {
	switch state {
	case Damaged:
		return config.Config.Spaceship.DamageDuration
	case Boosted:
		return config.Config.Spaceship.BoostDuration
	case Frozen:
		return config.Config.Spaceship.FreezeDuration
	case Hijacked:
		return config.Config.Spaceship.HijackDuration
	default:
		return 0
	}
}

// GetScale returns the scale of the spaceship based on its state.
func (state SpaceshipState) GetScale() numeric.Number {
	switch state {
	case Boosted:
		return numeric.Number(config.Config.Spaceship.BoostScaleSizeFactor)
	default:
		return 1
	}
}

// String returns the string representation of the spaceship state.
func (state SpaceshipState) String() string {
	return [...]string{"Neutral", "Damaged", "Boosted", "Frozen", "Hijacked"}[state]
}
