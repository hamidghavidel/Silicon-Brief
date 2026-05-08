package config

import "time"

// Feed holds feed-related config.
type Feed struct {
	MaxPosts         int           `env:"MAX_POSTS" envDefault:"15"`
	Keywords         []string      `env:"KEYWORDS" envSeparator:"," envDefault:"GPT,LLM,OpenAI,Anthropic,AI,machine learning,deep learning"`
	PublishDelay     time.Duration `env:"PUBLISH_DELAY" envDefault:"5m"`
	Tier1MaxPosts    int           `env:"TIER1_MAX_POSTS" envDefault:"20"`
	Tier1MaxAgeHours int           `env:"TIER1_MAX_AGE_HOURS" envDefault:"24"`
}
