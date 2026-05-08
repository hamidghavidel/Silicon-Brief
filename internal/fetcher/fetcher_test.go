package fetcher

import (
	"context"
	"testing"
	"time"
)

type mockFetcher struct {
	name    string
	stories []Story
	err     error
}

func (m *mockFetcher) Name() string { return m.name }

func (m *mockFetcher) Fetch(ctx context.Context) ([]Story, error) {
	return m.stories, m.err
}

func TestFetchAll_Concurrent(t *testing.T) {
	ctx := context.Background()
	fetchers := []Fetcher{
		&mockFetcher{name: "A", stories: []Story{{Title: "A1", URL: "http://a1", Source: "A", SourceType: "rss", PublishedAt: time.Now(), Score: 10}}},
		&mockFetcher{name: "B", stories: []Story{{Title: "B1", URL: "http://b1", Source: "B", SourceType: "rss", PublishedAt: time.Now(), Score: 20}}},
	}

	stories, err := FetchAll(ctx, fetchers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stories) != 2 {
		t.Fatalf("expected 2 stories, got %d", len(stories))
	}
}

func TestFetchAll_PartialFailure(t *testing.T) {
	ctx := context.Background()
	fetchers := []Fetcher{
		&mockFetcher{name: "A", stories: []Story{{Title: "A1", URL: "http://a1", Source: "A", SourceType: "rss", PublishedAt: time.Now(), Score: 10}}},
		&mockFetcher{name: "B", err: context.DeadlineExceeded},
	}

	stories, err := FetchAll(ctx, fetchers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stories) != 1 {
		t.Fatalf("expected 1 story, got %d", len(stories))
	}
}
