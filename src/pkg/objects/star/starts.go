package star

import (
	"math"
	"math/rand/v2"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

type Stars []Star

// Update is a method that updates the stars.
func (stars *Stars) Update(spaceshipSpeed float64) {
	for i := range *stars {
		star := &(*stars)[i]
		star.Accelerate(spaceshipSpeed)
		star.Move()
	}
}

// Explode is a function that creates a number of stars.
// It creates a grid of cells and places stars in random positions within these cells.
// The number of stars is determined by the input parameter.
func Explode(num int) Stars {
	// Define the grid size
	gridSize := math.Sqrt((config.CanvasWidth() * config.CanvasHeight()) / float64(num))
	cols := int(config.CanvasWidth() / gridSize)
	rows := int(config.CanvasHeight() / gridSize)

	// Create a grid of cells and place stars in random positions within these cells
	var stars Stars
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			if num <= 0 {
				return stars
			}

			x := (float64(col) + rand.Float64()) * gridSize
			y := (float64(row) + rand.Float64()) * gridSize
			stars = append(stars, *Twinkle(objects.Position{
				X: objects.Number(x),
				Y: objects.Number(y),
			}))

			num--
		}
	}

	return stars
}
