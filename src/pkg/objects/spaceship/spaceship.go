package spaceship

import (
	"fmt"
	"slices"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/bullet"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/planet"
)

// Spaceship represents the player's spaceship.
type Spaceship struct {
	IsAdmiral           bool                       // IsAdmiral is true if the spaceship is the admiral.
	Commandant          string                     // Commandant is the name of the spaceship's commander.
	Position            numeric.Position           // Position of the spaceship
	Speed               numeric.Position           // Speed of the spaceship in both directions
	Cooldown            time.Duration              // Time between shots
	Directions          Directions                 // Directions the spaceship can move
	Size                numeric.Size               // Size of the spaceship
	Bullets             bullet.Bullets             // Bullets fired by the spaceship
	Level               *SpaceshipLevel            // Spaceship level
	State               SpaceshipState             // Spaceship state
	HighScore           int                        // HighScore is the high score of the spaceship.
	lastFired           time.Time                  // Last time the spaceship fired
	lastStateTransition time.Time                  // Last time the spaceship changed state
	lastDiscovery       time.Time                  // Last time the spaceship discovered a planet
	discoveredPlanets   map[planet.PlanetType]bool // Discovered planets
}

// ifFrozen checks if the spaceship can move or shoot.
// If the spaceship is in the Frozen state, a message is sent
// to the message box indicating that the spaceship is still frozen.
// The message is throttled based on the log throttling duration.
func (spaceship *Spaceship) ifFrozen() bool {
	if spaceship.State == Frozen {
		config.SendMessageThrottled(
			config.Execute(config.Config.MessageBox.Messages.SpaceshipStillFrozen,
				config.Template{
					"FreezeDuration": time.Until(spaceship.lastStateTransition.
						Add(config.Config.Spaceship.FreezeDuration)).
						Round(config.Config.MessageBox.ChannelLogThrottling),
				},
			), false, true, config.Config.MessageBox.ChannelLogThrottling)

		return true
	}

	return false
}

// ApplyRepulsion applies repulsion to the spaceship and the enemy.
// The repulsion is applied based on the spaceship's and enemy speed and direction.
func (spaceship *Spaceship) ApplyRepulsion(e enemy.Enemy) numeric.Position {
	// Calculate the effective area (as substitute for mass)
	spaceshipArea, enemyArea := spaceship.Size.Area(), e.Size.Area()
	effectiveArea := spaceshipArea * enemyArea / (spaceshipArea + enemyArea)

	// Calculate the minimum translation vector (MTV)
	var mtv numeric.Position
	switch config.Config.Control.CollisionDetectionVersion.Get() {
	case 1:
		mtv = numeric.GetRectangularVertices(spaceship.Position, spaceship.Size, true).
			Vertices().
			MinimumTranslationVector(numeric.GetRectangularVertices(e.Position, e.Size, true).
				Vertices())

	case 2:
		mtv = numeric.GetSpaceshipVerticesV1(spaceship.Position, spaceship.Size, true).
			Vertices().
			MinimumTranslationVector(numeric.GetSpaceshipVerticesV1(e.Position, e.Size, e.Type == enemy.Goodie).
				Vertices())

	case 3:
		mtv = numeric.GetSpaceshipVerticesV2(spaceship.Position, spaceship.Size, true).
			Vertices().
			MinimumTranslationVector(numeric.GetSpaceshipVerticesV2(e.Position, e.Size, e.Type == enemy.Goodie).
				Vertices())

	}

	if mtv.IsZero() { // No collision
		return e.Position
	}

	// Apply the displacements
	spaceship.Position = spaceship.Position.Add(mtv.Mul(effectiveArea / spaceshipArea))
	return e.Position.Sub(mtv.Mul(effectiveArea / enemyArea))
}

// ChangeState changes the state of the spaceship.
// If the state is Boosted, the spaceship's cannons are doubled
// and its size is doubled. If the number of cannons exceeds
// the maximum number of cannons, it is set to the maximum number.
func (spaceship *Spaceship) ChangeState(state SpaceshipState) {
	spaceship.lastStateTransition = time.Now()
	if spaceship.State == state {
		return
	}

	if state == Boosted {
		spaceship.Level.Cannons *= 2
		if spaceship.Level.Cannons > config.Config.Spaceship.MaximumCannons {
			spaceship.Level.Cannons = config.Config.Spaceship.MaximumCannons
		}

		spaceship.Resize(numeric.Number(config.Config.Spaceship.BoostScaleSizeFactor))
	}

	spaceship.State = state

	switch spaceship.State {
	case Boosted:
		go config.PlayAudio("spaceship_boost.wav", false)

	case Frozen:
		go config.PlayAudio("spaceship_freeze.wav", false)

	case Damaged:
		go config.PlayAudio("spaceship_crash.wav", false)

	}
}

// DetectCollision checks if the spaceship has collided with an enemy.
// The collision detection is based on the separating axis theorem.
// The separating axis theorem states that if two convex shapes do not overlap
// on all axes, then they do not overlap.
// It uses the exact vertices of the spaceship and the enemy.
func (spaceship Spaceship) DetectCollision(e enemy.Enemy) bool {
	switch config.Config.Control.CollisionDetectionVersion.Get() {
	case 1:
		return !numeric.GetRectangularVertices(spaceship.Position, spaceship.Size, true).
			Vertices().
			HasSeparatingAxis(numeric.GetRectangularVertices(e.Position, e.Size, true).
				Vertices())

	case 2:
		return !numeric.
			GetSpaceshipVerticesV1(spaceship.Position, spaceship.Size, true).
			Vertices().
			HasSeparatingAxis(numeric.GetSpaceshipVerticesV1(e.Position, e.Size, e.Type == enemy.Goodie).
				Vertices())

	case 3:
		return !numeric.
			GetSpaceshipVerticesV2(spaceship.Position, spaceship.Size, true).
			Vertices().
			HasSeparatingAxis(numeric.GetSpaceshipVerticesV2(e.Position, e.Size, e.Type == enemy.Goodie).
				Vertices())

	}

	return false
}

// Discover discovers the planet.
// If the spaceship is close to the planet, the planet has not been discovered,
// the planet has not been discovered recently, and the planet is discovered based on the probability,
// the planet will be discovered.
// If all planets have been discovered, the spaceship will promote its commander to admiral.
func (spaceship *Spaceship) Discover(p *planet.Planet) bool {
	switch {
	case
		!p.Type.IsPlanet(), // If the celestial object is not an actual planet
		!p.WithinRange(spaceship.Position.Add(spaceship.Size.Half().ToVector())),                                                     // If the spaceship is not within range of the planet
		spaceship.discoveredPlanets[p.Type],                                                                                          // If the planet has been discovered
		time.Since(spaceship.lastDiscovery) < config.Config.Planet.DiscoveryCooldown*time.Duration(len(spaceship.discoveredPlanets)), // If a planet has been discovered recently
		!numeric.SampleUniform(config.Config.Planet.DiscoveryProbability):                                                            // If the planet is not discovered based on the probability

		return false
	}

	spaceship.lastDiscovery = time.Now()
	spaceship.discoveredPlanets[p.Type] = true

	return true
}

// Discovered returns the list of discovered planets.
func (spaceship *Spaceship) Discovered() []string {
	var discovered []string
	for t, d := range spaceship.discoveredPlanets {
		if d {
			discovered = append(discovered, t.String())
		}
	}

	slices.Sort(discovered)
	return discovered
}

// Draw draws the spaceship on the canvas.
// The spaceship is drawn in white color if it is in the Neutral state.
// The spaceship is drawn in dark red color if it is in the Damaged state.
// The spaceship is drawn in yellow color if it is in the Boosted state.
// The spaceship is drawn in blue color if it is in the Frozen state.
// If the control to draw object labels is enabled, the spaceship is drawn with the commander's name.
// If the control to draw the spaceship experience bar is enabled, the spaceship is drawn with the experience bar.
// If the control to draw the spaceship discovery progress bar is enabled, the spaceship is drawn with the discovery progress bar.
// If the control to draw the spaceship shield is enabled, the spaceship is drawn with the shield.
func (spaceship Spaceship) Draw() {
	var label string
	if config.Config.Control.DrawObjectLabels.Get() {
		label = spaceship.Commandant
	}

	var statusValues []float64
	var statusColors []string
	if config.Config.Control.DrawSpaceshipShield.Get() {
		// Reverse the shield charge to draw the damage impact on the shield
		statusValues = append(statusValues, 1-spaceship.Level.Shield.Health().Float())
		statusColors = append(statusColors, "rgba(240, 10, 10, 0.8)") // Red
	}

	if config.Config.Control.DrawSpaceshipExperienceBar.Get() {
		statusValues = append(statusValues, float64(spaceship.Level.Experience)/float64(spaceship.Level.GetRequiredExperience()))
		statusColors = append(statusColors, "rgba(240, 240, 0, 0.8)") // Yellow
	}

	if config.Config.Control.DrawSpaceshipDiscoveryProgressBar.Get() {
		statusValues = append(statusValues, float64(len(spaceship.discoveredPlanets))/float64(planet.PlanetsCount))
		statusColors = append(statusColors, "rgba(0, 0, 240, 0.8)") // Blue
	}

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
		label,
		statusValues,
		statusColors,
	)
}

// Fire fires bullets from the spaceship.
// The number of bullets fired is equal to the number of cannons
// the spaceship has. The damage of the bullets is calculated
// based on the spaceship's level.
// The trajectory of the bullets is skewed based on the position
// of the cannon.
func (spaceship *Spaceship) Fire() {
	if spaceship.ifFrozen() {
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
			numeric.Number(float64(i)/float64(spaceship.Level.Cannons+1)),
			// Speed boost of the bullet
			spaceship.Level.AccelerateRate,
		)
	}

	spaceship.lastFired = time.Now()

	go config.PlayAudio("spaceship_cannon_fire.wav", false)
}

// FixPosition fixes the position of the spaceship based on the canvas boundaries
// to prevent the spaceship from going out of bounds.
func (spaceship *Spaceship) FixPosition() {
	canvasDimensions := config.CanvasBoundingBox()
	if halfWidth := spaceship.Size.Width / 2; spaceship.Position.X-halfWidth < 0 {
		spaceship.Position.X = halfWidth
	} else if (spaceship.Position.X + halfWidth).Float() > canvasDimensions.OriginalWidth {
		spaceship.Position.X = numeric.Number(canvasDimensions.OriginalWidth) - halfWidth
	}

	if halfHeight := spaceship.Size.Height / 2; spaceship.Position.Y-halfHeight < 0 {
		spaceship.Position.Y = halfHeight
	} else if (spaceship.Position.Y + halfHeight).Float() > canvasDimensions.OriginalHeight {
		spaceship.Position.Y = numeric.Number(canvasDimensions.OriginalHeight) - halfHeight
	}
}

// GetBulletDamage returns the damage of the bullets fired by the spaceship.
func (spaceship Spaceship) GetBulletDamage() int {
	// Calculate the base damage
	base := config.Config.Bullet.InitialDamage + spaceship.Level.Progress
	// Calculate the modifier
	modifier := (spaceship.Level.Progress/config.Config.Bullet.ModifierProgressStep + 1) * spaceship.Level.Cannons

	damage := base*modifier + numeric.RandomRange(0, base*modifier).Int()

	// Allow critical hit
	if numeric.SampleUniform(config.Config.Bullet.CriticalHitChance) {
		damage *= config.Config.Bullet.CriticalHitFactor
	}

	// Amplify the damage if the spaceship is an admiral
	if spaceship.IsAdmiral {
		damage *= config.Config.Spaceship.AdmiralDamageAmplifier
	}

	// Return the damage
	return damage
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
	if spaceship.ifFrozen() {
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
	spaceship.FixPosition()

	go config.PlayAudio("spaceship_deceleration.wav", false)
}

// MoveLeft moves the spaceship to the left.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is less than 0, it is set to 0.
func (spaceship *Spaceship) MoveLeft() {
	if spaceship.ifFrozen() {
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
	spaceship.FixPosition()

	go config.PlayAudio("spaceship_whoosh.wav", false)
}

// MoveRight moves the spaceship to the right.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is greater than the canvas width,
// it is set to the canvas width.
func (spaceship *Spaceship) MoveRight() {
	if spaceship.ifFrozen() {
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
	spaceship.Position.X += spaceship.Speed.X
	spaceship.FixPosition()

	go config.PlayAudio("spaceship_whoosh.wav", false)
}

// MoveUp moves the spaceship up.
// The spaceship's position is updated based on the spaceship's speed.
// If the spaceship's position is less than 0, it is set to 0.
func (spaceship *Spaceship) MoveUp() {
	if spaceship.ifFrozen() {
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
	spaceship.FixPosition()

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
	if spaceship.State.AnyOf(Frozen) {
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

	// Fix the spaceship position
	spaceship.FixPosition()

	go config.PlayAudio([...]string{
		"spaceship_acceleration.wav",
		"spaceship_whoosh.wav",
	}[numeric.RandomRange(0, 1).Int()], false)
}

// Penalize penalizes the spaceship by downgrading its level.
// The spaceship is downgraded by the specified number of levels.
// If the spaceship's level has decreased, it returns true.
func (spaceship *Spaceship) Penalize(levels int) bool {
	if levels < 1 {
		return false
	}

	currentLvl := spaceship.Level.Progress
	for i := 0; i < levels && spaceship.Level.Progress > 0; i++ {
		if !spaceship.Level.Down() {
			return spaceship.Level.Progress < currentLvl
		}
	}

	return spaceship.Level.Progress < currentLvl
}

// Resize resizes the spaceship based on the scale.
// The spaceship's size is updated based on the scale.
// The spaceship's position is centered based on the new size and
// fixed based on the canvas boundaries.
func (spaceship *Spaceship) Resize(scale numeric.Number) {
	spaceship.Size, spaceship.Position = spaceship.Size.Resize(scale, spaceship.Position)
	if scale > 1 {
		spaceship.FixPosition()
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
			spaceship.Resize(1 / numeric.Number(config.Config.Spaceship.BoostScaleSizeFactor))
			spaceship.Level.Cannons /= 2

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
func Embark(commandant string) *Spaceship {
	canvasDimensions := config.CanvasBoundingBox()
	spaceship := Spaceship{
		Commandant: commandant,
		Position: numeric.Locate(
			canvasDimensions.OriginalWidth/2,
			canvasDimensions.OriginalHeight-config.Config.Spaceship.Height/2,
		),
		Size:     numeric.Locate(config.Config.Spaceship.Width, config.Config.Spaceship.Height).ToBox(),
		Cooldown: config.Config.Spaceship.Cooldown,
		Level: &SpaceshipLevel{
			AccelerateRate: numeric.Number(config.Config.Spaceship.Acceleration),
			Progress:       1,
			Cannons:        1,
			Shield: &Shield{
				Charge:         1,
				Capacity:       1,
				ChargeDuration: config.Config.Spaceship.ShieldChargeDuration,
			},
		},
		discoveredPlanets: make(map[planet.PlanetType]bool),
	}

	if spaceship.Commandant == "" {
		spaceship.Commandant = randomdata.FullName(randomdata.RandomGender)
	}

	return &spaceship
}
