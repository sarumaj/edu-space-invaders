package spaceship

const (
	Neutral SpaceshipState = iota // Neutral is the default state
	Damaged                       // Damaged is the state when the spaceship is hit
	Boosted                       // Boosted is the state when the spaceship is upgraded
	Frozen                        // Frozen is the state when the spaceship is frozen
)

// SpaceshipState represents the state of the spaceship (Neutral, Damaged, Boosted, Frozen)
type SpaceshipState int

// String returns the string representation of the spaceship state.
func (state SpaceshipState) String() string {
	return [...]string{"Neutral", "Damaged", "Boosted", "Frozen"}[state]
}
