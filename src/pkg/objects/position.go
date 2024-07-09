package objects

import (
	"fmt"
	"math"
)

// Position represents the position of an object (X, Y)
type Position struct {
	X, Y Number
}

// Add adds two positions.
func (pos Position) Add(other Position) Position {
	return Position{
		X: pos.X + other.X,
		Y: pos.Y + other.Y,
	}
}

// AddN adds a number to a position.
func (pos Position) AddN(n Number) Position {
	return Position{
		X: pos.X + n,
		Y: pos.Y + n,
	}
}

// Distance returns the Euclidean distance between two position.
func (pos Position) Distance(other Position) Number {
	return pos.Sub(other).Magnitude()
}

// Div divides a position by a number.
func (pos Position) Div(other Number) Position {
	if other == 0 {
		return Position{}
	}

	return Position{
		X: pos.X / other,
		Y: pos.Y / other,
	}
}

// Equal checks if two positions are equal.
func (pos Position) Equal(other Position) bool {
	return pos.X == other.X && pos.Y == other.Y
}

// Greater checks if a position is greater than another.
func (pos Position) Greater(other Position) bool {
	return pos.X > other.X && pos.Y > other.Y
}

// GreaterOrEqual checks if a position is greater or equal to another.
func (pos Position) GreaterOrEqual(other Position) bool {
	return pos.X >= other.X && pos.Y >= other.Y
}

// IsZero checks if the position is zero.
func (pos Position) IsZero() bool {
	return pos.X == 0 && pos.Y == 0
}

// Less checks if a position is less than another.
func (pos Position) Less(other Position) bool {
	return pos.X < other.X && pos.Y < other.Y
}

// LessOrEqual checks if a position is less or equal to another.
func (pos Position) LessOrEqual(other Position) bool {
	return pos.X <= other.X && pos.Y <= other.Y
}

// Magnitude returns the magnitude of the position.
func (pos Position) Magnitude() Number {
	return (pos.X*pos.X + pos.Y*pos.Y).Root()
}

// Mul multiplies a position by a number.
func (pos Position) Mul(other Number) Position {
	return Position{
		X: pos.X * other,
		Y: pos.Y * other,
	}
}

// Normalize returns the normalized position.
func (pos Position) Normalize() Position {
	mag := pos.Magnitude()
	if mag == 0 {
		return Position{}
	}

	return Position{
		X: pos.X / mag,
		Y: pos.Y / mag,
	}
}

// Pack returns the packed representation of the position.
func (pos Position) Pack() [2]float64 {
	return [2]float64{pos.X.Float(), pos.Y.Float()}
}

// Root returns the square root of the position.
func (pos Position) Root() Number {
	return Number(math.Sqrt(float64(pos.X * pos.Y)))
}

// String returns the string representation of the position.
func (pos Position) String() string {
	return fmt.Sprintf("(%g, %g)", pos.X, pos.Y)
}

// Sub subtracts two positions.
func (pos Position) Sub(other Position) Position {
	return Position{
		X: pos.X - other.X,
		Y: pos.Y - other.Y,
	}
}

func (pos Position) SubN(n Number) Position {
	return Position{
		X: pos.X - n,
		Y: pos.Y - n,
	}
}

// ToBox returns the position as a size.
func (pos Position) ToBox() Size {
	return Size{
		Width:  pos.X,
		Height: pos.Y,
	}
}
