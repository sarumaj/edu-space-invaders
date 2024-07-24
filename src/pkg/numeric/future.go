//go:build future

package numeric

// Area returns the area of the rectangle.
// The area is calculated by dividing the rectangle into two triangles and summing their areas.
// Shoelace Formula is used to calculate the area of a polygon.
func (rect Rectangle) Area(sort bool) Number { return ShoelaceFormula(rect[:], sort) }

// Area returns the area of the triangle.
func (tri Triangle) Area(sort bool) Number { return ShoelaceFormula(tri[:], sort) }

// Area returns the area of the spaceship polygon.
func (poly SpaceshipPolygon) Area(sort bool) Number {
	return ShoelaceFormula(poly[:], sort)
}

// Decompose decomposes the spaceship polygon into the main body, the wings, and the head.
func (poly SpaceshipPolygon) Decompose(sort bool) (body Rectangle, leftWing Triangle, rightWing Triangle, head Triangle) {
	for i, part := range SortVerticesClockwise(poly[:]) {
		poly[i] = part
	}

	body = Rectangle{poly[0], poly[1], poly[2], poly[3]}
	leftWing = Triangle{poly[0], poly[3], poly[4]}
	rightWing = Triangle{poly[1], poly[2], poly[5]}
	head = Triangle{poly[0], poly[1], poly[6]}

	return
}

// ShoelaceFormula calculates the area of a polygon using the Shoelace Formula.
func ShoelaceFormula(vertices []Position, sort bool) Number {
	// Sort the vertices in clockwise order
	// The order only matters if there are more than 3 vertices
	if sort && len(vertices) > 3 {
		vertices = SortVerticesClockwise(vertices)
	}

	var sum Number
	for i := range vertices {
		product := vertices[i].Cross(vertices[(i+1)%len(vertices)])
		sum += product
	}

	return sum.Abs() / 2
}
