package fetcher

import (
	"context"
	"time"
)

// Story represents a single news item.
type Story struct {
	Title       string
	URL         string
	Source      string
	SourceType  string
	PublishedAt time.Time
	Score       int // Raw score: upvotes, points, etc.
}

// Fetcher is the interface for all news sources.
type Fetcher interface {
	Fetch(ctx context.Context) ([]Story, error)
	Name() string
}

// FetchAll fetches from all provided fetchers concurrently.
func FetchAll(ctx context.Context, fetchers []Fetcher) ([]Story, error) {
	type result struct {
		stories []Story
		err     error
	}

	results := make(chan result, len(fetchers))
	for _, f := range fetchers {
		go func(fetcher Fetcher) {
			stories, err := fetcher.Fetch(ctx)
			results <- result{stories: stories, err: err}
		}(f)
	}

	var allStories []Story
	for i := 0; i < len(fetchers); i++ {
		res := <-results
		if res.err != nil {
			// Log error but continue with other sources
			continue
		}
		allStories = append(allStories, res.stories...)
	}

	return allStories, nil
}
