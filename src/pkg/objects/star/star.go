package star

import (
	"math/rand/v2"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// Star represents a star.
type Star struct {
	Position     objects.Position
	Radius       objects.Number
	CurrentScale objects.Position
	Spikes       objects.Number
	Exhausted    bool
	color        string
}

// Draw is a method that draws the star.
func (star Star) Draw() {
	if star.Exhausted {
		return
	}

	config.DrawStar(star.Position.Pack(), star.Spikes.Float(), star.Radius.Float(), star.color, config.Config.Star.Brightness)
}

// Exhaust is a method that sets the star as exhausted.
func (star *Star) Exhaust() {
	star.Exhausted = true
}

// Scale is a method that scales the star.
func (star *Star) Scale() {
	canvasDimensions := config.CanvasBoundingBox()
	scale := objects.Position{
		X: objects.Number(canvasDimensions.ScaleX),
		Y: objects.Number(canvasDimensions.ScaleY),
	}

	_ = objects.
		Measure(star.Position, star.Radius).
		Scale(objects.Ones().DivX(star.CurrentScale)).
		Scale(scale).
		ApplyPosition(&star.Position).
		ApplySize(&star.Radius)

	star.CurrentScale = scale
}

// Twinkle is a function that creates a new star.
func Twinkle(position objects.Position) *Star {
	star := Star{
		Position:     position,
		Radius:       objects.Number(rand.Float64()*config.Config.Star.MinimumRadius + (config.Config.Star.MaximumRadius - config.Config.Star.MinimumRadius)),
		Spikes:       objects.Number(rand.Float64()*config.Config.Star.MinimumSpikes + (config.Config.Star.MaximumSpikes - config.Config.Star.MinimumSpikes)),
		CurrentScale: objects.Ones(),
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

	star.Scale()
	return &star
}
