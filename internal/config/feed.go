package config

import "time"

// Feed holds feed-related config.
type Feed struct {
	MaxPosts         int           `env:"MAX_POSTS" envDefault:"15"`
	Keywords         []string      `env:"KEYWORDS" envSeparator:"," envDefault:"GPT,LLM,OpenAI,Anthropic,AI,machine learning,deep learning"`
	PublishDelay     time.Duration `env:"PUBLISH_DELAY" envDefault:"5m"`
	Tier1MaxPosts    int           `env:"TIER1_MAX_POSTS" envDefault:"20"`
	Tier1MaxAgeHours int           `env:"TIER1_MAX_AGE_HOURS" envDefault:"24"`
	Tier1Enabled     bool          `env:"TIER1_ENABLED" envDefault:"true"`
	Tier2Enabled     bool          `env:"TIER2_ENABLED" envDefault:"true"`
	Tier3Enabled     bool          `env:"TIER3_ENABLED" envDefault:"true"`
	Tier4Enabled     bool          `env:"TIER4_ENABLED" envDefault:"true"`
}
