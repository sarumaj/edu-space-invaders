package handler

import (
	"fmt"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// touchEvent represents a touch event.
type touchEvent struct {
	StartPosition, EndPosition objects.Position
	StartTime, EndTime         time.Time
}

// Delta returns the delta of the touch event.
func (t touchEvent) Delta() objects.Position {
	return t.EndPosition.Sub(t.StartPosition)
}

// TapDuration returns the duration of the tap.
func (t touchEvent) TapDuration() time.Duration {
	duration := t.EndTime.Sub(t.StartTime)
	if duration < 0 {
		duration = 0
	}

	return duration
}

// String returns the string representation of the touch event.
func (t touchEvent) String() string {
	return fmt.Sprintf("Touch (Start: %s, Delta: %s, Duration: %s)", t.StartPosition, t.Delta(), t.TapDuration())
}
