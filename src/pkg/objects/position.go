package objects

import "fmt"

// Position represents the position of an object (X, Y)
type Position struct {
	X, Y Number
}

// Pack returns the packed representation of the position.
func (pos Position) Pack() [2]float64 {
	return [2]float64{pos.X.Float(), pos.Y.Float()}
}

// String returns the string representation of the position.
func (pos Position) String() string {
	return fmt.Sprintf("(%g, %g)", pos.X, pos.Y)
}
