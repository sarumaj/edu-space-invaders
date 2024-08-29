package numeric

import (
	"math"
	"slices"
)

type Vertices []Position

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

// Axes returns the axes of a polygon.
// The axes are the normals to the edges of the polygon.
// The normals are the perpendicular vectors to the edges.
// It assumes that the vertices are sorted in clockwise or counter-clockwise order.
// If other is supplied, the axes are the normals to the edges of the two polygons.
func (vertices Vertices) Axes(other Vertices) []Position {
	axes := make([]Position, 0, len(vertices)+len(other))

	for i := 0; i < len(vertices)+len(other); i++ {
		var edge Position
		if i < len(vertices) {
			edge = vertices[i].Sub(vertices[(i+1)%len(vertices)])
		} else {
			edge = other[i-len(vertices)].Sub(other[(i-len(vertices)+1)%len(other)])
		}
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

// HasSeparatingAxis checks if an object collides with another using the Separating Axis Theorem.
// The axes to test are the normals to the edges of the two objects.
// If there is a separating axis, there is no collision.
// It assumes that the two objects are convex.
func (vertices Vertices) HasSeparatingAxis(other Vertices) bool {
	if vertices.Len() == 0 || other.Len() == 0 {
		return false
	}

	// Check for overlap on all axes
	for _, axis := range vertices.Axes(other) {
		axis = axis.Normalize() // Ensure the axis is normalized

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

// MaximumTranslationVector calculates the minimum translation vector to separate two objects.
// It uses the Separating Axis Theorem to check for collision.
// It assumes that the two objects are convex.
func (vertices Vertices) MinimumTranslationVector(other Vertices) (mtv Position) {
	minOverlap := Number(math.MaxFloat64)

	// Check for overlap on all axes
	for _, axis := range vertices.Axes(other) {
		axis = axis.Normalize() // Ensure the axis is normalized

		minA, maxA := axis.Project(vertices)
		minB, maxB := axis.Project(other)

		if minA > maxB || minB > maxA {
			// There is a separating axis, no collision
			return
		}

		// Calculate the overlap
		if overlap := maxA.Min(maxB) - minA.Max(minB); overlap < minOverlap {
			minOverlap = overlap

			mtv = axis.Mul(overlap)
			if minA < minB { // Change the direction of the MTV in the opposite direction of the axis
				mtv = mtv.Mul(-1)
			}
		}
	}

	return mtv
}
