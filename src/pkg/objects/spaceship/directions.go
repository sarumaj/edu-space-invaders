package spaceship

import (
	"fmt"

	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

const (
	Unknown Direction = iota
	Up
	Down
	Left
	Right
)

// Direction represents the direction of the spaceship.
type Direction int

// Opposite returns the opposite direction.
func (d Direction) Opposite() Direction {
	switch d {
	case Up:
		return Down
	case Down:
		return Up
	case Left:
		return Right
	case Right:
		return Left
	default:
		return Unknown
	}
}

// String returns the string representation of the direction.
func (d Direction) String() string {
	return [...]string{"Unknown", "Up", "Down", "Left", "Right"}[d]
}

// Directions represents the directions of the spaceship.
type Directions struct {
	Horizontal Direction // Horizontal Direction, either Left or Right
	Vertical   Direction // Vertical Direction, either Up or Down
}

// Brake stops the spaceship if it is headed to the opposite direction.
func (d Directions) Brake(delta numeric.Position) numeric.Position {
	speedFactor := numeric.Ones()
	switch {
	case
		delta.Y > 0 && d.Vertical == Up,
		delta.Y < 0 && d.Vertical == Down:

		speedFactor.Y = 0

	}

	switch {
	case
		delta.X > 0 && d.Horizontal == Left,
		delta.X < 0 && d.Horizontal == Right:

		speedFactor.X = 0

	}

	return speedFactor
}

// IsHeadedTo returns true if the spaceship is headed to the specified direction.
func (d Directions) IsHeadedTo(dir Direction) bool {
	switch dir {
	case Up, Down:
		return d.Vertical == dir

	case Left, Right:
		return d.Horizontal == dir

	default:
		return false
	}
}

// SetFromDelta sets the directions based on the delta.
func (d *Directions) SetFromDelta(delta numeric.Position) {
	switch {
	case delta.X < 0:
		d.Horizontal = Left

	case delta.X > 0:
		d.Horizontal = Right

	default:
		d.Horizontal = Unknown
	}

	switch {
	case delta.Y < 0:
		d.Vertical = Up

	case delta.Y > 0:
		d.Vertical = Down

	default:
		d.Vertical = Unknown
	}

}

// SetHorizontal sets the horizontal direction.
func (d *Directions) SetHorizontal(dir Direction) {
	switch dir {
	case Left, Right:
		d.Horizontal = dir
	}
}

// SetVertical sets the vertical direction.
func (d *Directions) SetVertical(dir Direction) {
	switch dir {
	case Up, Down:
		d.Vertical = dir
	}
}

// String returns the string representation of the directions.
func (d Directions) String() string {
	return fmt.Sprintf("Horizontal: %s, Vertical: %s", d.Horizontal, d.Vertical)
}
