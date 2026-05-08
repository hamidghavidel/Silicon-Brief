package scorer

import (
	"math"
	"strings"
	"time"

	"github.com/hamidghavidel/silicon-brief/internal/fetcher"
)

// ScoredStory wraps a Story with its calculated score.
type ScoredStory struct {
	fetcher.Story
	FinalScore float64
}

// Score calculates the ranking score for a story.
func Score(story fetcher.Story, keywords []string) ScoredStory {
	base := 0.0
	switch story.SourceType {
	case "hackernews":
		base = float64(story.Score) * 1.20
	case "reddit":
		base = float64(story.Score) * 1.10
	case "rss":
		base = 50.0
	case "github":
		base = float64(story.Score) * 1.0
	}

	// Recency boost: score += e^(-hours/24)
	hoursAgo := time.Since(story.PublishedAt).Hours()
	recency := math.Exp(-hoursAgo / 24.0)

	// Keyword boost
	keywordBoost := 0.0
	lowerTitle := strings.ToLower(story.Title)
	for _, kw := range keywords {
		if strings.Contains(lowerTitle, strings.ToLower(kw)) {
			keywordBoost += 10.0
		}
	}

	finalScore := base + recency + keywordBoost

	return ScoredStory{
		Story:      story,
		FinalScore: finalScore,
	}
}

// Rank sorts stories by final score descending.
func Rank(stories []ScoredStory) []ScoredStory {
	sorted := make([]ScoredStory, len(stories))
	copy(sorted, stories)

	// Simple bubble sort for clarity (small N)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].FinalScore > sorted[i].FinalScore {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// TopN returns the top N stories.
func TopN(stories []ScoredStory, n int) []ScoredStory {
	if n > len(stories) {
		n = len(stories)
	}
	return stories[:n]
}
