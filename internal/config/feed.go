package config

import "time"

// Feed holds feed-related config.
type Feed struct {
	MaxPosts     int           `env:"MAX_POSTS" envDefault:"15"`
	Keywords     []string      `env:"KEYWORDS" envSeparator:"," envDefault:"GPT,LLM,OpenAI,Anthropic,AI,machine learning,deep learning"`
	PublishDelay time.Duration `env:"PUBLISH_DELAY" envDefault:"5m"`
}
