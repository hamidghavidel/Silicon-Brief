package dedup

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/hamidghavidel/silicon-brief/internal/fetcher"
	"github.com/xrash/smetrics"
)

// Deduplicator handles merging duplicate stories.
type Deduplicator struct {
	seenURLs map[string]bool
}

// NewDeduplicator creates a new deduplicator.
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		seenURLs: make(map[string]bool),
	}
}

// Dedup merges duplicate stories by URL and fuzzy title matching.
func (d *Deduplicator) Dedup(stories []fetcher.Story) []fetcher.Story {
	var unique []fetcher.Story

	for _, story := range stories {
		// Normalize URL
		urlHash := hashURL(story.URL)
		if d.seenURLs[urlHash] {
			continue
		}

		// Check fuzzy title match against existing unique stories
		merged := false
		for i := range unique {
			if isSimilarTitle(unique[i].Title, story.Title) {
				// Merge: keep higher score, earlier publish time
				if story.Score > unique[i].Score {
					unique[i].Score = story.Score
				}
				merged = true
				break
			}
		}

		if !merged {
			d.seenURLs[urlHash] = true
			unique = append(unique, story)
		}
	}

	return unique
}

func hashURL(url string) string {
	h := sha256.New()
	h.Write([]byte(strings.ToLower(url)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func isSimilarTitle(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == b {
		return true
	}
	distance := smetrics.JaroWinkler(a, b, 0.7, 4)
	return distance > 0.85
}
