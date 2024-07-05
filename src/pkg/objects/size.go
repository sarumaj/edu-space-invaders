package objects

import "fmt"

// Size represents the size of an object (Width, Height)
type Size struct {
	Width, Height int
}

// String returns the string representation of the size.
func (size Size) String() string {
	return fmt.Sprintf("(%d, %d)", size.Width, size.Height)
}
