package star

import (
	"math/rand/v2"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// Star represents a star.
type Star struct {
	Position  objects.Position
	Radius    float64
	Spikes    float64
	Exhausted bool
	color     string
}

// Draw is a method that draws the star.
func (s *Star) Draw() {
	if s.Exhausted {
		return
	}
	config.DrawStar(s.Position.Pack(), s.Spikes, s.Radius, s.color, config.Config.Star.Brightness)
	s.Exhaust()
}

// Exhaust is a method that sets the star as exhausted.
func (s *Star) Exhaust() {
	s.Exhausted = true
}

// Twinkle is a function that creates a new star.
func Twinkle(position objects.Position) *Star {
	return &Star{
		Position: position,
		Radius:   rand.Float64()*config.Config.Star.MinimumRadius + (config.Config.Star.MaximumRadius - config.Config.Star.MinimumRadius),
		Spikes:   rand.Float64()*config.Config.Star.MinimumSpikes + (config.Config.Star.MaximumSpikes - config.Config.Star.MinimumSpikes),
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
