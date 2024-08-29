package numeric

import (
	"testing"
)

func TestSizeResize(t *testing.T) {
	size := Size{Width: 100, Height: 200, Scale: 0}
	scale := Number(2)
	position := Position{X: 50, Y: 100}

	var resultSize Size
	var resultPosition Position
	for i, s, p := 0, size, position; i < 10; i, s, p = func(i int, s Size, p Position) (int, Size, Position) {
		sN, pN := s.Resize(scale, p)
		return i + 1, sN, pN
	}(i, s, p) {
		resultSize, resultPosition = s, p
		expectedSize := Size{Scale: scale.Pow(Number(i))}
		expectedSize.Width = size.Width * expectedSize.Scale
		expectedSize.Height = size.Height * expectedSize.Scale
		if i == 0 {
			expectedSize.Scale = size.Scale
		}

		expectedPosition := Position{
			X: position.X - (expectedSize.Width-size.Width)/2,
			Y: position.Y - (expectedSize.Height-size.Height)/2,
		}

		if s != expectedSize {
			t.Errorf("[%d] Expected size to be %v, got %v", i, expectedSize, s)
		}

		if p != expectedPosition {
			t.Errorf("[%d] Expected position to be %v, got %v", i, expectedPosition, p)
		}
	}

	// Gradual restoration logic
	restoredSize, restoredPosition := resultSize.Resize(1/resultSize.Scale, resultPosition)
	restoredSize.Scale = size.Scale
	if restoredSize != size {
		t.Errorf("Expected size to be %v, got %v", size, restoredSize)
	}

	if restoredPosition != position {
		t.Errorf("Expected position to be %v, got %v", position, restoredPosition)
	}
}
