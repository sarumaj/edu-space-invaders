package graphics

import (
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

// Transition represents a color and object size transition.
type SizeTransition struct {
	animationDuration time.Duration         // Animation duration of the transition
	currentScale      numeric.Number        // Current scale used for the beginning of the transition
	targetScale       numeric.Number        // Target scale used for the end of the transition
	size              numeric.Size          // Current size resulting from the transition
	position          numeric.Position      // Current position resulting from the transition
	transitionEnd     func(*SizeTransition) // Transition end callback
	immutable         bool                  // If immutable, the transition cannot be changed, until it ends
}

// Interpolate interpolates the transition.
// It performs one transition step from the current size to the target size and position.
func (t *SizeTransition) Interpolate() {
	numberOfFrames := numeric.Number(config.Config.Control.DesiredFramesPerSecondRate) *
		numeric.Number(t.animationDuration.Seconds())

	if !numeric.Equal(t.currentScale, t.targetScale, 1e-9) {
		sizeFactor := numeric.E.Pow(t.targetScale.Log() / numberOfFrames)
		t.size, t.position = t.size.Resize(sizeFactor, t.position)
		t.size.Scale = 1
		t.currentScale *= sizeFactor
	} else if t.transitionEnd != nil {
		t.transitionEnd(t)
	}
}

// SetAnimationDuration sets the animation duration of the transition.
func (t *SizeTransition) SetAnimationDuration(duration time.Duration) *SizeTransition {
	t.animationDuration = duration
	return t
}

// SetImmutable sets the immutable property of the transition.
func (t *SizeTransition) SetImmutable(immutable bool) *SizeTransition {
	t.immutable = immutable
	return t
}

// SetPosition sets the position of the transition.
func (t *SizeTransition) SetPosition(position numeric.Position) *SizeTransition {
	t.position = position
	return t
}

// SetScale sets the target scale of the transition.
// Current scale is used as the starting point of the transition.
func (t *SizeTransition) SetScale(scale numeric.Number) *SizeTransition {
	if t.immutable {
		return t
	}

	if !numeric.Equal(t.targetScale, scale, 1e-9) {
		t.size, t.position = t.size.Resize(1/t.currentScale, t.position)
		t.size.Scale = 1
		t.targetScale, t.currentScale = scale, 1
	}

	return t
}

// SetPosition sets the position of the transition.
func (t *SizeTransition) SetSize(size numeric.Size) *SizeTransition {
	t.size = size
	return t
}

// SetTransitionEnd sets the transition end function of the transition.
func (t *SizeTransition) SetTransitionEnd(transitionEnd func(*SizeTransition)) *SizeTransition {
	t.transitionEnd = transitionEnd
	return t
}

// Position returns the current position of the transition.
func (t *SizeTransition) Position() numeric.Position { return t.position }

// Size returns the current size of the transition.
func (t *SizeTransition) Size() numeric.Size { return t.size }

// InitialSizeTransition initializes a size transition with the size, and position.
func InitialSizeTransition(size numeric.Size, position numeric.Position) *SizeTransition {
	t := SizeTransition{
		animationDuration: config.Config.Control.AnimationDuration,
		size:              size,
		currentScale:      1,
		targetScale:       1,
		position:          position,
	}

	t.size.Scale = 1

	return &t
}
