package handler

import (
	"fmt"

	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// touchEvent represents a touch event.
type touchEvent struct {
	Position    objects.Position
	Delta       objects.Position
	IsDoubleTap bool
}

// CalculateDelta calculates the delta difference of the touch event.
func (t *touchEvent) CalculateDelta(x, y float64) {
	t.Delta = objects.Position{
		X: objects.Number(x) - t.Position.X,
		Y: objects.Number(y) - t.Position.Y,
	}
}

// String returns the string representation of the touch event.
func (t touchEvent) String() string {
	return fmt.Sprintf("Touch (Pos: %s, Delta: %s)", t.Position, t.Delta)
}
