package objects

import (
	"fmt"
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

func testObject[P interface{ Number | Size }](t *testing.T, args testObjectArgs[P], want testObjectWant[P]) {
	var x P
	t.Run(fmt.Sprintf("%T", x), func(t *testing.T) {
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
	testObject(t, testObjectArgs[Size]{
		pos:    Position{X: 1, Y: 2},
		size:   Size{Width: 5, Height: 4},
		scales: Position{X: 0.1, Y: 0.3},
	}, testObjectWant[Size]{
		pos:  Position{X: 3.25, Y: 3.4},
		size: Size{Width: 0.5, Height: 1.2},
	})

	// Downscale a round object by 0.1 in the X direction and 0.3 in the Y direction.
	testObject(t, testObjectArgs[Number]{
		pos:    Position{X: 10, Y: 2},
		size:   10,
		scales: Position{X: 0.3, Y: 0.1},
	}, testObjectWant[Number]{
		pos:  Position{X: 14, Y: 6},
		size: 2,
	})

	// Upscale a rectangular object by 2.0 in the X direction and 1.2 in the Y direction.
	testObject(t, testObjectArgs[Size]{
		pos:    Position{X: 1, Y: 2},
		size:   Size{Width: 5, Height: 4},
		scales: Position{X: 2.0, Y: 1.2},
	}, testObjectWant[Size]{
		pos:  Position{X: -1.5, Y: 1.6},
		size: Size{Width: 10, Height: 4.8},
	})

	// Upscale a round object by 2.0 in the X direction and 1.2 in the Y direction.
	testObject(t, testObjectArgs[Number]{
		pos:    Position{X: 10, Y: 2},
		size:   10,
		scales: Position{X: 2.0, Y: 1.2},
	}, testObjectWant[Number]{
		pos:  Position{X: 7, Y: -1},
		size: 16,
	})
}
