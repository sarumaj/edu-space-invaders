package objects

import "fmt"

// Size represents the size of an object (Width, Height)
type Size struct {
	Width, Height Number
}

// AspectRatio returns the aspect ratio of the size.
func (size Size) AspectRatio() Number {
	if size.Height == 0 {
		return 0
	}

	return size.Width / size.Height
}

func (size Size) Equal(other Size) bool {
	return size.Width == other.Width && size.Height == other.Height
}

// Pack returns the packed representation of the size.
func (size Size) Pack() [2]float64 {
	return [2]float64{size.Width.Float(), size.Height.Float()}
}

// String returns the string representation of the size.
func (size Size) String() string {
	return fmt.Sprintf("(%g, %g)", size.Width, size.Height)
}

// ToVector returns the size as a position.
func (size Size) ToVector() Position {
	return Position{
		X: size.Width,
		Y: size.Height,
	}
}
