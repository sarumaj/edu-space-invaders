package planet

import (
	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

// PlanetType represents the type of a planet.
type Planet struct {
	Position numeric.Position
	Radius   numeric.Number
	Type     PlanetType
}

// Draw draws the planet on the canvas.
func (planet Planet) Draw() { planet.Type.Draw(planet.Position, planet.Radius) }

// Patch patches the planet with the new planet.
// It updates the position, radius, and type of the planet.
func (planet *Planet) Patch(newPlaner *Planet) {
	planet.Position = newPlaner.Position
	planet.Radius = newPlaner.Radius
	planet.Type = newPlaner.Type
}

// Update updates the planet position.
// The speed parameter is the speed at which the planet moves.
// If the planet reaches the bottom of the canvas, it will be reborn.
func (planet *Planet) Update(speed numeric.Number) {
	planet.Position.Y += speed
	canvasDimensions := config.CanvasBoundingBox()
	if (planet.Position.Y - planet.Radius).Float() > canvasDimensions.OriginalHeight {
		planet.Patch(Reveal(false))
	}
}

// Reveal reveals a new planet.
// If randomY is true, the planet will be revealed at a random Y position.
// Otherwise, the planet will be revealed at the top of the canvas.
// The planet will have a random radius and type.
// The planet will be exhausted prematurely based on the reveal probability.
// This way, the planet will not be drawn on the canvas.
func Reveal(randomY bool) *Planet {
	canvasDimensions := config.CanvasBoundingBox()
	planet := &Planet{
		Position: numeric.Locate(numeric.RandomRange(0, canvasDimensions.OriginalWidth), 0),
		Radius:   numeric.RandomRange(config.Config.Planet.MinimumRadius, config.Config.Planet.MaximumRadius),
		Type:     PlanetType(numeric.RandomRange(0, 7).Int()),
	}

	planet.Position.Y = -planet.Radius
	if randomY {
		planet.Position.Y = numeric.RandomRange(0, canvasDimensions.OriginalHeight)
	}

	return planet
}
