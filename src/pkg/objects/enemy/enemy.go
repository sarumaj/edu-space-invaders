package enemy

import (
	"fmt"
	"math/rand"

	"github.com/Pallinder/go-randomdata"
	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects"
)

// Enemy represents the enemy.
type Enemy struct {
	Name                string           // Name is the name of the enemy.
	Position            objects.Position // Position is the position of the enemy.
	direction           int              // direction is the horizontal direction the enemy is moving.
	Size                objects.Size     // Size is the size of the enemy.
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
	boost := struct {
		sizeFactor, speedFactor float64
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
		boost.speedFactor = config.Config.Enemy.Berserker.SpeedFactorBoost
		boost.sizeFactor = config.Config.Enemy.Berserker.SizeFactorBoost
		boost.nextType = Berserker

	case Berserker:
		boost.defense = config.Config.Enemy.Annihilator.DefenseBoost
		boost.healthPoints = config.Config.Enemy.Annihilator.HitpointsBoost
		boost.speedFactor = config.Config.Enemy.Annihilator.SpeedFactorBoost
		boost.sizeFactor = config.Config.Enemy.Annihilator.SizeFactorBoost
		boost.nextType = Annihilator

	case Annihilator:
		boost.defense = config.Config.Enemy.Annihilator.YetAgainFactor * config.Config.Enemy.Annihilator.DefenseBoost
		boost.healthPoints = config.Config.Enemy.Annihilator.YetAgainFactor * config.Config.Enemy.Annihilator.HitpointsBoost
		boost.speedFactor = config.Config.Enemy.Annihilator.SpeedFactorBoost
		boost.sizeFactor = config.Config.Enemy.Annihilator.SizeFactorBoost

	}

	if rand.Intn(int(1.0/enemy.Level.BerserkLikeliness)) != 0 {
		return
	}

	enemy.Type = boost.nextType
	enemy.Level.HitPoints += boost.healthPoints
	enemy.Level.Defense += boost.healthPoints
	enemy.Level.Speed *= boost.speedFactor

	if config.Config.Enemy.Width*boost.sizeFactor < config.CanvasWidth() && config.Config.Enemy.Height*boost.sizeFactor < config.CanvasHeight() {
		enemy.Size.Width = objects.Number(config.Config.Enemy.Width * boost.sizeFactor)
		enemy.Size.Height = objects.Number(config.Config.Enemy.Height * boost.sizeFactor)
		enemy.Position.X -= objects.Number(config.Config.Enemy.Width / boost.sizeFactor)
		enemy.Position.Y -= objects.Number(config.Config.Enemy.Height / boost.sizeFactor)
	}
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

// Draw draws the enemy.
// The enemy is drawn as a rectangle with the specified color.
// The color is based on the type of the enemy.
func (enemy Enemy) Draw() {
	config.DrawSpaceship(
		enemy.Position.Pack(),
		enemy.Size.Pack(),
		false,
		map[EnemyType]string{
			Goodie:      "green",
			Freezer:     "lightblue",
			Normal:      "gray",
			Berserker:   "red",
			Annihilator: "darkred",
		}[enemy.Type],
	)
}

// Hit reduces the health points of the enemy.
// The damage is reduced by the defense of the enemy.
// If the damage is less than 0, it is set to 0.
func (enemy *Enemy) Hit(damage int) int {
	damage = damage - rand.Intn(enemy.Level.Defense)
	if damage < 0 {
		damage = rand.Intn(config.Config.Bullet.InitialDamage)
	}

	enemy.Level.HitPoints -= damage
	return damage
}

// Move moves the enemy.
// The enemy moves downwards and changes its horizontal direction.
// If the enemy is a goodie, it moves only downwards and does not change its horizontal direction.
// The direction of the enemy is based on the position of the spaceship.
// If the spaceship is below the enemy, the enemy moves towards the spaceship.
// Otherwise, the enemy moves randomly.
func (enemy *Enemy) Move(spaceshipPosition objects.Position) {
	enemy.Position.Y += objects.Number(enemy.Level.Speed)
	if enemy.Type == Goodie {
		return
	}

	// If the spaceship is below the enemy, the enemy moves towards the spaceship.
	// The detection range is half of the canvas height.
	if spaceshipPosition.Y.Float()-enemy.Position.Y.Float() < config.CanvasHeight()/2 {
		// Check if the spaceship is on the left or right side of the enemy.
		switch {
		case enemy.Position.X < spaceshipPosition.X:
			enemy.direction = 1

		case enemy.Position.X > spaceshipPosition.X:
			enemy.direction = -1

		}

		// Surprise dash of the enemy towards spaceship
		enemy.Position.Y += objects.Number(rand.Float64() * enemy.Level.Speed)
		enemy.direction *= (rand.Intn(3) + 1)
	} else {
		// Otherwise, the enemy moves randomly.
		enemy.direction += (rand.Intn(3) - 1)
		// Check if the enemy is at the edge of the canvas.
		switch {
		case enemy.Position.X <= 0:
			enemy.direction = 1

		case enemy.Position.X.Float()+enemy.Size.Width.Float() >= config.CanvasWidth():
			enemy.direction = -1

		}
	}

	enemy.Position.X += objects.Number(enemy.direction)
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

	if enemy.Type == Normal && rand.Intn(int(1.0/enemy.SpecialtyLikeliness)) == 0 {
		index := 0
		if rand.Float64()*(total-goodies)/total >= 0.5 {
			index = 1
		}
		enemy.Type = [...]EnemyType{Freezer, Goodie}[index]

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

	y := 0.0
	if randomY {
		y = rand.Float64() * (config.CanvasHeight()/2 - config.Config.Enemy.Height)
	}

	return &Enemy{
		Position: objects.Position{
			X: objects.Number(rand.Float64() * (config.CanvasWidth() - config.Config.Enemy.Width)),
			Y: objects.Number(y),
		},
		Size: objects.Size{
			Width:  objects.Number(config.Config.Enemy.Width),
			Height: objects.Number(config.Config.Enemy.Height),
		},
		SpecialtyLikeliness: config.Config.Enemy.SpecialtyLikeliness,
		Level: &EnemyLevel{
			Progress:          1,
			Speed:             config.Config.Enemy.InitialSpeed,
			HitPoints:         config.Config.Enemy.InitialHitpoints,
			Defense:           config.Config.Enemy.InitialDefense,
			BerserkLikeliness: config.Config.Enemy.BerserkLikeliness,
		},
		Type: Normal,
		Name: name,
	}
}
