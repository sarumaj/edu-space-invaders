package spaceship

import (
	"fmt"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/bullet"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
)

// Spaceship represents the player's spaceship.
type Spaceship struct {
	Position            numeric.Position // Position of the spaceship
	Speed               numeric.Position // Speed of the spaceship in both directions
	Directions          Directions       // Directions the spaceship can move
	Size                numeric.Size     // Size of the
	CurrentScale        numeric.Position // Scale of the spaceship
	Bullets             bullet.Bullets   // Bullets fired by the spaceship
	Cooldown            time.Duration    // Time between shots
	Level               *SpaceshipLevel  // Spaceship level
	State               SpaceshipState   // Spaceship state
	HighScore           int              // HighScore is the high score of the spaceship.
	lastFired           time.Time        // Last time the spaceship fired
	lastStateTransition time.Time        // Last time the spaceship changed state
	lastThrottledLog    time.Time        // Last time the spaceship throttled log messages
}

// isFrozen checks if the spaceship can move or shoot.
// If the spaceship is in the Frozen state, a message is sent
// to the message box indicating that the spaceship is still frozen.
// The message is throttled based on the spaceship's log throttling duration.
func (spaceship *Spaceship) isFrozen() bool {
	if spaceship.State == Frozen {
		now := time.Now()
		if now.Sub(spaceship.lastThrottledLog) >= config.Config.Spaceship.LogThrottling {
			config.SendMessage(config.Execute(config.Config.MessageBox.Messages.Templates.SpaceshipStillFrozen, config.Template{
				"FreezeDuration": time.Until(spaceship.lastStateTransition.Add(config.Config.Spaceship.FreezeDuration)).
					Round(config.Config.Spaceship.LogThrottling),
			}))
			spaceship.lastThrottledLog = now
		}

		return true
	}

	return false
}

// ChangeState changes the state of the spaceship.
// If the state is Boosted, the spaceship's cannons are doubled
// and its size is doubled. If the number of cannons exceeds
// the maximum number of cannons, it is set to the maximum number.
func (spaceship *Spaceship) ChangeState(state SpaceshipState) {
	if state == Boosted {
		spaceship.Level.Cannons *= 2
		if spaceship.Level.Cannons > config.Config.Spaceship.MaximumCannons {
			spaceship.Level.Cannons = config.Config.Spaceship.MaximumCannons
		}
		spaceship.Size.Width = numeric.Number(config.Config.Spaceship.Width * 2)
		spaceship.Position.X -= numeric.Number(config.Config.Spaceship.Width / 2)
	}

	spaceship.State = state
	spaceship.lastStateTransition = time.Now()

	switch spaceship.State {
	case Boosted:
		go config.PlayAudio("spaceship_boost.wav", false)

	case Frozen:
		go config.PlayAudio("spaceship_freeze.wav", false)

	case Damaged:
		go config.PlayAudio("spaceship_crash.wav", false)

	}
}

// DetectCollisionV1 checks if the spaceship has collided with an enemy.
// The collision detection is based on the bounding box method.
// It is less accurate than DetectCollisionV2.
func (spaceship Spaceship) DetectCollisionV1(e enemy.Enemy) bool {
	return spaceship.Position.Less(e.Position.Add(e.Size.ToVector())) &&
		spaceship.Position.Add(spaceship.Size.ToVector()).Greater(e.Position)
}

// DetectCollision V2 checks if the spaceship has collided with an enemy.
// The collision detection is based on the separating axis theorem.
// The separating axis theorem states that if two convex shapes do not overlap
// on all axes, then they do not overlap.
// This version is more accurate than DetectCollisionV1.
// It uses the triangular vertices of the spaceship and the enemy.
func (spaceship Spaceship) DetectCollisionV2(e enemy.Enemy) bool {
	// Get the vertices of the triangles
	spaceshipVertices := numeric.GetSpaceshipVerticesV1(spaceship.Position, spaceship.Size, true)
	enemyVertices := numeric.GetSpaceshipVerticesV1(e.Position, e.Size, false)

	// Check for overlap on all axes
	return !numeric.HaveSeparatingAxis(spaceshipVertices[:], enemyVertices[:])
}

// DetectCollisionV3 checks if the spaceship has collided with an enemy.
// The collision detection is based on the separating axis theorem.
// The separating axis theorem states that if two convex shapes do not overlap
// on all axes, then they do not overlap.
// This version is more accurate than DetectCollisionV2.
// It uses the exact vertices of the spaceship and the enemy.
func (spaceship Spaceship) DetectCollisionV3(e enemy.Enemy) bool {
	// Get the vertices of the spaceship polygons
	spaceshipVertices := numeric.GetSpaceshipVerticesV2(spaceship.Position, spaceship.Size, true)
	enemyVertices := numeric.GetSpaceshipVerticesV2(e.Position, e.Size, false)

	// Check for overlap on all axes
	return !numeric.HaveSeparatingAxis(spaceshipVertices[:], enemyVertices[:])
}

// Draw draws the spaceship on the canvas.
// The spaceship is drawn in white color if it is in the Neutral state.
// The spaceship is drawn in dark red color if it is in the Damaged state.
// The spaceship is drawn in yellow color if it is in the Boosted state.
// The spaceship is drawn in blue color if it is in the Frozen state.
func (spaceship Spaceship) Draw() {
	config.DrawSpaceship(
		spaceship.Position.Pack(),
		spaceship.Size.Pack(),
		true,
		map[SpaceshipState]string{
			Neutral: "Lavender",
			Damaged: "Crimson",
			Boosted: "Chartreuse",
			Frozen:  "DeepSkyBlue",
		}[spaceship.State],
	)
}

// Fire fires bullets from the spaceship.
// The number of bullets fired is equal to the number of cannons
// the spaceship has. The damage of the bullets is calculated
// based on the spaceship's level.
// The trajectory of the bullets is skewed based on the position
// of the cannon.
func (spaceship *Spaceship) Fire() {
	if spaceship.isFrozen() {
		return
	}

	if time.Since(spaceship.lastFired) < spaceship.Cooldown {
		return
	}

	for i := 1; i < spaceship.Level.Cannons+1; i++ {
		spaceship.Bullets.Reload(
			spaceship.Position.Add(numeric.Locate(
				// X position of the bullet
				// The X position of the bullet is calculated based on the position of the cannon.
				// The X position of the bullet is the X position of the spaceship plus the width of the spaceship
				// times the position of the cannon minus the width of the bullet divided by 2.
				spaceship.Size.Width*numeric.Number(i)/numeric.Number(spaceship.Level.Cannons+1)-numeric.Number(config.Config.Bullet.Width/2),
				0,
			)),
			spaceship.GetBulletDamage(),
			// Skew of the bullet
			// Skew is the skew of the bullet based on the position of the cannon.
			float64(i)/float64(spaceship.Level.Cannons+1),
			// Speed boost of the bullet
			spaceship.Level.AccelerateRate.Float(),
		)
	}

	spaceship.lastFired = time.Now()

	go config.PlayAudio("spaceship_cannon_fire.wav", false)
}

// GetBulletDamage returns the damage of the bullets fired by the spaceship.
func (spaceship Spaceship) GetBulletDamage() int {
	// Calculate the base damage
	base := config.Config.Bullet.InitialDamage + spaceship.Level.Progress
	// Calculate the modifier
	modifier := (spaceship.Level.Progress/config.Config.Bullet.ModifierProgressStep + 1) * spaceship.Level.Cannons
	// Return the damage
	return base*modifier + numeric.RandomRange(0, base*modifier).Int()
}

// IsDestroyed checks if the spaceship is destroyed.
func (spaceship Spaceship) IsDestroyed() bool {
	return spaceship.Level.Progress == 0
}

// MoveDown moves the spaceship down.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is greater than the canvas height,
// it is set to the canvas height.
func (spaceship *Spaceship) MoveDown() {
	if spaceship.isFrozen() {
		return
	}

	// Brake the spaceship if it is moving in the opposite direction
	if spaceship.Directions.IsHeadedTo(Up) {
		spaceship.Speed.Y = 0
	}

	// Set the vertical direction
	spaceship.Directions.SetVertical(Down)

	// Accelerate the spaceship vertically
	spaceship.Speed.Y += numeric.Number(spaceship.Level.AccelerateRate)

	// Limit the speed of the spaceship
	if spaceship.Speed.Magnitude().Float() > config.Config.Spaceship.MaximumSpeed {
		spaceship.Speed = spaceship.Speed.Normalize().Mul(numeric.Number(config.Config.Spaceship.MaximumSpeed))
	}

	// Check the vertical boundaries and update the spaceship position
	spaceship.Position.Y += spaceship.Speed.Y
	canvasDimensions := config.CanvasBoundingBox()
	if spaceship.Position.Y.Float() > canvasDimensions.OriginalHeight {
		spaceship.Position.Y = numeric.Number(canvasDimensions.OriginalHeight)
		spaceship.Speed.Y = 0
	}

	go config.PlayAudio("spaceship_deceleration.wav", false)
}

// MoveLeft moves the spaceship to the left.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is less than 0, it is set to 0.
func (spaceship *Spaceship) MoveLeft() {
	if spaceship.isFrozen() {
		return
	}

	// Brake the spaceship if it is moving in the opposite direction
	if spaceship.Directions.IsHeadedTo(Right) {
		spaceship.Speed.X = 0
	}

	// Set the horizontal direction
	spaceship.Directions.SetHorizontal(Left)

	// Accelerate the spaceship horizontally
	spaceship.Speed.X += numeric.Number(spaceship.Level.AccelerateRate)

	// Limit the speed of the spaceship
	if spaceship.Speed.Magnitude().Float() > config.Config.Spaceship.MaximumSpeed {
		spaceship.Speed = spaceship.Speed.Normalize().Mul(numeric.Number(config.Config.Spaceship.MaximumSpeed))
	}

	// Check the horizontal boundaries and update the spaceship position
	spaceship.Position.X -= spaceship.Speed.X
	if spaceship.Position.X < 0 {
		spaceship.Position.X = 0
		spaceship.Speed.X = 0
	}

	go config.PlayAudio("spaceship_whoosh.wav", false)
}

// MoveRight moves the spaceship to the right.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is greater than the canvas width,
// it is set to the canvas width.
func (spaceship *Spaceship) MoveRight() {
	if spaceship.isFrozen() {
		return
	}

	// Brake the spaceship if it is moving in the opposite direction
	if spaceship.Directions.IsHeadedTo(Left) {
		spaceship.Speed.X = 0
	}

	// Set the horizontal direction
	spaceship.Directions.SetHorizontal(Right)

	// Accelerate the spaceship horizontally
	spaceship.Speed.X += numeric.Number(spaceship.Level.AccelerateRate)

	// Limit the speed of the spaceship
	if spaceship.Speed.Magnitude().Float() > config.Config.Spaceship.MaximumSpeed {
		spaceship.Speed = spaceship.Speed.Normalize().Mul(numeric.Number(config.Config.Spaceship.MaximumSpeed))
	}

	// Check the horizontal boundaries and update the spaceship position
	canvasDimensions := config.CanvasBoundingBox()
	spaceship.Position.X += spaceship.Speed.X
	if spaceship.Position.X.Float() > canvasDimensions.OriginalWidth {
		spaceship.Position.X = numeric.Number(canvasDimensions.OriginalWidth)
		spaceship.Speed.X = 0
	}

	go config.PlayAudio("spaceship_whoosh.wav", false)
}

// MoveUp moves the spaceship up.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is less than 0, it is set to 0.
func (spaceship *Spaceship) MoveUp() {
	if spaceship.isFrozen() {
		return
	}

	// Brake the spaceship if it is moving in the opposite direction
	if spaceship.Directions.IsHeadedTo(Down) {
		spaceship.Speed.Y = 0
	}

	// Set the vertical direction
	spaceship.Directions.SetVertical(Up)

	// Accelerate the spaceship vertically
	spaceship.Speed.Y += numeric.Number(spaceship.Level.AccelerateRate)

	// Limit the speed of the spaceship
	if spaceship.Speed.Magnitude().Float() > config.Config.Spaceship.MaximumSpeed {
		spaceship.Speed = spaceship.Speed.Normalize().Mul(numeric.Number(config.Config.Spaceship.MaximumSpeed))
	}

	// Check the vertical boundaries and update the spaceship position
	spaceship.Position.Y -= spaceship.Speed.Y
	if spaceship.Position.Y < 0 {
		spaceship.Position.Y = 0
		spaceship.Speed.Y = 0
	}

	go config.PlayAudio("spaceship_acceleration.wav", false)
}

// MoveTo moves the spaceship to the specified position.
// The spaceship's position is updated based on the delta.
// If the spaceship's position is less than 0, it is set to 0.
// If the spaceship's position is greater than the canvas width,
// it is set to the canvas width.
// If the spaceship's position is less than 0, it is set to 0.
// If the spaceship's position is greater than the canvas height,
// it is set to the canvas height.
func (spaceship *Spaceship) MoveTo(target numeric.Position) {
	if spaceship.isFrozen() {
		return
	}

	// Accelerate the spaceship
	spaceship.Speed = spaceship.Speed.AddN(spaceship.Level.AccelerateRate)

	// Limit the speed of the spaceship
	if spaceship.Speed.Magnitude().Float() > config.Config.Spaceship.MaximumSpeed {
		spaceship.Speed = spaceship.Speed.Normalize().Mul(numeric.Number(config.Config.Spaceship.MaximumSpeed))
	}

	// Calculate the delta
	delta := spaceship.Position.Sub(target).Normalize()

	// Brake the spaceship if it is moving in an opposite direction
	spaceship.Speed = spaceship.Speed.MulX(spaceship.Directions.Brake(delta))

	// Multiply the delta by the speed
	delta = delta.Mul(spaceship.Speed.Magnitude())

	// Set the new directions based on the delta
	spaceship.Directions.SetFromDelta(delta)

	// Update the spaceship position
	spaceship.Position = spaceship.Position.Sub(delta)

	// Check the horizontal boundaries
	canvasDimensions := config.CanvasBoundingBox()
	switch {
	case spaceship.Position.X < 0:
		spaceship.Position.X = 0
	case spaceship.Position.X.Float() > canvasDimensions.OriginalWidth:
		spaceship.Position.X = numeric.Number(canvasDimensions.OriginalWidth)
	}

	// Check the vertical boundaries
	switch {
	case spaceship.Position.Y < 0:
		spaceship.Position.Y = 0
	case spaceship.Position.Y.Float() > canvasDimensions.OriginalHeight:
		spaceship.Position.Y = numeric.Number(canvasDimensions.OriginalHeight)
	}

	go config.PlayAudio([...]string{
		"spaceship_acceleration.wav",
		"spaceship_whoosh.wav",
	}[numeric.RandomRange(0, 1).Int()], false)
}

// Penalize penalizes the spaceship by downgrading its level.
// The spaceship is downgraded by the specified number of levels.
// If the spaceship's level is less than 1, it is set to 1.
func (spaceship *Spaceship) Penalize(levels int) {
	for i := 0; i < levels && spaceship.Level.Progress > 0; i++ {
		if !spaceship.Level.Down() {
			return
		}
	}
}

// String returns a string representation of the spaceship.
func (spaceship Spaceship) String() string {
	return fmt.Sprintf("Spaceship (Lvl: %d, Pos: %s, State: %s)", spaceship.Level.Progress, spaceship.Position, spaceship.State)
}

// UpdateHighScore updates the high score of the spaceship.
func (spaceship *Spaceship) UpdateHighScore() {
	if spaceship.Level.Progress > spaceship.HighScore {
		spaceship.HighScore = spaceship.Level.Progress
	}
}

// UpdateState updates the state of the spaceship.
// If the time since the last state transition is greater than
// the spaceship state duration, the spaceship's state is set to Neutral.
func (spaceship *Spaceship) UpdateState() {
	switch spaceship.State {
	case Boosted:
		if time.Since(spaceship.lastStateTransition) > config.Config.Spaceship.BoostDuration {
			spaceship.Level.Cannons /= 2
			spaceship.Size.Width = numeric.Number(config.Config.Spaceship.Width)
			spaceship.Position.X += numeric.Number(config.Config.Spaceship.Width / 2)

			if spaceship.Level.Cannons == 0 {
				spaceship.Level.Cannons = 1
			}

			spaceship.State = Neutral
		}

	case Frozen:
		if time.Since(spaceship.lastStateTransition) > config.Config.Spaceship.FreezeDuration {
			spaceship.State = Neutral
		}

	case Damaged:
		if time.Since(spaceship.lastStateTransition) > config.Config.Spaceship.DamageDuration {
			spaceship.State = Neutral
		}

	}
}

// Embark creates a new spaceship.
// The spaceship is created at the bottom of the canvas.
// The spaceship's position, size, cooldown, level, and state are set.
func Embark() *Spaceship {
	canvasDimensions := config.CanvasBoundingBox()
	spaceship := Spaceship{
		Position:     numeric.Locate(canvasDimensions.OriginalWidth/2, canvasDimensions.OriginalHeight),
		Size:         numeric.Locate(config.Config.Spaceship.Width, config.Config.Spaceship.Height).ToBox(),
		CurrentScale: numeric.Ones(),
		Cooldown:     config.Config.Spaceship.Cooldown,
		Level: &SpaceshipLevel{
			AccelerateRate: numeric.Number(config.Config.Spaceship.Acceleration),
			Progress:       1,
			Cannons:        1,
		},
	}

	return &spaceship
}
