package numeric

import (
	"math/rand/v2"
)

// Equal checks if the two objects are equal.
func Equal[P interface{ Number | Position | Size }](a, b P, tolerance Number) bool {
	switch a := any(a).(type) {
	case Number:
		return (a - any(b).(Number)).Abs() <= tolerance

	case Position:
		b := any(b).(Position)
		return (a.X-b.X).Abs() <= tolerance && (a.Y-b.Y).Abs() <= tolerance

	case Size:
		b := any(b).(Size)
		return Equal(a.ToVector(), b.ToVector(), tolerance)

	}

	return false
}

// RandomRange returns a random number between min and max.
func RandomRange[Numeric1, Numeric2 interface{ ~float64 | ~int }](min Numeric1, max Numeric2) Number {
	return Number(min) + Number(rand.Float64())*(Number(max)-Number(min))
}

// SampleUniform returns true with the given probability.
func SampleUniform[Numeric interface{ ~float64 | ~int }](probability Numeric) bool {
	if probability <= 0 {
		return false
	}

	if probability >= 1 {
		return true
	}

	return rand.Float64() < float64(probability)
}
