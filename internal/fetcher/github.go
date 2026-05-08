package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// GitHubFetcher scrapes trending repositories from GitHub.
type GitHubFetcher struct {
	url string
}

// NewGitHubFetcher creates a new GitHub trending fetcher.
func NewGitHubFetcher(url string) *GitHubFetcher {
	return &GitHubFetcher{url: url}
}

func (g *GitHubFetcher) Name() string { return "GitHub Trending" }

func (g *GitHubFetcher) Fetch(ctx context.Context) ([]Story, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.url, nil)
	if err != nil {
		return nil, fmt.Errorf("github: create request: %w", err)
	}
	req.Header.Set("User-Agent", "Silicon-Brief/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github: unexpected status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("github: parse html: %w", err)
	}

	var stories []Story
	aiKeywords := []string{"ai", "ml", "llm", "gpt", "neural", "deep-learning", "machine-learning", "openai", "anthropic", "model", "transformer", "diffusion"}

	doc.Find("article.Box-row").Each(func(i int, s *goquery.Selection) {
		linkElem := s.Find("h2 a")
		repoPath := strings.TrimSpace(linkElem.Text())
		repoPath = strings.Join(strings.Fields(repoPath), "")
		if repoPath == "" {
			return
		}

		description := strings.TrimSpace(s.Find("p.col-9").Text())
		title := fmt.Sprintf("%s: %s", repoPath, description)
		url := fmt.Sprintf("https://github.com%s", linkElem.AttrOr("href", ""))

		// Filter by AI keywords
		lowerTitle := strings.ToLower(title)
		lowerDesc := strings.ToLower(description)
		isAI := false
		for _, kw := range aiKeywords {
			if strings.Contains(lowerTitle, kw) || strings.Contains(lowerDesc, kw) {
				isAI = true
				break
			}
		}
		if !isAI {
			return
		}

		// Extract star count if available
		starsText := s.Find("a[href$=\"stargazers\"]").Text()
		stars := 0
		fmt.Sscanf(starsText, "%d", &stars)

		stories = append(stories, Story{
			Title:       title,
			URL:         url,
			Source:      "GitHub Trending",
			SourceType:  "github",
			PublishedAt: time.Now(),
			Score:       stars,
		})
	})

	return stories, nil
}
