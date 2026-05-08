package dedup

import (
	"testing"
	"time"

	"github.com/hamidghavidel/silicon-brief/internal/fetcher"
)

func TestDedup_UniqueStories(t *testing.T) {
	d := NewDeduplicator()
	stories := []fetcher.Story{
		{Title: "A", URL: "http://a", Source: "S1", PublishedAt: time.Now(), Score: 10},
		{Title: "B", URL: "http://b", Source: "S2", PublishedAt: time.Now(), Score: 20},
	}
	got := d.Dedup(stories)
	if len(got) != 2 {
		t.Fatalf("expected 2 stories, got %d", len(got))
	}
}

func TestDedup_DuplicateURL(t *testing.T) {
	d := NewDeduplicator()
	stories := []fetcher.Story{
		{Title: "A", URL: "http://a", Source: "S1", PublishedAt: time.Now(), Score: 10},
		{Title: "A", URL: "http://a", Source: "S2", PublishedAt: time.Now(), Score: 20},
	}
	got := d.Dedup(stories)
	if len(got) != 1 {
		t.Fatalf("expected 1 story, got %d", len(got))
	}
}

func TestDedup_FuzzyTitleMerge(t *testing.T) {
	d := NewDeduplicator()
	stories := []fetcher.Story{
		{Title: "OpenAI releases GPT-5", URL: "http://a", Source: "S1", PublishedAt: time.Now(), Score: 10},
		{Title: "OpenAI releases GPT-5", URL: "http://b", Source: "S2", PublishedAt: time.Now(), Score: 20},
	}
	got := d.Dedup(stories)
	if len(got) != 1 {
		t.Fatalf("expected 1 story after fuzzy merge, got %d", len(got))
	}
	if got[0].Score != 20 {
		t.Fatalf("expected merged score 20, got %d", got[0].Score)
	}
}
