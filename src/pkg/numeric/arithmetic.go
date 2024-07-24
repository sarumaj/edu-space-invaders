package numeric

import (
	"math/rand/v2"
	"slices"
)

// sortVerticesClockwise returns a function that sorts the vertices of a polygon in clockwise order.
// The centroid is the center of the polygon.
// To be used with slices.SortFunc.
func sortVerticesClockwise(centroid Position) func(i, j Position) int {
	return func(i, j Position) int {
		switch angleI, angleJ := centroid.AngleTo(i), centroid.AngleTo(j); {
		case
			angleI > angleJ, // Largest angle first
			angleI == angleJ && centroid.Distance(i) < centroid.Distance(j): // Same angle, sort by distance, closest first

			return -1

		default:
			return 1

		}
	}
}

// AreVerticesSortedClockwise checks if the vertices of a polygon are sorted in clockwise order.
func AreVerticesSortedClockwise(vertices []Position) bool {
	return slices.IsSortedFunc(vertices, sortVerticesClockwise(Centroid(vertices)))
}

// Centroid calculates the centroid of a polygon.
func Centroid(vertices []Position) Position {
	var sum Position
	for _, vertex := range vertices {
		sum = sum.Add(vertex)
	}

	return sum.Div(Number(len(vertices)))
}

// Equal checks if the two objects are equal.
func Equal[P interface {
	Number | Position | Size
}](a, b P, tolerance Number) bool {
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

// HaveSeparatingAxis checks if two objects collide using the Separating Axis Theorem.
// The axes to test are the normals to the edges of the two objects.
// If there is a separating axis, there is no collision.
// It assumes that the two objects are convex.
func HaveSeparatingAxis(verticesA, verticesB []Position, sort bool) bool {
	if len(verticesA) == 0 || len(verticesB) == 0 {
		return false
	}

	// Sort the vertices in clockwise order
	// The order only matters if there are more than 3 vertices
	if sort && len(verticesA) > 3 {
		verticesA = SortVerticesClockwise(verticesA)
	}

	if sort && len(verticesB) > 3 {
		verticesB = SortVerticesClockwise(verticesB)
	}

	// Calculate the normals to the edges of the two objects
	axes := make([]Position, 0, len(verticesA)+len(verticesB))

	// Add the normals to the edges of the first object
	for i := 0; i < len(verticesA); i++ {
		edge := verticesA[i].Sub(verticesA[(i+1)%len(verticesA)])
		axes = append(axes, edge.Perpendicular())
	}

	// Add the normals to the edges of the second object
	for i := 0; i < len(verticesB); i++ {
		edge := verticesB[i].Sub(verticesB[(i+1)%len(verticesB)])
		axes = append(axes, edge.Perpendicular())
	}

	// Check for overlap on all axes
	for _, axis := range axes {
		minA, maxA := axis.Project(verticesA)
		minB, maxB := axis.Project(verticesB)

		if minA > maxB || minB > maxA {
			// There is a separating axis, no collision
			return true
		}
	}

	// No separating axis found, there is a collision
	return false
}

// RandomRange returns a random number between min and max.
func RandomRange[Numeric1, Numeric2 interface{ ~float64 | ~int }](min Numeric1, max Numeric2) Number {
	return Number(min) + Number(rand.Float64())*(Number(max)-Number(min))
}

// SampleUniform returns true with the given probability.
func SampleUniform[Numeric interface{ ~float64 | ~int }](probability Numeric) bool {
	if probability <= 0 {
		return false
	}

	if probability >= 1 {
		return true
	}

	return rand.Float64() < float64(probability)
}

// SortVerticesClockwise sorts the vertices of a polygon in clockwise order.
func SortVerticesClockwise(vertices []Position) []Position {
	slices.SortFunc(vertices, sortVerticesClockwise(Centroid(vertices)))

	return vertices
}
