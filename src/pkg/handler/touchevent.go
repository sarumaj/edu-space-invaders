package handler

import (
	"fmt"
	"sync"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

// touchEvent represents a touch event.
type touchEvent struct {
	mutex                                       *sync.Mutex
	StartPosition, CurrentPosition, EndPosition numeric.Position
	StartTime, EndTime                          time.Time
	Type                                        touchType
	MultiTap                                    bool
}

// Reset resets the touch event.
func (t *touchEvent) Reset() *touchEvent {
	t.mutex.Lock()
	t.StartPosition = numeric.Position{}
	t.CurrentPosition = numeric.Position{}
	t.EndPosition = numeric.Position{}
	t.StartTime = time.Time{}
	t.EndTime = time.Time{}
	t.MultiTap = false
	t.mutex.Unlock()

	return t
}

// Send sends the touch event to the specified channel.
func (t touchEvent) Send(rcv chan<- touchEvent) {
	rcv <- touchEvent{
		StartPosition:   t.StartPosition,
		CurrentPosition: t.CurrentPosition,
		EndPosition:     t.EndPosition,
		StartTime:       t.StartTime,
		EndTime:         t.EndTime,
		MultiTap:        t.MultiTap,
		Type:            t.Type,
	}
}

// SetCurrentPosition sets the current position of the touch event.
func (t *touchEvent) SetCurrentPosition(pos numeric.Position) *touchEvent {
	t.mutex.Lock()
	t.CurrentPosition = pos
	t.mutex.Unlock()

	return t
}

// SetEndPosition sets the end position of the touch event.
func (t *touchEvent) SetEndPosition(p numeric.Position) *touchEvent {
	t.mutex.Lock()
	t.EndPosition = p
	t.mutex.Unlock()

	return t
}

// SetEndTime sets the end time of the touch event.
func (t *touchEvent) SetEndTime(when time.Time) *touchEvent {
	t.mutex.Lock()
	t.EndTime = when
	t.mutex.Unlock()

	return t
}

// SetMultiTap sets the multi-tap state of the touch event.
func (t *touchEvent) SetMultiTap(m bool) *touchEvent {
	t.mutex.Lock()
	t.MultiTap = m
	t.mutex.Unlock()

	return t
}

// SetStartPosition sets the start position of the touch event.
func (t *touchEvent) SetStartPosition(pos numeric.Position) *touchEvent {
	t.mutex.Lock()
	t.StartPosition = pos
	t.mutex.Unlock()

	return t
}

// SetStartTime sets the start time of the touch event.
func (t *touchEvent) SetStartTime(when time.Time) *touchEvent {
	t.mutex.Lock()
	t.StartTime = when
	t.mutex.Unlock()

	return t
}

// SetType sets the type of the touch event.
func (t *touchEvent) SetType(tt touchType) *touchEvent {
	t.mutex.Lock()
	t.Type = tt
	t.mutex.Unlock()

	return t
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
	return fmt.Sprintf("Touch (Start: %s, End: %s, Duration: %s, MultiTap: %t, Type: %s)",
		t.StartPosition, t.EndPosition, t.TapDuration(), t.MultiTap, t.Type)
}

const (
	TouchTypeUnknown touchType = iota // touchTypeUnknown represents an unknown touch event.
	TouchTypeStart                    // touchTypeStart represents the start of the touch event.
	TouchTypeMove                     // touchTypeMove represents the move of the touch event.
	TouchTypeEnd                      // touchTypeEnd represents the end of the touch event.
)

// touchType represents the type of the touch event.
type touchType int

// String returns the string representation of the touch type.
func (t touchType) String() string {
	if t > TouchTypeEnd {
		return "Unknown"
	}

	return [...]string{"Unknown", "Start", "Move", "End"}[t]
}
