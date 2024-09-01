package numeric

import (
	"math/rand/v2"
	"slices"
)

// Equal checks if the two objects are equal.
func Equal[P interface {
	Number | Position | Size
}](a, b P, tolerance Number) bool {
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

// Randomize returns a random number within the given probability.
func Randomize[Numeric interface{ ~float64 | ~int }](n Numeric, cutoff Number) Numeric {
	return n + Numeric(RandomRange(-Number(n)*cutoff, Number(n)*cutoff))
}

// RandomRange returns a random number between min and max.
func RandomRange[Numeric1, Numeric2 interface{ ~float64 | ~int }](min Numeric1, max Numeric2) Number {
	return Number(min) + Number(rand.Float64())*(Number(max)-Number(min))
}

// RandomSort sorts the slice randomly.
func RandomSort[Numeric interface{ ~float64 | ~int }](slice []Numeric) []Numeric {
	slices.SortStableFunc(slice, func(a, b Numeric) int {
		return rand.IntN(3) - 1
	})

	return slice
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
