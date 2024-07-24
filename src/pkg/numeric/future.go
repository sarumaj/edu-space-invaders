//go:build future

package numeric

// Area returns the area of the rectangle.
// The area is calculated by dividing the rectangle into two triangles and summing their areas.
// Shoelace Formula is used to calculate the area of a polygon.
func (rect Rectangle) Area() Number { return ShoelaceFormula(rect[:]) }

// Area returns the area of the triangle.
func (tri Triangle) Area() Number { return ShoelaceFormula(tri[:]) }

// Area returns the area of the spaceship polygon.
func (poly SpaceshipPolygon) Area() Number {
	// The spaceship polygon is divided into two parts: the main body and the wings.
	// The main body is a rectangle, and the wings are triangles.
	// The area of the spaceship polygon is the sum of the areas of the main body and the wings.
	main, leftWing, rightWing, head := poly.Decompose()
	return main.Area() + leftWing.Area() + rightWing.Area() + head.Area()
}

// Decompose decomposes the spaceship polygon into the main body, the wings, and the head.
func (poly SpaceshipPolygon) Decompose() (body Rectangle, leftWing Triangle, rightWing Triangle, head Triangle) {
	body = Rectangle{poly[0], poly[1], poly[2], poly[3]}
	leftWing = Triangle{poly[0], poly[3], poly[4]}
	rightWing = Triangle{poly[1], poly[2], poly[5]}
	head = Triangle{poly[0], poly[1], poly[6]}

	return
}

// ShoelaceFormula calculates the area of a polygon using the Shoelace Formula.
func ShoelaceFormula(vertices []Position) Number {
	var sum Number
	for i := range vertices {
		sum += vertices[i].Cross(vertices[(i+1)%len(vertices)])
	}

	return sum.Abs() / 2
}
