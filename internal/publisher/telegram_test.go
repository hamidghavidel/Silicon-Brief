package publisher

import (
	"testing"

	"github.com/hamidghavidel/silicon-brief/internal/fetcher"
	"github.com/hamidghavidel/silicon-brief/internal/scorer"
)

func TestFormatMessage(t *testing.T) {
	story := scorer.ScoredStory{
		Story: fetcher.Story{
			Title:  "OpenAI GPT-5",
			URL:    "https://openai.com/blog/gpt-5",
			Source: "OpenAI Blog",
		},
		FinalScore: 85.5,
	}
	msg := formatMessage(story)
	if msg == "" {
		t.Fatal("expected non-empty message")
	}
	if !contains(msg, "OpenAI GPT-5") {
		t.Fatal("expected message to contain title")
	}
	if !contains(msg, "Read more") {
		t.Fatal("expected message to contain 'Read more' link")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
