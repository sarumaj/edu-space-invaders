package numeric

import (
	"reflect"
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

func testHasSeparatingAxis[F1, F2 interface {
	Triangle | Rectangle | SpaceshipPolygon
}](t *testing.T, verticesA F1, verticesB F2, sort bool, want bool) {
	t.Helper()

	if sort {
		verticesA = F1(reflect.ValueOf(&verticesA).Elem().MethodByName("Vertices").Call(nil)[0].Interface().(Vertices).Sort(true))
		verticesB = F2(reflect.ValueOf(&verticesB).Elem().MethodByName("Vertices").Call(nil)[0].Interface().(Vertices).Sort(true))
	}

	vertices := reflect.ValueOf(&verticesA).Elem().MethodByName("Vertices").Call(nil)[0].Interface().(Vertices)
	other := reflect.ValueOf(&verticesB).Elem().MethodByName("Vertices").Call(nil)[0].Interface().(Vertices)

	if got := vertices.HasSeparatingAxis(other); got != want {
		t.Errorf("HasSeparatingAxis(%v, %v) = %t, want %t", verticesA, verticesB, got, want)
	}
}

func TestArea(t *testing.T) {
	type args struct {
		vertices Vertices
		sort     bool
	}
	for _, tt := range []struct {
		name string
		args args
		want Number
	}{
		{"Triangle", args{Vertices{{0, 0}, {1, 0}, {0, 1}}, false}, 0.5},
		{"Rectangle", args{Vertices{{0, 0}, {1, 1}, {1, 0}, {0, 1}}, false}, 0},
		{"Rectangle", args{Vertices{{0, 0}, {1, 1}, {1, 0}, {0, 1}}, true}, 1},
		{"Rhombus", args{Vertices{{0, 0}, {2, 0}, {1, 1}, {1, -1}}, false}, 0},
		{"Rhombus", args{Vertices{{0, 0}, {2, 0}, {1, 1}, {1, -1}}, true}, 2},
		{"Trapezoid", args{Vertices{{0, 0}, {2, 1}, {3, 0}, {1, 1}}, false}, 0},
		{"Trapezoid", args{Vertices{{0, 0}, {2, 1}, {3, 0}, {1, 1}}, true}, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.sort {
				tt.args.vertices = tt.args.vertices.Sort(true)
			}
			if got := tt.args.vertices.Area(); !Equal(got, tt.want, 1e-9) {
				t.Errorf("ShoelaceFormula(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestCentroid(t *testing.T) {
	for _, tt := range []struct {
		name string
		args Vertices
		want Position
	}{
		{"Triangle", Vertices{{0, 0}, {1, 0}, {0, 1}}, Position{X: 1.0 / 3, Y: 1.0 / 3}},
		{"Rectangle", Vertices{{0, 0}, {1, 0}, {1, 1}, {0, 1}}, Position{X: 0.5, Y: 0.5}},
		{"Rhombus", Vertices{{0, 0}, {1, 1}, {2, 0}, {1, -1}}, Position{X: 1, Y: 0}},
		{"Trapezoid", Vertices{{0, 0}, {3, 0}, {2, 1}, {1, 1}}, Position{X: 1.5, Y: 0.5}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.Centroid(); !Equal(got, tt.want, 1e-9) {
				t.Errorf("Centroid(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestHasSeparatingAxis(t *testing.T) {
	type p = Position
	type s = Size

	testHasSeparatingAxis(t, makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), false, false)
	testHasSeparatingAxis(t, makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestRect(p{X: 1, Y: 1}, s{Width: 1, Height: 1}), false, false)
	testHasSeparatingAxis(t, makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestRect(p{X: 1.1, Y: 1.1}, s{Width: 1, Height: 1}), false, true)
	testHasSeparatingAxis(t, makeTestRect(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestTriangle(p{X: 1, Y: 1}, s{Width: 1, Height: 1}), false, false)
	testHasSeparatingAxis(t, makeTestTriangle(p{X: 0, Y: 0}, s{Width: 1, Height: 1}), makeTestTriangle(p{X: 1.1, Y: 1}, s{Width: 1, Height: 1}), false, true)
}
