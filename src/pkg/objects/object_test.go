package objects

import (
	"testing"
)

type testObjectArgs[P interface{ Number | Size }] struct {
	pos    Position
	size   P
	scales Position
}

type testObjectWant[P interface{ Number | Size }] struct {
	pos  Position
	size P
}

func testObject[P interface{ Number | Size }](t *testing.T, name string, args testObjectArgs[P], want testObjectWant[P]) {
	t.Run(name, func(t *testing.T) {
		object := Measure(args.pos, args.size)
		object.Scale(args.scales)

		var gotPos Position
		object.ApplyPosition(&gotPos)
		if !Equal(gotPos, want.pos, 1e-9) {
			t.Errorf("Position: got %v, want %v", gotPos, want.pos)
		}

		var gotSize P
		object.ApplySize(&gotSize)
		if !Equal(gotSize, want.size, 1e-9) {
			t.Errorf("Size: got %v, want %v", gotSize, want.size)
		}
	})
}

func TestObject(t *testing.T) {
	// Downscale a rectangular object by 0.1 in the X direction and 0.3 in the Y direction.
	testObject(t, "downscaleOnSize", testObjectArgs[Size]{
		pos:    Position{X: 1, Y: 2},
		size:   Size{Width: 5, Height: 4},
		scales: Position{X: 0.1, Y: 0.3},
	}, testObjectWant[Size]{
		pos:  Position{X: 0.1, Y: 0.6},
		size: Size{Width: 0.5, Height: 1.2},
	})

	// Downscale a round object by 0.1 in the X direction and 0.3 in the Y direction.
	testObject(t, "downScaleOnNumber", testObjectArgs[Number]{
		pos:    Position{X: 10, Y: 2},
		size:   10,
		scales: Position{X: 0.3, Y: 0.1},
	}, testObjectWant[Number]{
		pos:  Position{X: 3, Y: 0.2},
		size: 2,
	})

	// Upscale a rectangular object by 2.0 in the X direction and 1.2 in the Y direction.
	testObject(t, "upscaleOnSize", testObjectArgs[Size]{
		pos:    Position{X: 1, Y: 2},
		size:   Size{Width: 5, Height: 4},
		scales: Position{X: 2.0, Y: 1.2},
	}, testObjectWant[Size]{
		pos:  Position{X: 2, Y: 2.4},
		size: Size{Width: 10, Height: 4.8},
	})

	// Upscale a round object by 2.0 in the X direction and 1.2 in the Y direction.
	testObject(t, "upscaleOnNumber", testObjectArgs[Number]{
		pos:    Position{X: 10, Y: 2},
		size:   10,
		scales: Position{X: 2.0, Y: 1.2},
	}, testObjectWant[Number]{
		pos:  Position{X: 20, Y: 2.4},
		size: 16,
	})
}
