package numeric

import (
	"testing"
)

func figureToSlice[F interface {
	Triangle | Rectangle | SpaceshipPolygon
}](figure F) []Position {
	switch f := any(figure).(type) {
	case Rectangle:
		return f[:]

	case Triangle:
		return f[:]

	case SpaceshipPolygon:
		return f[:]

	default:
		return nil

	}
}

func makeTestRect(pos Position, size Size) Rectangle {
	return Rectangle{
		pos,
		pos.Add(Position{X: size.Width}),
		pos.Add(size.ToVector()),
		pos.Add(Position{Y: size.Height}),
	}
}

func makeTestTriangle(pos Position, size Size) Triangle {
	return Triangle{
		pos,
		pos.Add(Position{X: size.Width}),
		pos.Add(size.ToVector()),
	}
}

func testHaveSeparatingAxis[F1, F2 interface {
	Triangle | Rectangle | SpaceshipPolygon
}](t *testing.T, verticesA F1, verticesB F2, want bool) {
	t.Helper()

	if got := HaveSeparatingAxis(figureToSlice(verticesA), figureToSlice(verticesB)); got != want {
		t.Errorf("HaveSeparatingAxis(%v, %v) = %t, want %t", verticesA, verticesB, got, want)
	}
}

func TestHaveSeparatingAxis(t *testing.T) {
	type p = Position
	type s = Size

	testHaveSeparatingAxis(t, makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), false)
	testHaveSeparatingAxis(t, makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestRect(p{X: 1, Y: 1}, s{Width: 1, Height: 1}), false)
	testHaveSeparatingAxis(t, makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestRect(p{X: 1.1, Y: 1.1}, s{Width: 1, Height: 1}), true)
	testHaveSeparatingAxis(t, makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestTriangle(p{X: 1, Y: 1}, s{Width: 1, Height: 1}), false)
	testHaveSeparatingAxis(t, makeTestTriangle(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestTriangle(p{X: 1.1, Y: 1}, s{Width: 1, Height: 1}), true)
}
