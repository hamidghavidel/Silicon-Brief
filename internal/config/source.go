package config

// Source represents a news source configuration.
type Source struct {
	Name   string  `json:"name"`
	URL    string  `json:"url"`
	Weight float64 `json:"weight"`
	Type   string  `json:"type"` // "rss", "hackernews", "reddit", "github"
}
