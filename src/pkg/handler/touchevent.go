package handler

import (
	"fmt"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// touchEvent represents a touch event.
type touchEvent struct {
	StartPosition, CurrentPosition, EndPosition objects.Position
	StartTime, EndTime                          time.Time
	Correlations                                []touchEvent
}

// Clear clears the touch event.
func (t *touchEvent) Clear() {
	t.StartPosition = objects.Position{}
	t.CurrentPosition = objects.Position{}
	t.EndPosition = objects.Position{}
	t.StartTime = time.Time{}
	t.EndTime = time.Time{}
	t.Correlations = nil
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
	return fmt.Sprintf("Touch (Start: %s, End: %s, Duration: %s)", t.StartPosition, t.EndPosition, t.TapDuration())
}
