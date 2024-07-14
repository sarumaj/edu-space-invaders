package star

import (
	"math/rand/v2"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

type Stars []Star

// Explode is a function that creates a number of stars.
// It creates a grid of cells and places stars in random positions within these cells.
// The number of stars is determined by the input parameter.
func Explode(num int) Stars {
	// Define the grid size
	canvasDimensions := config.CanvasBoundingBox()
	gridSize := objects.Number((canvasDimensions.Width * canvasDimensions.Height) / float64(num)).Root()
	newBox := objects.Position{
		X: objects.Number(canvasDimensions.Width),
		Y: objects.Number(canvasDimensions.Height),
	}.Div(gridSize).ToBox()

	// Create a grid of cells and place stars in random positions within these cells
	var stars Stars
	for row := objects.Number(0); row < newBox.Height; row++ {
		for col := objects.Number(0); col < newBox.Width; col++ {
			if num <= 0 {
				return stars
			}

			stars = append(stars, *Twinkle(objects.Position{
				X: col + objects.Number(rand.Float64()),
				Y: row + objects.Number(rand.Float64()),
			}.Mul(objects.Number(gridSize))))

			num--
		}
	}

	return stars
}
