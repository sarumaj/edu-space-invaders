package enemy

import (
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

// Enemy represents the enemy.
type Enemy struct {
	Name                string           // Name is the name of the enemy.
	Position            numeric.Position // Position is the position of the enemy.
	Size                numeric.Size     // Size is the size of the enemy.
	SpecialtyLikeliness float64          // SpecialtyLikeliness is the likelihood of the enemy being a goodie or a freezer (expected to be lower than 1).
	Level               *EnemyLevel      // Level is the level of the enemy.
	Type                EnemyType        // Type is the type of the enemy.
}

// Berserk turns the enemy into a berserker or an annihilator.
// If the enemy is a goodie or a freezer, it does nothing.
// If the enemy is a normal enemy, it has a chance to become a berserker.
// If the enemy is a berserker, it has a chance to become an annihilator.
// If the enemy is an annihilator, it increases its size, health points and defense.
func (enemy *Enemy) Berserk() {
	if !numeric.SampleUniform(enemy.Level.BerserkLikeliness) {
		return
	}

	boost := struct {
		sizeFactor, speedFactor numeric.Number
		healthPoints, defense   int
		nextType                EnemyType
	}{
		sizeFactor: 1,
		nextType:   enemy.Type,
	}
	switch enemy.Type {
	case Goodie, Freezer:
		return

	case Normal:
		boost.defense = config.Config.Enemy.Berserker.DefenseBoost
		boost.healthPoints = config.Config.Enemy.Berserker.HitpointsBoost
		boost.speedFactor = numeric.Number(config.Config.Enemy.Berserker.SpeedModifier)
		boost.sizeFactor = numeric.Number(config.Config.Enemy.Berserker.SizeFactorBoost)
		boost.nextType = Berserker

	case Berserker:
		boost.defense = config.Config.Enemy.Annihilator.DefenseBoost
		boost.healthPoints = config.Config.Enemy.Annihilator.HitpointsBoost
		boost.speedFactor = numeric.Number(config.Config.Enemy.Annihilator.SpeedModifier)
		boost.sizeFactor = numeric.Number(config.Config.Enemy.Annihilator.SizeFactorBoost)
		boost.nextType = Annihilator

	case Annihilator:
		boost.defense = config.Config.Enemy.Annihilator.YetAgainFactor * config.Config.Enemy.Annihilator.DefenseBoost
		boost.healthPoints = config.Config.Enemy.Annihilator.YetAgainFactor * config.Config.Enemy.Annihilator.HitpointsBoost
		boost.speedFactor = numeric.Number(config.Config.Enemy.Annihilator.SpeedModifier)
		boost.sizeFactor = numeric.Number(config.Config.Enemy.Annihilator.SizeFactorBoost)

	}

	if enemy.Type != boost.nextType {
		enemy.Resize(boost.sizeFactor)
		enemy.Type = boost.nextType
	}

	enemy.Level.HitPoints += boost.healthPoints
	enemy.Level.Defense += boost.healthPoints
	enemy.Level.Speed *= boost.speedFactor
}

// BerserkGivenAncestor increases the chance of the enemy to become a berserker or an annihilator
// by repeating the berserk for the new enemy given the enemy type of the ancestor.
func (enemy *Enemy) BerserkGivenAncestor(oldType EnemyType) {
	enemy.Berserk()

	// repeat the berserk for the new enemy
	switch oldType {
	case Annihilator:
		enemy.Berserk()
		fallthrough

	case Berserker:
		enemy.Berserk()
		fallthrough

	default:
		enemy.Berserk()

	}
}

// Destroy destroys the enemy.
// The health points of the enemy are set to 0.
func (enemy *Enemy) Destroy() {
	enemy.Level.HitPoints = 0
	go config.PlayAudio("enemy_destroyed.wav", false)
}

// Draw draws the enemy.
// The enemy is drawn as a rectangle with the specified color.
// The color is based on the type of the enemy.
// If the control to draw object labels is enabled, the name of the enemy is drawn.
// If the control to draw enemy hitpoint bars is enabled, the hitpoint bar is drawn.
func (enemy Enemy) Draw() {
	var label string
	if config.Config.Control.DrawObjectLabels.Get() {
		label = enemy.Name
	}

	var statusValues []float64
	var statusColors []string
	if enemy.Type != Goodie && config.Config.Control.DrawEnemyHitpointBars.Get() {
		statusValues = append(statusValues, float64(enemy.Level.HitPoints)/float64(enemy.Level.HitPoints+enemy.Level.HitPointsLoss))
		statusColors = append(statusColors, "rgba(240, 0, 0, 0.8)")
	}

	config.DrawSpaceship(
		enemy.Position.Pack(),
		enemy.Size.Pack(),
		enemy.Type == Goodie, // Face up if the enemy is a goodie
		map[EnemyType]string{
			Goodie:      "Chartreuse",
			Freezer:     "DeepSkyBlue",
			Normal:      "DarkseaGreen",
			Berserker:   "Firebrick",
			Annihilator: "MidnightBlue",
		}[enemy.Type],
		label,
		statusValues,
		statusColors,
	)
}

// GetPenalty returns the penalty of the enemy.
func (enemy Enemy) GetPenalty() int {
	if v, ok := map[EnemyType]int{
		Freezer:     config.Config.Enemy.Freezer.Penalty,
		Normal:      config.Config.Enemy.DefaultPenalty,
		Berserker:   config.Config.Enemy.Berserker.Penalty,
		Annihilator: config.Config.Enemy.Annihilator.Penalty,
	}[enemy.Type]; ok {
		return v + enemy.Level.Progress
	}

	return 0
}

// Hit reduces the health points of the enemy.
// The damage is reduced by the defense of the enemy.
// If the damage is less than 0, it is set to 0.
func (enemy *Enemy) Hit(damage int) int {
	damage = damage - enemy.Level.Defense - numeric.RandomRange(0, enemy.Level.Defense*enemy.Level.Progress).Int()
	if damage < 0 {
		damage = numeric.RandomRange(0, config.Config.Bullet.InitialDamage).Int()
	}

	if damage > enemy.Level.HitPoints {
		damage = enemy.Level.HitPoints
	}

	enemy.Level.HitPointsLoss += damage
	enemy.Level.HitPoints -= damage

	go config.PlayAudio("enemy_hit.wav", false)

	return damage
}

// IsDestroyed returns true if the enemy is destroyed.
func (enemy Enemy) IsDestroyed() bool {
	return enemy.Level.HitPoints <= 0
}

// Move moves the enemy.
// The enemy moves downwards and changes its horizontal direction.
// If the enemy is a goodie, it moves only downwards and does not change its horizontal direction.
// The direction of the enemy is based on the position of the spaceship.
// If the spaceship is below the enemy, the enemy moves towards the spaceship.
// Otherwise, the enemy moves randomly.
func (enemy *Enemy) Move(spaceshipPosition numeric.Position) {
	if enemy.Type == Goodie {
		enemy.Position.Y += numeric.Number(enemy.Level.Speed)
		return
	}

	// Calculate the horizontal and vertical distances to the spaceship
	delta := spaceshipPosition.Sub(enemy.Position)

	// Calculate the distance to the spaceship
	distance := delta.Magnitude()

	// Define the strength formula
	strength := numeric.Number(enemy.Level.Progress) / (distance + 1) // Add 1 to avoid division by zero

	// Add randomness to the chase based on strength
	delta = delta.Add(numeric.Locate(
		numeric.RandomRange(-0.5, 0.5), // Random number between -0.5 and 0.5
		numeric.RandomRange(-1, 0),     // Random number between -1 and 0
	)).Mul(strength)

	// Limit the speed of the enemy
	if delta.Magnitude().Float() > config.Config.Enemy.MaximumSpeed {
		delta = delta.Normalize().Mul(numeric.Number(config.Config.Enemy.MaximumSpeed))
	}

	// Move down using the speed
	enemy.Position.Y += enemy.Level.Speed

	// Dash off screen if the spaceship is below the enemy and the enemy is close to the spaceship
	if enemy.Position.Y > spaceshipPosition.Y && enemy.Position.Sub(spaceshipPosition).X.Abs() < enemy.Size.ToVector().Magnitude() {
		delta.Y = numeric.Number(config.Config.Enemy.MaximumSpeed)
	}

	// Move horizontally and vertically towards the spaceship
	enemy.Position = enemy.Position.Add(delta)
}

// Resize resizes the enemy.
// The enemy is resized by the specified scale.
// The position is adjusted to the center of the new size.
func (enemy *Enemy) Resize(scale numeric.Number) {
	enemy.Size, enemy.Position = enemy.Size.Resize(scale, enemy.Position)
}

// String returns the string representation of the enemy.
func (enemy Enemy) String() string {
	return fmt.Sprintf("%s (Lvl: %d, Pos: %s, HP: %d, Type: %s)", enemy.Name, enemy.Level.Progress, enemy.Position, enemy.Level.HitPoints, enemy.Type)
}

// Surprise turns the enemy into a freezer or a goodie.
// If the enemy is a normal enemy, it has a chance to become a freezer or a goodie.
// The likelihood of the enemy becoming a freezer or a goodie is based on the SpecialtyLikeliness.
// The stats are the number of enemies by type.
// They are used to lower the likelihood of the enemy becoming a goodie.
func (enemy *Enemy) Surprise(stats map[EnemyType]int) {
	goodies, total := 0.0, 0.0
	for k, v := range stats {
		if k == Goodie {
			goodies = float64(v)
		}
		total += float64(v)
	}

	if total == 0.0 {
		total = 1.0
	}

	if enemy.Type == Normal && numeric.SampleUniform(enemy.SpecialtyLikeliness) {
		enemy.Type = [...]EnemyType{Freezer, Goodie}[numeric.RandomRange(0, (total-goodies)/total).Int()]
	}
}

// ToProgressLevel sets the progress level of the enemy (up or down).
func (enemy *Enemy) ToProgressLevel(progress int) {
	if progress < 1 {
		progress = 1
	}

	for enemy.Level.Progress < progress {
		enemy.Level.Up()
	}

	for enemy.Level.Progress > progress {
		enemy.Level.Down()
	}
}

// Challenge creates a new enemy.
// If the name is empty, a random name is generated.
// If randomY is true, the enemy is placed at a random Y position
// in the top half of the canvas.
// Otherwise, the enemy is placed at the top of the canvas.
// The enemy is a normal enemy.
// The enemy has the initial speed, hit points and defense.
// The likelihood of the enemy becoming a goodie is based on the GoodieLikeliness.
// The likelihood of the enemy becoming a berserker is based on the BerserkLikeliness.
// The enemy has the initial level.
func Challenge(name string, randomY bool) *Enemy {
	if name == "" {
		name = randomdata.SillyName()
	}

	canvasDimensions := config.CanvasBoundingBox()
	var y numeric.Number
	if randomY {
		y = numeric.RandomRange(0, canvasDimensions.OriginalHeight/2)
	}

	enemy := Enemy{
		Position:            numeric.Locate(numeric.RandomRange(0, canvasDimensions.OriginalWidth), y),
		Size:                numeric.Locate(config.Config.Enemy.Width, config.Config.Enemy.Height).ToBox(),
		SpecialtyLikeliness: config.Config.Enemy.SpecialtyLikeliness,
		Level: &EnemyLevel{
			Progress:          1,
			Speed:             numeric.Number(config.Config.Enemy.InitialSpeed),
			HitPoints:         config.Config.Enemy.InitialHitpoints,
			Defense:           config.Config.Enemy.InitialDefense,
			BerserkLikeliness: config.Config.Enemy.BerserkLikeliness,
		},
		Type: Normal,
		Name: name,
	}

	return &enemy
}
