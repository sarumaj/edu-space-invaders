package objects

import "fmt"

// Size represents the size of an object (Width, Height)
type Size struct {
	Width, Height float64
}

// String returns the string representation of the size.
func (size Size) String() string {
	return fmt.Sprintf("(%g, %g)", size.Width, size.Height)
}
