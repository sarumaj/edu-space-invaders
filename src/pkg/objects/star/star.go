package star

import (
	"math/rand/v2"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// Star represents a star.
type Star struct {
	Position objects.Position
	Radius   float64
	Spikes   float64
	Speed    float64
	color    string
}

// Accelerate is a method that accelerates the star.
func (s *Star) Accelerate(spaceshipSpeed float64) {
	// The star's speed is proportional to the spaceship's speed.
	s.Speed = spaceshipSpeed * config.Config.Star.SpeedRatio
}

// Draw is a method that draws the star.
func (s *Star) Draw() {
	config.DrawStar(s.Position.Pack(), s.Spikes, s.Radius, s.Radius/2, s.color, config.Config.Star.Brightness)
}

// Move is a method that moves the star.
func (s *Star) Move() {
	s.Position.Y += objects.Number(s.Speed)
	if s.Position.Y.Float() > config.CanvasHeight() {
		*s = *Twinkle(objects.Position{X: s.Position.X, Y: 0})
	}
}

// Twinkle is a function that creates a new star.
func Twinkle(pos objects.Position) *Star {
	return &Star{
		Position: pos,
		Radius:   rand.Float64()*config.Config.Star.MinimumRadius + (config.Config.Star.MaximumRadius - config.Config.Star.MinimumRadius),
		Spikes:   rand.Float64()*config.Config.Star.MinimumSpikes + (config.Config.Star.MaximumSpikes - config.Config.Star.MinimumSpikes),
		Speed:    config.Config.Spaceship.InitialSpeed * config.Config.Star.SpeedRatio,
		color: [...]string{
			"white",
			"LightYellow",
			"PaleGoldenrod",
			"LightCyan",
			"SkyBlue",
			"LightSteelBlue",
			"LightSalmon",
			"LightCoral",
			"LightPink",
			"LavenderBlush",
		}[rand.IntN(10)],
	}
}
