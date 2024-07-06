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
	if enemy.Type == Goodie {
		return
	}

	var sizeFactor, speedFactor float64
	var healthPoints, defense int
	nextType := enemy.Type
	switch enemy.Type {
	case Goodie, Freezer:
		return

	case Normal:
		sizeFactor, healthPoints, defense, speedFactor, nextType = 2, 2000, 500, 1.2, Berserker

	case Berserker:
		sizeFactor, healthPoints, defense, speedFactor, nextType = 3, 10000, 2500, 0.6, Annihilator

	case Annihilator:
		sizeFactor, healthPoints, defense, speedFactor = 4, 50000, 12500, 0.2

	}

	if rand.Intn(int(1.0/enemy.Level.BerserkLikeliness)) != 0 {
		return
	}

	enemy.Type = nextType
	enemy.Level.Speed = config.EnemyMaximumSpeed
	enemy.Level.HitPoints += healthPoints
	enemy.Level.Defense += defense
	enemy.Size.Width = config.EnemyWidth * sizeFactor
	enemy.Size.Height = config.EnemyHeight * sizeFactor
	enemy.Position.X -= config.EnemyWidth / sizeFactor
	enemy.Position.Y -= config.EnemyHeight / sizeFactor
	enemy.Level.Speed *= speedFactor
}

// Draw draws the enemy.
// The enemy is drawn as a rectangle with the specified color.
// The color is based on the type of the enemy.
func (enemy Enemy) Draw() {
	color := map[EnemyType]string{
		Goodie:      "green",
		Freezer:     "lightblue",
		Normal:      "gray",
		Berserker:   "red",
		Annihilator: "darkred",
	}[enemy.Type]
	config.DrawSpaceship(enemy.Position.X, enemy.Position.Y, enemy.Size.Width, enemy.Size.Height, false, color)
}

// Hit reduces the health points of the enemy.
// The damage is reduced by the defense of the enemy.
// If the damage is less than 0, it is set to 0.
func (enemy *Enemy) Hit(damage int) int {
	damage = damage - enemy.Level.Defense
	if damage < 0 {
		return 0
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
	enemy.Position.Y += enemy.Level.Speed
	if enemy.Type == Goodie {
		return
	}

	// If the spaceship is below the enemy, the enemy moves towards the spaceship.
	// The detection range is half of the canvas height.
	if spaceshipPosition.Y-enemy.Position.Y < config.CanvasHeight()/2 {
		// Check if the spaceship is on the left or right side of the enemy.
		switch {
		case enemy.Position.X < spaceshipPosition.X:
			enemy.direction = 1

		case enemy.Position.X > spaceshipPosition.X:
			enemy.direction = -1

		}
	} else {
		// Otherwise, the enemy moves randomly.
		enemy.direction += (rand.Intn(3) - 1)
		// Check if the enemy is at the edge of the canvas.
		switch {
		case enemy.Position.X <= 0:
			enemy.direction = 1

		case enemy.Position.X+enemy.Size.Width >= config.CanvasWidth():
			enemy.direction = -1

		}
	}

	enemy.Position.X += float64(enemy.direction)
}

// String returns the string representation of the enemy.
func (enemy Enemy) String() string {
	return fmt.Sprintf("%s (Lvl: %d, Pos: %s, HP: %d, Type: %s)", enemy.Name, enemy.Level.ID, enemy.Position, enemy.Level.HitPoints, enemy.Type)
}

// Surprise turns the enemy into a freezer or a goodie.
// If the enemy is a normal enemy, it has a chance to become a freezer or a goodie.
// The likelihood of the enemy becoming a freezer or a goodie is based on the SpecialtyLikeliness.
func (enemy *Enemy) Surprise() {
	if enemy.Type == Normal && rand.Intn(int(1.0/enemy.SpecialtyLikeliness)) == 0 {
		enemy.Type = [...]EnemyType{Freezer, Goodie}[rand.Intn(2)]
	}
}

// ToLevel sets the level of the enemy (up or down).
func (enemy *Enemy) ToLevel(levels int) {
	if levels < 1 {
		levels = 1
	}

	for enemy.Level.ID < levels {
		enemy.Level.Up()
	}

	for enemy.Level.ID > levels {
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
		y = rand.Float64() * (config.CanvasHeight()/2 - config.EnemyHeight)
	}

	return &Enemy{
		Position: objects.Position{
			X: rand.Float64() * (config.CanvasWidth() - config.EnemyWidth),
			Y: y,
		},
		Size: objects.Size{
			Width:  config.EnemyWidth,
			Height: config.EnemyHeight,
		},
		SpecialtyLikeliness: config.EnemySpecialtyLikeliness,
		Level: &EnemyLevel{
			ID:                1,
			Speed:             config.EnemyInitialSpeed,
			HitPoints:         100,
			Defense:           0,
			BerserkLikeliness: config.EnemyBerserkLikeliness,
		},
		Type: Normal,
		Name: name,
	}
}
