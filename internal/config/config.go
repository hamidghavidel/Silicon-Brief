package config

import (
	"fmt"
	"os"
	"time"
)

// Source represents a news source configuration.
type Source struct {
	Name   string
	URL    string
	Weight float64
	Type   string // "rss", "hackernews", "reddit", "github"
}

// Config holds all application configuration.
type Config struct {
	TelegramBotToken       string
	TelegramChannelID      string
	FirebaseProjectID      string
	FirebaseServiceAccount string
	Sources                []Source
	Keywords               []string
	MaxPosts               int
	FetchTimeout           time.Duration
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	channelID := os.Getenv("TELEGRAM_CHANNEL_ID")
	if channelID == "" {
		return nil, fmt.Errorf("TELEGRAM_CHANNEL_ID is required")
	}

	firebaseProjectID := os.Getenv("FIREBASE_PROJECT_ID")
	if firebaseProjectID == "" {
		return nil, fmt.Errorf("FIREBASE_PROJECT_ID is required")
	}

	firebaseSA := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON")
	if firebaseSA == "" {
		return nil, fmt.Errorf("FIREBASE_SERVICE_ACCOUNT_JSON is required")
	}

	return &Config{
		TelegramBotToken:       botToken,
		TelegramChannelID:      channelID,
		FirebaseProjectID:      firebaseProjectID,
		FirebaseServiceAccount: firebaseSA,
		Sources:                defaultSources(),
		Keywords:               []string{"GPT", "LLM", "OpenAI", "Anthropic", "AI", "machine learning", "deep learning"},
		MaxPosts:               15,
		FetchTimeout:           30 * time.Second,
	}, nil
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
