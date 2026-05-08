package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HNFetcher fetches stories from Hacker News Algolia API.
type HNFetcher struct {
	url string
}

// NewHNFetcher creates a new Hacker News fetcher.
func NewHNFetcher(url string) *HNFetcher {
	return &HNFetcher{url: url}
}

func (h *HNFetcher) Name() string { return "Hacker News" }

// hnResponse represents the Algolia API response.
type hnResponse struct {
	Hits []struct {
		Title     string `json:"title"`
		URL       string `json:"url"`
		Points    int    `json:"points"`
		CreatedAt string `json:"created_at"`
		ObjectID  string `json:"objectID"`
	} `json:"hits"`
}

func (h *HNFetcher) Fetch(ctx context.Context) ([]Story, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.url, nil)
	if err != nil {
		return nil, fmt.Errorf("hn: create request: %w", err)
	}
	req.Header.Set("User-Agent", "Silicon-Brief/1.0")

	resp, err := doWithRetry(ctx, req, 2)
	if err != nil {
		return nil, fmt.Errorf("hn: fetch: %w", err)
	}
	defer resp.Body.Close()

	var result hnResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("hn: decode: %w", err)
	}

	var stories []Story
	for _, hit := range result.Hits {
		publishedAt := parseHNDate(hit.CreatedAt)
		url := hit.URL
		if url == "" {
			url = fmt.Sprintf("https://news.ycombinator.com/item?id=%s", hit.ObjectID)
		}
		stories = append(stories, Story{
			Title:       hit.Title,
			URL:         url,
			Source:      "Hacker News",
			SourceType:  "hackernews",
			PublishedAt: publishedAt,
			Score:       hit.Points,
		})
	}

	return stories, nil
}

func parseHNDate(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Now()
}
