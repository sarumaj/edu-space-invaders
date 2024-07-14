package handler

const (
	ArrowDown  keyBinding = "ArrowDown"  // ArrowDown represents the down arrow key.
	ArrowLeft  keyBinding = "ArrowLeft"  // ArrowLeft represents the left arrow key.
	ArrowRight keyBinding = "ArrowRight" // ArrowRight represents the right arrow key.
	ArrowUp    keyBinding = "ArrowUp"    // ArrowUp represents the up arrow key.
	Pause      keyBinding = "Pause"      // Pause represents the pause key.
	Space      keyBinding = "Space"      // Space represents the space key.
)

// keyBinding represents a key binding.
type keyBinding string

// keyEvent represents a key event.
type keyEvent struct {
	Key     keyBinding
	Pressed bool
}
