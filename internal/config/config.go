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
		{Name: "OpenAI Blog", URL: "https://openai.com/blog/rss.xml", Weight: 1.0, Type: "rss"},
		{Name: "Google AI Blog", URL: "https://ai.googleblog.com/feeds/posts/default", Weight: 1.0, Type: "rss"},
		{Name: "Anthropic Blog", URL: "https://www.anthropic.com/rss.xml", Weight: 1.0, Type: "rss"},
		{Name: "TechCrunch AI", URL: "https://techcrunch.com/category/artificial-intelligence/feed/", Weight: 1.0, Type: "rss"},
		{Name: "Hacker News", URL: "https://hn.algolia.com/api/v1/search_by_date?tags=story&query=AI|LLM|machine%20learning", Weight: 1.2, Type: "hackernews"},
		{Name: "Reddit", URL: "https://www.reddit.com/r/MachineLearning+technology+programming/new.json?limit=25", Weight: 1.1, Type: "reddit"},
		{Name: "GitHub Trending", URL: "https://github.com/trending?spoken_language_code=en", Weight: 1.0, Type: "github"},
	}
}
