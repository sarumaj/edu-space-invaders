package numeric

import (
	"fmt"
)

// Size represents the size of an object (Width, Height)
type Size struct {
	Width, Height Number
	Scale         Number
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

// Half returns the half of the size.
func (size Size) Half() Size {
	return Size{
		Width:  size.Width / 2,
		Height: size.Height / 2,
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

// Resize resizes the size by the scale and position.
// The position is the center of the new size.
// The function returns the new size and adjusted position.
func (size Size) Resize(scale Number, position Position) (Size, Position) {
	newSize := Size{
		Width:  size.Width * scale,
		Height: size.Height * scale,
		Scale:  scale,
	}

	if size.Scale > 0 {
		newSize.Scale = size.Scale * scale
	}

	position.X -= (newSize.Width - size.Width) / 2
	position.Y -= (newSize.Height - size.Height) / 2

	return newSize, position
}

// Restore restores the size and position to the original state.
func (size Size) Restore(position Position) (Size, Position) {
	if size.Scale > 0 && size.Scale != 1 {
		return size.Resize(1/size.Scale, position)
	}

	return size, position
}

// String returns the string representation of the size.
func (size Size) String() string {
	return fmt.Sprintf("(%g, %g) / %g", size.Width, size.Height, size.Scale)
}

// ToVector returns the size as a position.
func (size Size) ToVector() Position {
	return Position{
		X: size.Width,
		Y: size.Height,
	}
}
