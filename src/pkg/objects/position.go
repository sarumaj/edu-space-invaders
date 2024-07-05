package objects

import "fmt"

// Position represents the position of an object (X, Y)
type Position struct {
	X, Y int
}

// String returns the string representation of the position.
func (pos Position) String() string {
	return fmt.Sprintf("(%d, %d)", pos.X, pos.Y)
}
