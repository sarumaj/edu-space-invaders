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
		CountProgressStep         int
		BerserkLikeliness         float64
		BerserkLikelinessProgress float64
		DefenseProgress           int
		Height                    float64
		HitpointProgress          int
		InitialDefense            int
		InitialHitpoints          int
		InitialSpeed              float64
		Margin                    float64
		MaximumCount              int
		MaximumSpeed              float64
		Regenerate                *bool
		SpecialtyLikeliness       float64
		Width                     float64

		Annihilator struct {
			DefenseBoost     int
			HitpointsBoost   int
			SizeFactorBoost  float64
			SpeedFactorBoost float64
			YetAgainFactor   int
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
		Acceleration       float64
		AnnihilatorPenalty int
		BerserkPenalty     int
		BoostDuration      time.Duration
		CannonProgress     int
		Cooldown           time.Duration
		DamageDuration     time.Duration
		DefaultPenalty     int
		FreezeDuration     time.Duration
		FreezerPenalty     int
		Height             float64
		LogThrottling      time.Duration
		MaximumCannons     int
		MaximumSpeed       float64
		Width              float64
	}

	Star struct {
		Brightness    float64
		Count         int
		MinimumRadius float64
		MinimumSpikes float64
		MaximumRadius float64
		MaximumSpikes float64
		SpeedRatio    float64
	}

	Control struct {
		AudioEnabled                *bool
		Debug                       envVariable[bool]
		CriticalFramesPerSecondRate float64
		DesiredFramesPerSecondRate  float64
		SuspensionFrames            int
		SwipeProximityRange         float64
	}

	Messages struct {
		GamePausedNoTouchDevice   string
		GamePausedTouchDevice     string
		GameStartedNoTouchDevice  string
		GameStartedTouchDevice    string
		HowToRestartNoTouchDevice string
		HowToRestartTouchDevice   string
		HowToStartNoTouchDevice   string
		HowToStartTouchDevice     string

		Templates struct {
			EnemyDestroyed               templateString
			EnemyHit                     templateString
			GameOver                     templateString
			PerformanceDropped           templateString
			PerformanceImproved          templateString
			SpaceshipDowngradedByEnemy   templateString
			SpaceshipFrozen              templateString
			SpaceshipStillFrozen         templateString
			SpaceshipUpgradedByEnemyKill templateString
			SpaceshipUpgradedByGoodie    templateString
		} `ini:"Messages.Templates"`
	}
}
