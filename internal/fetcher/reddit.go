package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RedditFetcher fetches posts from Reddit JSON endpoints.
type RedditFetcher struct {
	url string
}

// NewRedditFetcher creates a new Reddit fetcher.
func NewRedditFetcher(url string) *RedditFetcher {
	return &RedditFetcher{url: url}
}

func (r *RedditFetcher) Name() string { return "Reddit" }

// redditResponse represents the Reddit API response.
type redditResponse struct {
	Data struct {
		Children []struct {
			Data struct {
				Title     string  `json:"title"`
				URL       string  `json:"url"`
				Score     int     `json:"score"`
				CreatedAt float64 `json:"created_utc"`
				Permalink string  `json:"permalink"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func (r *RedditFetcher) Fetch(ctx context.Context) ([]Story, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url, nil)
	if err != nil {
		return nil, fmt.Errorf("reddit: create request: %w", err)
	}
	req.Header.Set("User-Agent", "Silicon-Brief/1.0 (by /u/siliconbrief)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reddit: fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reddit: unexpected status %d", resp.StatusCode)
	}

	var result redditResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("reddit: decode: %w", err)
	}

	var stories []Story
	for _, child := range result.Data.Children {
		post := child.Data
		url := post.URL
		if url == "" || url == "self" {
			url = fmt.Sprintf("https://reddit.com%s", post.Permalink)
		}
		stories = append(stories, Story{
			Title:       post.Title,
			URL:         url,
			Source:      "Reddit",
			SourceType:  "reddit",
			PublishedAt: time.Unix(int64(post.CreatedAt), 0),
			Score:       post.Score,
		})
	}

	return stories, nil
}
