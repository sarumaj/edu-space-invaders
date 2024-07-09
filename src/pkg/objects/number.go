package objects

import (
	"fmt"
	"math"
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

// Int returns the int representation of the number.
func (n Number) Int() int {
	return int(math.Round(float64(n)))
}

// Float returns the float64 representation of the number.
func (n Number) Float() float64 {
	return float64(n)
}

// Root returns the square root of the number.
func (n Number) Root() Number {
	return Number(math.Sqrt(float64(n)))
}

// String returns the string representation of the number.
func (n Number) String() string {
	return fmt.Sprintf("%g", n)
}
