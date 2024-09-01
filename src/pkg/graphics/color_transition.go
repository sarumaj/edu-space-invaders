package graphics

import (
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

// ColorTransition represents a color transition.
type ColorTransition struct {
	animationDuration time.Duration          // Animation duration of the transition
	currentColor      Color                  // Current color used for the beginning of the transition
	targetColor       Color                  // Target color used for the end of the transition
	currentGradient   Color                  // Intermediate color resulting from the transition
	transitionEnd     func(*ColorTransition) // Transition end callback
	immutable         bool                   // If immutable, the transition cannot be changed, until it ends
}

// Gradient returns the current gradient color of the transition.
func (t *ColorTransition) Gradient() Color {
	return t.currentGradient
}

// Interpolate interpolates the transition.
// It performs one transition step from the current color to the target color.
func (t *ColorTransition) Interpolate() {
	numberOfFrames := numeric.Number(config.Config.Control.DesiredFramesPerSecondRate) *
		numeric.Number(t.animationDuration.Seconds())

	if !t.currentGradient.Equal(t.targetColor) {
		for i := 0; i < 4; i++ {
			t.currentGradient[i] += (t.targetColor[i] - t.currentColor[i]) / numberOfFrames
		}
	} else if t.transitionEnd != nil {
		t.transitionEnd(t)
	}
}

// SetAnimationDuration sets the animation duration of the transition.
func (t *ColorTransition) SetAnimationDuration(duration time.Duration) *ColorTransition {
	t.animationDuration = duration
	return t
}

// SetColor sets the target color of the transition.
// Current gradient is used as the starting point of the transition.
func (t *ColorTransition) SetColor(to Color) *ColorTransition {
	if t.immutable {
		return t
	}

	if !t.targetColor.Equal(to) {
		t.currentColor, t.targetColor = t.currentGradient, to
	}

	return t
}

// SetGradient allows to set the current gradient color of the transition.
func (t *ColorTransition) SetGradient(color Color) *ColorTransition {
	t.currentGradient = color
	return t
}

// SetImmutable sets the immutable property of the transition.
func (t *ColorTransition) SetImmutable(immutable bool) *ColorTransition {
	t.immutable = immutable
	return t
}

// SetTransitionEnd sets the transition end callback.
func (t *ColorTransition) SetTransitionEnd(transitionEnd func(*ColorTransition)) *ColorTransition {
	t.transitionEnd = transitionEnd
	return t
}

// InitialColorTransition initializes a color transition with the given color.
func InitialColorTransition(color Color) *ColorTransition {
	return &ColorTransition{
		animationDuration: config.Config.Control.AnimationDuration,
		currentColor:      color,
		targetColor:       color,
		currentGradient:   color,
	}
}
