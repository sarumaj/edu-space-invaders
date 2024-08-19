package spaceship

import (
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

// Shield represents the shield of the spaceship.
type Shield struct {
	Charge         int // Charge is the charge of the shield.
	Capacity       int // Capacity is the capacity of the shield.
	ChargeDuration time.Duration
	lastChargedAt  time.Time
}

// Health returns the health of the shield.
// It is the charge divided by the capacity.
func (shield Shield) Health() numeric.Number {
	return numeric.Number(shield.Charge) / numeric.Number(shield.Capacity)
}

// Recharge recharges the shield.
func (shield *Shield) Recharge() {
	switch {
	case
		shield.Charge == shield.Capacity,
		time.Since(shield.lastChargedAt) < shield.ChargeDuration:

		return
	}

	shield.Charge += 1
	shield.lastChargedAt = time.Now()
}

// Reduce reduces the shield charge and capacity.
func (shield *Shield) Reduce() {
	if shield.Capacity > 0 {
		shield.Capacity -= 1
	}

	if shield.Charge > shield.Capacity {
		shield.Charge = shield.Capacity
	}
}

// Reinforce reinforces the shield.
// It increases the capacity and charge by 1.
func (shield *Shield) Reinforce() {
	shield.Capacity += 1
	shield.Charge += 1
}

// Use uses the shield.
func (shield *Shield) Use() bool {
	if shield.Charge > 0 {
		shield.Charge -= 1
		return true
	}

	return false
}
