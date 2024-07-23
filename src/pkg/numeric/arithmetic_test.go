package numeric

import "testing"

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

func testHaveSeparatingAxis(t *testing.T, verticesA Figure, verticesB Figure, want bool) {
	t.Helper()
	if got := HaveSeparatingAxis(verticesA.Slice(), verticesB.Slice()); got != want {
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
