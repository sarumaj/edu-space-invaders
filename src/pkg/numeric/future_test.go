//go:build future

package numeric

import "testing"

func TestFigureArea(t *testing.T) {
	type args struct {
		figure interface{ Area(bool) Number }
		sort   bool
	}
	for _, tt := range []struct {
		name string
		args args
		want Number
	}{
		{"Rectangle", args{GetRectangularVertices(Locate(0, 0), Size{Width: 1, Height: 1}), false}, 1},
		{"Rectangle", args{GetRectangularVertices(Locate(0, 0), Size{Width: 1, Height: 1}), true}, 1},
		{"Rhombus", args{Rectangle{{0, 0}, {1, 1}, {1, -1}, {2, 0}}, false}, 0},
		{"Rhombus", args{Rectangle{{0, 0}, {1, 1}, {1, -1}, {2, 0}}, true}, 2},
		{"Trapezoid", args{Rectangle{{0, 0}, {2, 1}, {3, 0}, {1, 1}}, false}, 0},
		{"Trapezoid", args{Rectangle{{0, 0}, {2, 1}, {3, 0}, {1, 1}}, true}, 2},
		{"Triangle", args{Triangle{{0, 0}, {1, 0}, {0, 1}}, false}, 0.5},
		{"SpaceshipPolygon", args{GetSpaceshipVerticesV2(Locate(0, 0), Size{Width: 1, Height: 1}, true), false}, 0.38},
		{"SpaceshipPolygon", args{GetSpaceshipVerticesV2(Locate(0, 0), Size{Width: 1, Height: 1}, true), true}, 0.38},
		{"Unordered", args{SpaceshipPolygon{{3, 0}, {4, 1}, {1, 1}, {2, 1}, {-1, -2}, {8, 9}, {0.5, 12}}, false}, 32.25},
		{"Unordered", args{SpaceshipPolygon{{3, 0}, {4, 1}, {1, 1}, {2, 1}, {-1, -2}, {8, 9}, {0.5, 12}}, true}, 55},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.figure.Area(tt.args.sort)
			if !Equal(got, tt.want, 1e-9) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShoelaceFormula(t *testing.T) {
	type args struct {
		vertices []Position
		sort     bool
	}
	for _, tt := range []struct {
		name string
		args args
		want Number
	}{
		{"Triangle", args{[]Position{{0, 0}, {1, 0}, {0, 1}}, false}, 0.5},
		{"Rectangle", args{[]Position{{0, 0}, {1, 1}, {1, 0}, {0, 1}}, false}, 0},
		{"Rectangle", args{[]Position{{0, 0}, {1, 1}, {1, 0}, {0, 1}}, true}, 1},
		{"Rhombus", args{[]Position{{0, 0}, {2, 0}, {1, 1}, {1, -1}}, false}, 0},
		{"Rhombus", args{[]Position{{0, 0}, {2, 0}, {1, 1}, {1, -1}}, true}, 2},
		{"Trapezoid", args{[]Position{{0, 0}, {2, 1}, {3, 0}, {1, 1}}, false}, 0},
		{"Trapezoid", args{[]Position{{0, 0}, {2, 1}, {3, 0}, {1, 1}}, true}, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShoelaceFormula(tt.args.vertices, tt.args.sort); !Equal(got, tt.want, 1e-9) {
				t.Errorf("ShoelaceFormula(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
