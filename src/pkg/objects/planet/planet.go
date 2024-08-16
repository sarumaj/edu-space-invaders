package planet

import (
	"fmt"
	"sync"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

// PlanetType represents the type of a planet or other celestial body.
type Planet struct {
	Position       numeric.Position
	Radius         numeric.Number
	Type           PlanetType
	additionalMass numeric.Number
	once           *sync.Once
}

// Again reveals the planet yet again as new planet.
// The planet will be revealed at the top of the canvas.
func (planet *Planet) Again() {
	newPlanet := Reveal(false)
	planet.Position = newPlanet.Position
	planet.Radius = newPlanet.Radius
	planet.Type = newPlanet.Type
	planet.additionalMass = 0
	planet.once = &sync.Once{}
}

// ApplyGravity applies the gravitational force of the planet to the point.
// The center is the center of the object that the planet is attracting.
// The mass is the mass of the object that the planet is attracting.
// The additionalMass flag indicates whether the mass should contribute persistently to the planet's gravitational force.
// The function returns the new position of the point after applying the gravitational force.
func (planet *Planet) ApplyGravity(center numeric.Position, mass numeric.Number, additionalMass, reverse bool) numeric.Position {
	// Compute the distance between the planet and the point
	delta := planet.Position.Sub(center)
	distance := delta.Magnitude()

	// If the distance is zero, the point is already on the planet
	if numeric.Equal(distance, 0, 1e-3) {
		if additionalMass {
			planet.additionalMass += mass
		}

		return center
	}

	// The gravitational field strength
	fieldStrength := numeric.Number(config.Config.Planet.GravityStrength) * numeric.Pi * planet.Radius.Pow(2) * (mass + planet.additionalMass) / distance.Pow(2)
	switch planet.Type { // Black holes and supernovas have a stronger gravitational force
	case Supernova: // Push the point away from the supernova to distort the field
		fieldStrength *= -numeric.Number(config.Config.Planet.AnomalyGravityModifier)

	case BlackHole: // Pull the point towards the black hole to trap the player
		fieldStrength *= numeric.Number(config.Config.Planet.AnomalyGravityModifier)

	}

	if reverse {
		fieldStrength *= -1
	}

	// Ensure that fieldStrength does not exceed the distance
	if fieldStrength > distance {
		fieldStrength = distance // Clamp the movement to exactly reach the planet's position
	}

	// Update the position
	return center.Add(delta.Normalize().Mul(fieldStrength))
}

// DoOnce executes the action only once during the lifetime of the planet.
func (planet *Planet) DoOnce(action func()) { planet.once.Do(action) }

// Draw draws the planet on the canvas.
// The planet will be drawn at the specified position with the specified radius.
// The planet will be drawn with the specified type.
func (planet Planet) Draw() { planet.Type.Draw(planet.Position, planet.Radius) }

// String returns the string representation of the planet.
func (planet Planet) String() string {
	return fmt.Sprintf("Planet{%s, R:%f, %s}", planet.Position, planet.Radius, planet.Type)
}

// Update updates the planet position.
// The speed parameter is the speed at which the planet moves.
// If the planet reaches the bottom of the canvas, it will be reborn.
func (planet *Planet) Update(speed numeric.Number) {
	planet.Position.Y += speed

	canvasDimensions := config.CanvasBoundingBox()
	if (planet.Position.Y - planet.Radius).Float() > canvasDimensions.OriginalHeight {
		planet.Again()
	}
}

// WithinRange returns true if the position is within the planet's range.
func (planet Planet) WithinRange(center numeric.Position) bool {
	return planet.Position.Sub(center).Magnitude() < planet.Radius
}

// Reveal reveals a new planet.
// If randomY is true, the planet will be revealed at a random Y position.
// Otherwise, the planet will be revealed at the top of the canvas.
// The planet will have a random radius and type.
func Reveal(randomY bool) *Planet {
	canvasDimensions := config.CanvasBoundingBox()
	planet := &Planet{
		Position: numeric.Locate(numeric.RandomRange(0, canvasDimensions.OriginalWidth), 0),
		Radius:   numeric.RandomRange(config.Config.Planet.MinimumRadius, config.Config.Planet.MaximumRadius),
		Type:     PlanetType(numeric.RandomRange(Mercury, Sun).Int()),
		once:     &sync.Once{},
	}

	planet.Position.Y = -planet.Radius
	if randomY {
		planet.Position.Y = numeric.RandomRange(0, canvasDimensions.OriginalHeight)
	}

	return planet
}
