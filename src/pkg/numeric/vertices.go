package numeric

import "slices"

type Vertices []Position

// Axes returns the axes of a polygon.
// The axes are the normals to the edges of the polygon.
// The normals are the perpendicular vectors to the edges.
// It assumes that the vertices are sorted in clockwise or counter-clockwise order.
func (vertices Vertices) Axes() []Position {
	axes := make([]Position, 0, len(vertices))

	for i := 0; i < len(vertices); i++ {
		edge := vertices[i].Sub(vertices[(i+1)%len(vertices)])
		axes = append(axes, edge.Perpendicular())
	}

	return axes
}

// Area calculates the area of a polygon using the Shoelace Formula.
// It assumes that the vertices are sorted in clockwise or counter-clockwise order.
func (vertices Vertices) Area() Number {
	var sum Number
	for i := range vertices {
		product := vertices[i].Cross(vertices[(i+1)%len(vertices)])
		sum += product
	}

	return sum.Abs() / 2
}

// AreSorted checks if the vertices of a polygon are sorted in clockwise or counter-clockwise order.
func (vertices Vertices) AreSorted(clockwise bool) bool {
	return slices.IsSortedFunc(vertices, vertices.sortVerticesClockwise(clockwise))
}

// Centroid calculates the centroid of a polygon.
func (vertices Vertices) Centroid() Position {
	var sum Position
	for _, vertex := range vertices {
		sum = sum.Add(vertex)
	}

	return sum.Div(Number(len(vertices)))
}

// Len returns the number of vertices of a polygon.
func (vertices Vertices) Len() int {
	return len(vertices)
}

// Sort sorts the vertices of a polygon in clockwise or counter-clockwise order.
func (vertices Vertices) Sort(clockwise bool) Vertices {
	slices.SortFunc(vertices, vertices.sortVerticesClockwise(clockwise))
	return vertices
}

// sortVertices returns a function that sorts the vertices of a polygon in clockwise or counter-clockwise order.
// The centroid is the center of the polygon.
// To be used with slices.SortFunc.
func (vertices Vertices) sortVerticesClockwise(clockwise bool) func(i, j Position) int {
	sign, c := 1, vertices.Centroid()
	if !clockwise {
		sign *= -1
	}

	return func(i, j Position) int {
		switch angleI, angleJ := c.AngleTo(i), c.AngleTo(j); {
		case
			angleI > angleJ, // Largest angle first
			angleI == angleJ && c.Distance(i) < c.Distance(j): // Same angle, sort by distance, closest first

			return -1 * sign

		default:
			return 1 * sign

		}
	}
}

// HasSeparatingAxis checks if an object collides with another using the Separating Axis Theorem.
// The axes to test are the normals to the edges of the two objects.
// If there is a separating axis, there is no collision.
// It assumes that the two objects are convex.
func (vertices Vertices) HasSeparatingAxis(other Vertices) bool {
	if vertices.Len() == 0 || other.Len() == 0 {
		return false
	}

	// Calculate the normals to the edges of the two objects
	axes := make([]Position, 0, len(vertices)+len(other))

	// Add the normals to the edges of the first object
	axes = append(axes, vertices.Axes()...)
	axes = append(axes, other.Axes()...)

	// Check for overlap on all axes
	for _, axis := range axes {
		minA, maxA := axis.Project(vertices)
		minB, maxB := axis.Project(other)

		if minA > maxB || minB > maxA {
			// There is a separating axis, no collision
			return true
		}
	}

	// No separating axis found, there is a collision
	return false
}
