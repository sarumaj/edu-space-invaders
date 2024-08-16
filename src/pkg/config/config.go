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
		CriticalHitChance    float64
		GravityAmplifier     float64
		Height               float64
		InitialDamage        int
		ModifierProgressStep int
		Speed                float64
		Width                float64
	}

	Control struct {
		AudioEnabled                *bool
		BackgroundAnimationEnabled  *bool
		CollisionDetectionVersion   envVariable[int]
		Debug                       envVariable[bool]
		CriticalFramesPerSecondRate float64
		DesiredFramesPerSecondRate  float64
		GodMode                     envVariable[bool]
		SuspensionFrames            int
	}

	Enemy struct {
		AccelerationProgress      float64
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

		Messages struct {
			GamePausedNoTouchDevice   string
			GamePausedTouchDevice     string
			GameStartedNoTouchDevice  string
			GameStartedTouchDevice    string
			HowToRestartNoTouchDevice string
			HowToRestartTouchDevice   string
			HowToStartNoTouchDevice   string
			HowToStartTouchDevice     string
			ScoreBoardUpdated         string
			WaitForScoreBoardUpdate   string

			Templates struct {
				AllPlanetsDiscovered         templateString
				EnemyDestroyed               templateString
				EnemyHit                     templateString
				GameOver                     templateString
				Greeting                     templateString
				PerformanceDropped           templateString
				PerformanceImproved          templateString
				PlanetDiscovered             templateString
				PlanetImpactsSystem          templateString
				Prompt                       templateString
				SpaceshipDowngradedByEnemy   templateString
				SpaceshipFrozen              templateString
				SpaceshipStillFrozen         templateString
				SpaceshipUpgradedByEnemyKill templateString
				SpaceshipUpgradedByGoodie    templateString
			} `ini:"MessageBox.Messages.Templates"`
		} `ini:"MessageBox.Messages"`
	}

	Planet struct {
		AnomalyGravityModifier float64
		DiscoveryCooldown      time.Duration
		DiscoveryProbability   float64
		GravityStrength        float64
		MaximumRadius          float64
		MinimumRadius          float64
		SpeedRatio             float64
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
		ExperienceScaler   float64
		FreezeDuration     time.Duration
		FreezerPenalty     int
		Height             float64
		LogThrottling      time.Duration
		MaximumCannons     int
		MaximumLabelLength int
		MaximumSpeed       float64
		Width              float64
	}

	Star struct {
		Brightness         float64
		Count              int
		MinimumInnerRadius float64
		MinimumRadius      float64
		MinimumSpikes      float64
		MaximumInnerRadius float64
		MaximumRadius      float64
		MaximumSpikes      float64
		SpeedRatio         float64
	}
}
