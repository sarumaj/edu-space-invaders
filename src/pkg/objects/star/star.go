package star

import (
	"math/rand/v2"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// Star represents a star.
type Star struct {
	Position  objects.Position
	Radius    objects.Number
	Scales    objects.Position
	Spikes    objects.Number
	Exhausted bool
	color     string
}

// Draw is a method that draws the star.
func (s Star) Draw() {
	if s.Exhausted {
		return
	}

	config.DrawStar(s.Position.Pack(), s.Spikes.Float(), s.Radius.Float(), s.color, config.Config.Star.Brightness)
}

// Exhaust is a method that sets the star as exhausted.
func (s *Star) Exhaust() {
	s.Exhausted = true
}

// Scale is a method that scales the star.
func (s *Star) Scale(scales objects.Position) {
	_ = objects.
		Measure(s.Position, s.Radius).
		Scale(scales).
		ApplyPosition(&s.Position).
		ApplySize(&s.Radius)

	if scales.X != s.Scales.X && scales.Y != s.Scales.Y {
		s.Scales = scales
	}
}

// Twinkle is a function that creates a new star.
func Twinkle(position objects.Position) *Star {
	star := Star{
		Position: position,
		Radius:   objects.Number(rand.Float64()*config.Config.Star.MinimumRadius + (config.Config.Star.MaximumRadius - config.Config.Star.MinimumRadius)),
		Spikes:   objects.Number(rand.Float64()*config.Config.Star.MinimumSpikes + (config.Config.Star.MaximumSpikes - config.Config.Star.MinimumSpikes)),
		color: [...]string{
			"White",
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

	canvasDimensions := config.CanvasBoundingBox()
	star.Scales = objects.Position{
		X: objects.Number(canvasDimensions.ScaleX),
		Y: objects.Number(canvasDimensions.ScaleY),
	}
	star.Scale(objects.Position{
		X: objects.Number(canvasDimensions.ScaleX),
		Y: objects.Number(canvasDimensions.ScaleY),
	})

	return &star
}
