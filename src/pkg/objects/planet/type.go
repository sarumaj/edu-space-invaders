package planet

import (
	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

const (
	// Planets
	Mercury PlanetType = iota
	Venus
	Earth
	Mars
	Jupiter
	Saturn
	Uranus
	Neptune

	// Other
	Pluto
	Sun

	// Anomalies
	BlackHole
	Supernova
)

const PlanetsCount = int(Neptune + 1)

// PlanetType represents the type of a planet or other celestial object.
type PlanetType int

// AnyOf returns true if the planet type is any of the given types.
func (t PlanetType) AnyOf(types ...PlanetType) bool {
	for _, typ := range types {
		if t == typ {
			return true
		}
	}

	return false
}

// DrawFunc returns the draw function for the planet type.
func (t PlanetType) Draw(position numeric.Position, radius numeric.Number) {
	map[PlanetType]func([2]float64, float64){
		Mercury:   config.DrawPlanetMercury,
		Venus:     config.DrawPlanetVenus,
		Earth:     config.DrawPlanetEarth,
		Mars:      config.DrawPlanetMars,
		Jupiter:   config.DrawPlanetJupiter,
		Saturn:    config.DrawPlanetSaturn,
		Uranus:    config.DrawPlanetUranus,
		Neptune:   config.DrawPlanetNeptune,
		Pluto:     config.DrawPlanetPluto,
		Sun:       config.DrawSun,
		BlackHole: config.DrawAnomalyBlackHole,
		Supernova: config.DrawAnomalySupernova,
	}[t](position.Pack(), radius.Float())
}

// IsPlanet returns true if the planet type is a planet.
func (t PlanetType) IsPlanet() bool { return t <= Neptune }

// String returns the string representation of the planet type.
func (t PlanetType) String() string {
	return [...]string{"Mercury", "Venus", "Earth", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune", "Pluto", "Sun", "BlackHole", "Supernova"}[t]
}
