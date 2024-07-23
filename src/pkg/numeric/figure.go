package numeric

import "fmt"

var _ Figure = Rectangle{}
var _ Figure = Triangle{}

// Figure represents a geometric figure.
type Figure interface {
	Slice() []Position
	String() string
}

// Rectangle represents a rectangle.
type Rectangle [4]Position

// Slice returns the slice of positions.
func (rect Rectangle) Slice() []Position { return rect[:] }

func (rect Rectangle) String() string {
	return fmt.Sprintf("Rectangle{%v, %v, %v, %v}", rect[0], rect[1], rect[2], rect[3])
}

// Triangle represents a triangle.
type Triangle [3]Position

// Slice returns the slice of positions.
func (tri Triangle) Slice() []Position { return tri[:] }

// String returns the string representation of the triangle.
func (tri Triangle) String() string {
	return fmt.Sprintf("Triangle{%v, %v, %v}", tri[0], tri[1], tri[2])
}
