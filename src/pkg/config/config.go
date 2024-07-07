package config

import (
	_ "embed"
	"time"

	"gopkg.in/ini.v1"
)

//go:embed config.ini
var configFile []byte

// Config is the configuration of the game.
var Config config = func() (config config) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowBooleanKeys:          true,
		DebugFunc:                 Log,
		AllowNestedValues:         true,
		Insensitive:               true,
		SkipUnrecognizableLines:   true,
		UnescapeValueDoubleQuotes: true,
	}, configFile)
	ThrowError(err)
	ThrowError(cfg.MapTo(&config))

	return
}()

// config represents the configuration of the game.
type config struct {
	Bullet struct {
		Height               float64
		InitialDamage        int
		ModifierProgressStep int
		Speed                float64
		Width                float64
	}

	Enemy struct {
		Count                     int
		BerserkLikeliness         float64
		BerserkLikelinessProgress float64
		DefenseProgress           int
		Height                    float64
		HitpointProgress          int
		InitialDefense            int
		InitialHitpoints          int
		InitialSpeed              float64
		Margin                    float64
		MaximumSpeed              float64
		SpecialtyLikeliness       float64
		Width                     float64

		Annihilator struct {
			AgainFactor      int
			DefenseBoost     int
			HitpointsBoost   int
			SizeFactorBoost  float64
			SpeedFactorBoost float64
		} `ini:"Enemy.Annihilator"`

		Berserker struct {
			DefenseBoost     int
			HitpointsBoost   int
			SizeFactorBoost  float64
			SpeedFactorBoost float64
		} `ini:"Enemy.Berserker"`
	}

	MessageBox struct {
		BufferSize int
	}

	Spaceship struct {
		AnnihilatorPenalty int
		BerserkPenalty     int
		CannonProgress     int
		Cooldown           time.Duration
		DefaultPenalty     int
		Height             float64
		InitialSpeed       float64
		MaximumCannons     int
		MaximumSpeed       float64
		StateDuration      time.Duration
		Width              float64
	}

	Control struct {
		SwipeThreshold float64
	}

	Messages struct {
		GameStartedNoTouchDevice string
		GameStartedTouchDevice   string
		GameOver                 string
		HowToStartNoTouchDevice  string
		HowToStartTouchDevice    string
		SpaceshipFrozen          string

		Templates struct {
			EnemyDestroyed               templateString
			EnemyHit                     templateString
			SpaceshipDowngradedByEnemy   templateString
			SpaceshipUpgradedByEnemyKill templateString
			SpaceshipUpgradedByGoodie    templateString
		} `ini:"Messages.Templates"`
	}
}
