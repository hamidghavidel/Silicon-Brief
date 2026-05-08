package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hamidghavidel/silicon-brief/internal/config"
	"github.com/hamidghavidel/silicon-brief/internal/dedup"
	"github.com/hamidghavidel/silicon-brief/internal/fetcher"
	"github.com/hamidghavidel/silicon-brief/internal/publisher"
	"github.com/hamidghavidel/silicon-brief/internal/scorer"
	"github.com/hamidghavidel/silicon-brief/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Load configuration
	cfg, err := config.Configure()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize store
	st, err := store.New("silicon-brief.db")
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}
	defer st.Close()

	// Initialize publisher
	pub := publisher.New(cfg.Telegram.BotToken, cfg.Telegram.ChannelID)

	// Build fetchers from config
	var fetchers []fetcher.Fetcher
	for _, src := range cfg.Sources {
		switch src.Type {
		case "rss":
			fetchers = append(fetchers, fetcher.NewRSSFetcher(src.Name, src.URL))
		case "hackernews":
			fetchers = append(fetchers, fetcher.NewHNFetcher(src.URL))
		case "reddit":
			fetchers = append(fetchers, fetcher.NewRedditFetcher(src.URL))
		case "github":
			fetchers = append(fetchers, fetcher.NewGitHubFetcher(src.URL))
		}
	}

	// Fetch all stories
	log.Println("Fetching stories...")
	allStories, err := fetcher.FetchAll(ctx, fetchers)
	if err != nil {
		return fmt.Errorf("fetch all: %w", err)
	}
	log.Printf("Fetched %d stories", len(allStories))

	// Deduplicate
	d := dedup.NewDeduplicator()
	uniqueStories := d.Dedup(allStories)
	log.Printf("After dedup: %d stories", len(uniqueStories))

	// Score and rank
	var scored []scorer.ScoredStory
	for _, story := range uniqueStories {
		scored = append(scored, scorer.Score(story, cfg.Feed.Keywords))
	}
	ranked := scorer.Rank(scored)
	topStories := scorer.TopN(ranked, cfg.Feed.MaxPosts)
	log.Printf("Top %d stories selected", len(topStories))

	// Publish new stories
	publishedCount := 0
	for _, story := range topStories {
		isPublished, err := st.IsPublished(ctx, story.Story)
		if err != nil {
			log.Printf("Error checking published status for %s: %v", story.Title, err)
			continue
		}
		if isPublished {
			continue
		}

		if err := pub.Publish(ctx, story); err != nil {
			log.Printf("Error publishing %s: %v", story.Title, err)
			continue
		}

		if err := st.MarkPublished(ctx, story.Story, story.FinalScore); err != nil {
			log.Printf("Error marking published %s: %v", story.Title, err)
			continue
		}

		publishedCount++
		log.Printf("Published: %s", story.Title)
	}

	log.Printf("Done. Published %d new stories.", publishedCount)
	return nil
}
