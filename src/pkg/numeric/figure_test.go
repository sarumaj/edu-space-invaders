package numeric

import (
	"testing"
)

func TestGetRectangularVertices(t *testing.T) {
	type args struct {
		pos  Position
		size Size
	}

	for _, tt := range []struct {
		name string
		args args
	}{
		{"test#1", args{Position{0, 0}, Size{Width: 1, Height: 1}}},
		{"test#2", args{Position{-1, 3}, Size{Width: 1, Height: 2}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRectangularVertices(tt.args.pos, tt.args.size)
			if !AreVerticesSortedClockwise(got.Slice()) {
				sorted := Rectangle(SortVerticesClockwise(got.Slice()))
				t.Errorf("got %v, want %v", got, sorted)
			}
		})
	}
}

func TestGetSpaceshipVerticesV1(t *testing.T) {
	type args struct {
		pos    Position
		size   Size
		faceUp bool
	}

	for _, tt := range []struct {
		name string
		args args
	}{
		{"test#1", args{Position{0, 0}, Size{Width: 1, Height: 1}, true}},
		{"test#2", args{Position{-1, 3}, Size{Width: 1, Height: 2}, false}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSpaceshipVerticesV1(tt.args.pos, tt.args.size, tt.args.faceUp)
			if !AreVerticesSortedClockwise(got.Slice()) {
				sorted := Triangle(SortVerticesClockwise(got.Slice()))
				t.Errorf("got %v, want %v", got, sorted)
			}
		})
	}
}

func TestGetSpaceshipVerticesV2(t *testing.T) {
	type args struct {
		pos    Position
		size   Size
		faceUp bool
	}

	for _, tt := range []struct {
		name string
		args args
	}{
		{"test#1", args{Position{0, 0}, Size{Width: 1, Height: 1}, true}},
		{"test#2", args{Position{-1, 3}, Size{Width: 1, Height: 2}, false}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSpaceshipVerticesV2(tt.args.pos, tt.args.size, tt.args.faceUp)
			if !AreVerticesSortedClockwise(got.Slice()) {
				sorted := SpaceshipPolygon(SortVerticesClockwise(got.Slice()))
				t.Errorf("got %v, want %v", got, sorted)
			}
		})
	}
}
