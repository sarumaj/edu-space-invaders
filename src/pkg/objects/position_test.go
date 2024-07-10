package objects

import "testing"

func TestPosition(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		pos := Position{X: 1, Y: 2}
		other := Position{X: 3, Y: 4}
		want := Position{X: 4, Y: 6}

		got := pos.Add(other)
		if got != want {
			t.Errorf("Position.Add() = %v, want %v", got, want)
		}
	})

	t.Run("AddN", func(t *testing.T) {
		pos := Position{X: 1, Y: 2}
		n := Number(3)
		want := Position{X: 4, Y: 5}

		got := pos.AddN(n)
		if got != want {
			t.Errorf("Position.AddN() = %v, want %v", got, want)
		}
	})

	t.Run("Distance", func(t *testing.T) {
		pos := Position{X: 1, Y: 2}
		other := Position{X: 4, Y: 6}
		want := Number(5)

		got := pos.Distance(other)
		if got != want {
			t.Errorf("Position.Distance() = %v, want %v", got, want)
		}
	})

	t.Run("Div", func(t *testing.T) {
		pos := Position{X: 4, Y: 6}
		other := Number(2)
		want := Position{X: 2, Y: 3}

		got := pos.Div(other)
		if got != want {
			t.Errorf("Position.Div() = %v, want %v", got, want)
		}
	})

	t.Run("Equal", func(t *testing.T) {
		pos := Position{X: 1, Y: 2}
		other := Position{X: 1, Y: 2}

		if !pos.Equal(other) {
			t.Errorf("Position.Equal() failed, got: false, want: true")
		}
	})

	t.Run("Greater", func(t *testing.T) {
		pos := Position{X: 4, Y: 6}
		other := Position{X: 1, Y: 2}

		if !pos.Greater(other) {
			t.Errorf("Position.Greater() failed, got: false, want: true")
		}
	})

	t.Run("GreaterOrEqual", func(t *testing.T) {
		pos := Position{X: 4, Y: 6}
		other := Position{X: 4, Y: 6}

		if !pos.GreaterOrEqual(other) {
			t.Errorf("Position.GreaterOrEqual() failed, got: false, want: true")
		}
	})

	t.Run("IsZero", func(t *testing.T) {
		pos := Position{X: 0, Y: 0}

		if !pos.IsZero() {
			t.Errorf("Position.IsZero() failed, got: false, want: true")
		}
	})

	t.Run("Less", func(t *testing.T) {
		pos := Position{X: 1, Y: 2}
		other := Position{X: 4, Y: 6}

		if !pos.Less(other) {
			t.Errorf("Position.Less() failed, got: false, want: true")
		}
	})

	t.Run("LessOrEqual", func(t *testing.T) {
		pos := Position{X: 1, Y: 2}
		other := Position{X: 4, Y: 6}

		if !pos.LessOrEqual(other) {
			t.Errorf("Position.LessOrEqual() failed, got: false, want: true")
		}
	})

	t.Run("Magnitude", func(t *testing.T) {
		pos := Position{X: 3, Y: 4}
		want := Number(5)

		got := pos.Magnitude()
		if got != want {
			t.Errorf("Position.Magnitude() = %v, want %v", got, want)
		}
	})

	t.Run("Mul", func(t *testing.T) {
		pos := Position{X: 1, Y: 2}
		other := Number(3)
		want := Position{X: 3, Y: 6}

		got := pos.Mul(other)
		if got != want {
			t.Errorf("Position.Mul() = %v, want %v", got, want)
		}
	})

	t.Run("Normalize", func(t *testing.T) {
		pos := Position{X: 3, Y: 4}
		want := Position{X: 0.6, Y: 0.8}

		got := pos.Normalize()
		if got != want {
			t.Errorf("Position.Normalize() = %v, want %v", got, want)
		}
	})

	t.Run("Root", func(t *testing.T) {
		pos := Position{X: 4, Y: 4}
		want := Number(4)

		got := pos.Root()
		if got != want {
			t.Errorf("Position.Root() = %v, want %v", got, want)
		}
	})

	t.Run("Sub", func(t *testing.T) {
		pos := Position{X: 4, Y: 6}
		other := Position{X: 1, Y: 2}
		want := Position{X: 3, Y: 4}

		got := pos.Sub(other)
		if got != want {
			t.Errorf("Position.Sub() = %v, want %v", got, want)
		}
	})

	t.Run("SubN", func(t *testing.T) {
		pos := Position{X: 4, Y: 6}
		n := Number(2)
		want := Position{X: 2, Y: 4}

		got := pos.SubN(n)
		if got != want {
			t.Errorf("Position.SubN() = %v, want %v", got, want)
		}
	})

}
