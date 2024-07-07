package objects

import "fmt"

// Size represents the size of an object (Width, Height)
type Size struct {
	Width, Height Number
}

// Pack returns the packed representation of the size.
func (size Size) Pack() [2]float64 {
	return [2]float64{size.Width.Float(), size.Height.Float()}
}

// String returns the string representation of the size.
func (size Size) String() string {
	return fmt.Sprintf("(%g, %g)", size.Width, size.Height)
}
