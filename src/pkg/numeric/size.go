package numeric

import "fmt"

// Size represents the size of an object (Width, Height)
type Size struct {
	Width, Height Number
}

// Area returns the area of the bounding box.
func (size Size) Area() Number { return size.Width * size.Height }

// AspectRatio returns the aspect ratio of the size.
func (size Size) AspectRatio() Number {
	if size.Height == 0 {
		return 0
	}

	return size.Width / size.Height
}

// Center returns the relative center of the bounding box.
func (size Size) Center() Position {
	return Position{
		X: size.Width / 2,
		Y: size.Height / 2,
	}
}

// Pack returns the packed representation of the size.
func (size Size) Pack() [2]float64 {
	return [2]float64{size.Width.Float(), size.Height.Float()}
}

// Radius returns the radius of a circle with equivalent area.
func (size Size) Radius() Number {
	return (size.Width * size.Height / Pi).Root()
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
