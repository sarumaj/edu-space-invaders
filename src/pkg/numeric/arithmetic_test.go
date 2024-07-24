package numeric

import (
	"testing"
)

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

func testHaveSeparatingAxis[F1, F2 interface{ Slice() []Position }](t *testing.T, verticesA F1, verticesB F2, sort bool, want bool) {
	t.Helper()

	if got := HaveSeparatingAxis(verticesA.Slice(), verticesB.Slice(), sort); got != want {
		t.Errorf("HaveSeparatingAxis(%v, %v) = %t, want %t", verticesA, verticesB, got, want)
	}
}

func TestCentroid(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []Position
		want Position
	}{
		{"Triangle", []Position{{0, 0}, {1, 0}, {0, 1}}, Position{X: 1.0 / 3, Y: 1.0 / 3}},
		{"Rectangle", []Position{{0, 0}, {1, 0}, {1, 1}, {0, 1}}, Position{X: 0.5, Y: 0.5}},
		{"Rhombus", []Position{{0, 0}, {1, 1}, {2, 0}, {1, -1}}, Position{X: 1, Y: 0}},
		{"Trapezoid", []Position{{0, 0}, {3, 0}, {2, 1}, {1, 1}}, Position{X: 1.5, Y: 0.5}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := Centroid(tt.args); !Equal(got, tt.want, 1e-9) {
				t.Errorf("Centroid(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestHaveSeparatingAxis(t *testing.T) {
	type p = Position
	type s = Size

	testHaveSeparatingAxis(t, makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), false, false)
	testHaveSeparatingAxis(t, makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestRect(p{X: 1, Y: 1}, s{Width: 1, Height: 1}), false, false)
	testHaveSeparatingAxis(t, makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestRect(p{X: 1.1, Y: 1.1}, s{Width: 1, Height: 1}), false, true)
	testHaveSeparatingAxis(t, makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestTriangle(p{X: 1, Y: 1}, s{Width: 1, Height: 1}), false, false)
	testHaveSeparatingAxis(t, makeTestTriangle(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestTriangle(p{X: 1.1, Y: 1}, s{Width: 1, Height: 1}), false, true)
}
