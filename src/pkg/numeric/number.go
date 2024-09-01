package numeric

import (
	"fmt"
	"math"
)

const (
	E  = Number(math.E)
	Pi = Number(math.Pi)
)

// Number represents a number.
type Number float64

// Abs returns the absolute value of the number.
func (n Number) Abs() Number {
	if n < 0 {
		return -n
	}

	return n
}

// Clamp clamps the number between the minimum and maximum numbers.
func (n Number) Clamp(min, max Number) Number { return n.Max(min.Min(max)).Min(max.Max(min)) }

// Int returns the int representation of the number.
func (n Number) Int() int { return int(math.Round(float64(n))) }

func (n Number) InRange(min, max Number) bool { return n >= min.Min(max) && n <= max.Max(min) }

// Float returns the float64 representation of the number.
func (n Number) Float() float64 { return float64(n) }

// Log returns the natural logarithm of the number.
func (n Number) Log() Number { return Number(math.Log(float64(n))) }

// Max returns the maximum number from the list of numbers (including oneself).
func (n Number) Max(others ...Number) Number { return n.MaxMin(true, others...) }

// MaxMin returns the maximum or minimum number from the list of numbers (including oneself).
func (n Number) MaxMin(max bool, others ...Number) Number {
	result := n
	for _, other := range others {
		switch {
		case
			max && other > result,
			!max && other < result:
			result = other
		}
	}

	return result
}

// Min returns the minimum number from the list of numbers (including oneself).
func (n Number) Min(others ...Number) Number { return n.MaxMin(false, others...) }

// Polarity returns the polarity (sign) of the number.
func (n Number) Polarity() Number { return n / n.Abs() }

// Root returns the square root of the number.
func (n Number) Root() Number { return Number(math.Sqrt(float64(n))) }

// Pow returns the number raised to the power of the other number.
func (n Number) Pow(other Number) Number { return Number(math.Pow(float64(n), float64(other))) }

// String returns the string representation of the number.
func (n Number) String() string { return fmt.Sprintf("%g", n) }

// Inf returns positive infinity if sign is positive, negative infinity if sign is negative.
func Inf(sign int) Number { return Number(math.Inf(sign)) }
