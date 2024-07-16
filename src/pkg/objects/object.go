package objects

// Object is an auxiliary struct that represents an object.
type Object[P interface{ Number | Size }] struct {
	Position Position
	Size     P
}

// ApplyPosition applies the position of the object to the specified position.
func (object Object[P]) ApplyPosition(position *Position) *Object[P] {
	if position == nil {
		return &object
	}

	*position = object.Position
	return &object
}

// ApplySize applies the size of the object to the specified size.
func (object Object[P]) ApplySize(size *P) *Object[P] {
	if size == nil {
		return &object
	}

	*size = object.Size
	return &object
}

// Scale scales the object by the specified scales.
func (object *Object[P]) Scale(scales Position) *Object[P] {
	switch size := any(object.Size).(type) {
	case Size:
		newSize := size.ToVector().MulX(scales).ToBox()
		if newSize.ToVector().Equal(Zeroes()) {
			return object
		}

		object.Size = any(newSize).(P)

	case Number:
		newSize := size * scales.Average()
		if newSize.Equal(0) {
			return object
		}

		object.Size = any(newSize).(P)
	}

	object.Position = object.Position.MulX(scales)
	return object
}

// Equal checks if the two objects are equal.
func Equal[P interface{ Number | Position | Size }](a, b P, tolerance Number) bool {
	switch a := any(a).(type) {
	case Number:
		return (a - any(b).(Number)).Abs() <= tolerance

	case Position:
		b := any(b).(Position)
		return (a.X-b.X).Abs() <= tolerance && (a.Y-b.Y).Abs() <= tolerance

	case Size:
		b := any(b).(Size)
		return Equal(a.ToVector(), b.ToVector(), tolerance)

	}

	return false
}

// Measure creates a new object with the specified position, size and radius.
func Measure[Perimeter interface{ Number | Size }](pos Position, size Perimeter) *Object[Perimeter] {
	return &Object[Perimeter]{
		Position: pos,
		Size:     size,
	}
}
