package config

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config is the top-level configuration.
type Config struct {
	App      App      `envPrefix:"APP_"`
	Telegram Telegram `envPrefix:"TELEGRAM_"`
	Feed     Feed     `envPrefix:"FEED_"`
	Sources  []Source `envPrefix:"FEED_SOURCES"`
}

func Configure() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		slog.With("err", err.Error()).Error("reading .env file error")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parsing configuration error: %w", err)
	}

	cfg.Sources = defaultSources()
	return cfg, nil
}

func defaultSources() []Source {
	return []Source{
		// Tier 1: Priority sources — all new posts published
		{Name: "OpenAI Blog", URL: "https://openai.com/blog/rss.xml", Weight: 1.0, Type: "rss", Tier: 1},
		{Name: "OpenAI News", URL: "https://openai.com/news/rss.xml", Weight: 1.0, Type: "rss", Tier: 1},
		{Name: "Google AI Blog", URL: "https://blog.google/technology/ai/rss/", Weight: 1.0, Type: "rss", Tier: 1},
		{Name: "Claude Blog", URL: "https://claude.com/blog", Weight: 1.0, Type: "claudeblog", Tier: 1},

		// Tier 2: Normal ranked sources
		{Name: "TechCrunch AI", URL: "https://techcrunch.com/category/artificial-intelligence/feed/", Weight: 1.0, Type: "rss", Tier: 2},
		{Name: "Hacker News", URL: "https://hn.algolia.com/api/v1/search_by_date?tags=story&query=AI|LLM|machine%20learning", Weight: 1.2, Type: "hackernews", Tier: 2},
		{Name: "Reddit", URL: "https://www.reddit.com/r/MachineLearning+technology+programming/new.json?limit=25", Weight: 1.1, Type: "reddit", Tier: 2},

		// Tier 3: Low priority ranked sources
		{Name: "GitHub Trending", URL: "https://github.com/trending?spoken_language_code=en", Weight: 1.0, Type: "github", Tier: 3},
	}
}
