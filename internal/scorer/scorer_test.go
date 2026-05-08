package scorer

import (
	"testing"
	"time"

	"github.com/hamidghavidel/silicon-brief/internal/fetcher"
)

func TestScore_RSSBase(t *testing.T) {
	story := fetcher.Story{
		Title:       "Test",
		SourceType:  "rss",
		PublishedAt: time.Now(),
		Score:       0,
	}
	scored := Score(story, []string{})
	if scored.FinalScore < 50 {
		t.Fatalf("expected RSS base >= 50, got %.2f", scored.FinalScore)
	}
}

func TestScore_KeywordBoost(t *testing.T) {
	story := fetcher.Story{
		Title:       "OpenAI GPT-5 LLM",
		SourceType:  "rss",
		PublishedAt: time.Now(),
		Score:       0,
	}
	scored := Score(story, []string{"GPT", "LLM", "OpenAI"})
	expectedBase := 50.0 + 30.0 // 3 keywords * 10
	if scored.FinalScore < expectedBase {
		t.Fatalf("expected score >= %.2f, got %.2f", expectedBase, scored.FinalScore)
	}
}

func TestRank(t *testing.T) {
	stories := []ScoredStory{
		{Story: fetcher.Story{Title: "A"}, FinalScore: 10},
		{Story: fetcher.Story{Title: "B"}, FinalScore: 30},
		{Story: fetcher.Story{Title: "C"}, FinalScore: 20},
	}
	ranked := Rank(stories)
	if ranked[0].Title != "B" {
		t.Fatalf("expected top story B, got %s", ranked[0].Title)
	}
}

func TestTopN(t *testing.T) {
	stories := []ScoredStory{
		{Story: fetcher.Story{Title: "A"}, FinalScore: 10},
		{Story: fetcher.Story{Title: "B"}, FinalScore: 20},
		{Story: fetcher.Story{Title: "C"}, FinalScore: 30},
	}
	ranked := Rank(stories)
	top := TopN(ranked, 2)
	if len(top) != 2 {
		t.Fatalf("expected 2 stories, got %d", len(top))
	}
	if top[0].Title != "C" {
		t.Fatalf("expected top story C, got %s", top[0].Title)
	}
}
