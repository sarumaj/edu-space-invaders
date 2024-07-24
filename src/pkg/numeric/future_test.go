//go:build future

package numeric

import "testing"

func TestFigureArea(t *testing.T) {
	for _, tt := range []struct {
		name string
		args interface{ Area() Number }
		want Number
	}{
		{"Rectangle", GetRectangularVertices(Locate(0, 0), Size{Width: 1, Height: 1}), 1},
		{"Rhombus", Rectangle{{0, 0}, {1, 1}, {2, 0}, {1, -1}}, 2},
		{"Trapezoid", Rectangle{{0, 0}, {3, 0}, {2, 1}, {1, 1}}, 2},
		{"Triangle", Triangle{{0, 0}, {1, 0}, {0, 1}}, 0.5},
		{"SpaceshipPolygon", GetSpaceshipVerticesV2(Locate(0, 0), Size{Width: 1, Height: 1}, true), 0.38},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.Area()
			if !Equal(got, tt.want, 1e-9) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShoelaceFormula(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []Position
		want Number
	}{
		{"Triangle", []Position{{0, 0}, {1, 0}, {0, 1}}, 0.5},
		{"Rectangle", []Position{{0, 0}, {1, 0}, {1, 1}, {0, 1}}, 1},
		{"Rhombus", []Position{{0, 0}, {1, 1}, {2, 0}, {1, -1}}, 2},
		{"Trapezoid", []Position{{0, 0}, {3, 0}, {2, 1}, {1, 1}}, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShoelaceFormula(tt.args); !Equal(got, tt.want, 1e-9) {
				t.Errorf("ShoelaceFormula(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
