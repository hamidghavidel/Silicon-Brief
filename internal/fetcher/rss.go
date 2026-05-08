package fetcher

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

// RSSFetcher fetches stories from RSS/Atom feeds.
type RSSFetcher struct {
	name string
	url  string
}

// NewRSSFetcher creates a new RSS fetcher.
func NewRSSFetcher(name, url string) *RSSFetcher {
	return &RSSFetcher{name: name, url: url}
}

func (r *RSSFetcher) Name() string { return r.name }

// rssFeed represents a minimal RSS structure.
type rssFeed struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

func (r *RSSFetcher) Fetch(ctx context.Context) ([]Story, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url, nil)
	if err != nil {
		return nil, fmt.Errorf("rss %s: create request: %w", r.name, err)
	}
	req.Header.Set("User-Agent", "Silicon-Brief/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rss %s: fetch: %w", r.name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rss %s: unexpected status %d", r.name, resp.StatusCode)
	}

	var feed rssFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("rss %s: decode: %w", r.name, err)
	}

	var stories []Story
	for _, item := range feed.Channel.Items {
		publishedAt := parseRSSDate(item.PubDate)
		stories = append(stories, Story{
			Title:       item.Title,
			URL:         item.Link,
			Source:      r.name,
			SourceType:  "rss",
			PublishedAt: publishedAt,
			Score:       50, // Default baseline for RSS
		})
	}

	return stories, nil
}

func parseRSSDate(s string) time.Time {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.RFC3339,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"2006-01-02T15:04:05Z",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Now()
}
