package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ClaudeBlogFetcher scrapes claude.com/blog for articles.
type ClaudeBlogFetcher struct {
	url string
}

// NewClaudeBlogFetcher creates a new Claude blog scraper.
func NewClaudeBlogFetcher(url string) *ClaudeBlogFetcher {
	return &ClaudeBlogFetcher{url: url}
}

func (c *ClaudeBlogFetcher) Name() string { return "Claude Blog" }

// articleLink matches href="/blog/slug" patterns.
var articleLink = regexp.MustCompile(`href="(/blog/[^"]+)"`)

func (c *ClaudeBlogFetcher) Fetch(ctx context.Context) ([]Story, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, fmt.Errorf("claude blog: create request: %w", err)
	}
	req.Header.Set("User-Agent", "Silicon-Brief/1.0")

	resp, err := doWithRetry(ctx, req, 2)
	if err != nil {
		return nil, fmt.Errorf("claude blog: fetch: %w", err)
	}
	defer resp.Body.Close()

	// Read full response (limit to 2MB to avoid memory issues).
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("claude blog: read body: %w", err)
	}
	content := string(body)

	// Extract unique article paths.
	matches := articleLink.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)
	var stories []Story
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		path := m[1]
		if seen[path] {
			continue
		}
		seen[path] = true

		// Skip non-article paths.
		if strings.Contains(path, "/blog/") && !strings.HasSuffix(path, "/blog") {
			stories = append(stories, Story{
				Title:       slugToTitle(path),
				URL:         "https://claude.com" + path,
				Source:      c.Name(),
				SourceType:  "scraper",
				PublishedAt: time.Now(),
				Score:       50,
			})
		}
	}

	return stories, nil
}

// slugToTitle converts a URL slug like "new-in-claude" to a title.
func slugToTitle(path string) string {
	// Extract slug from /blog/slug.
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return path
	}
	slug := parts[len(parts)-1]
	slug = strings.ReplaceAll(slug, "-", " ")
	// Capitalize first letter of each word.
	words := strings.Fields(slug)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
