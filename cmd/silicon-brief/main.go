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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
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
	if !cfg.Telegram.PublishEnabled {
		log.Println("Telegram publishing is disabled")
	}

	// Build source tier lookup
	tierBySource := make(map[string]int)
	for _, src := range cfg.Sources {
		tierBySource[src.Name] = src.Tier
	}

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
		case "claudeblog":
			fetchers = append(fetchers, fetcher.NewClaudeBlogFetcher(src.URL))
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

	// Group stories by tier
	tierStories := map[int][]fetcher.Story{
		1: {},
		2: {},
		3: {},
	}
	for _, story := range uniqueStories {
		tier := tierBySource[story.Source]
		if tier == 0 {
			tier = 2 // default fallback
		}
		tierStories[tier] = append(tierStories[tier], story)
	}

	publishedCount := 0

	// Tier 1: Publish ALL new posts (no ranking)
	if len(tierStories[1]) > 0 {
		log.Printf("Processing Tier 1: %d priority stories", len(tierStories[1]))
		for _, story := range tierStories[1] {
			if done, err := publishIfNew(ctx, pub, st, story, cfg.Feed.Keywords, cfg.Telegram.PublishEnabled); err != nil {
				log.Printf("Error publishing tier 1 story: %v", err)
				continue
			} else if done {
				publishedCount++
				if err := waitDelay(ctx, cfg.Feed.PublishDelay); err != nil {
					return err
				}
			}
		}
	}

	// Tier 2: Score, rank, publish top N
	if len(tierStories[2]) > 0 {
		log.Printf("Processing Tier 2: %d stories", len(tierStories[2]))
		tier2Published := publishTier(ctx, pub, st, tierStories[2], cfg.Feed.Keywords, cfg.Feed.MaxPosts, cfg.Feed.PublishDelay, cfg.Telegram.PublishEnabled)
		publishedCount += tier2Published
	}

	// Tier 3: Score, rank, publish top N
	if len(tierStories[3]) > 0 {
		log.Printf("Processing Tier 3: %d stories", len(tierStories[3]))
		tier3Published := publishTier(ctx, pub, st, tierStories[3], cfg.Feed.Keywords, cfg.Feed.MaxPosts, cfg.Feed.PublishDelay, cfg.Telegram.PublishEnabled)
		publishedCount += tier3Published
	}

	log.Printf("Done. Published %d new stories.", publishedCount)
	return nil
}

// publishTier scores, ranks, and publishes the top N new stories from a tier.
func publishTier(ctx context.Context, pub *publisher.Publisher, st *store.Store, stories []fetcher.Story, keywords []string, maxPosts int, delay time.Duration, publishEnabled bool) int {
	var scored []scorer.ScoredStory
	for _, story := range stories {
		scored = append(scored, scorer.Score(story, keywords))
	}
	ranked := scorer.Rank(scored)
	topStories := scorer.TopN(ranked, maxPosts)

	count := 0
	for _, story := range topStories {
		if done, err := publishIfNew(ctx, pub, st, story.Story, keywords, publishEnabled); err != nil {
			log.Printf("Error publishing tier story: %v", err)
			continue
		} else if done {
			count++
			if err := waitDelay(ctx, delay); err != nil {
				return count
			}
		}
	}
	return count
}

// publishIfNew checks if a story is already published, and if not, publishes and marks it.
func publishIfNew(ctx context.Context, pub *publisher.Publisher, st *store.Store, story fetcher.Story, keywords []string, publishEnabled bool) (bool, error) {
	isPublished, err := st.IsPublished(ctx, story)
	if err != nil {
		return false, fmt.Errorf("check published status for %s: %w", story.Title, err)
	}
	if isPublished {
		return false, nil
	}

	scored := scorer.Score(story, keywords)

	if publishEnabled {
		if err := pub.Publish(ctx, scored); err != nil {
			return false, fmt.Errorf("publish %s: %w", story.Title, err)
		}
	} else {
		log.Printf("[DRY RUN] Would publish: %s", story.Title)
	}

	if err := st.MarkPublished(ctx, story, scored.FinalScore); err != nil {
		return false, fmt.Errorf("mark published %s: %w", story.Title, err)
	}

	log.Printf("Published: %s", story.Title)
	return true, nil
}

// waitDelay sleeps for the given duration or returns an error if context is cancelled.
func waitDelay(ctx context.Context, d time.Duration) error {
	log.Printf("Waiting %s before next publish...", d)
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		log.Println("Context cancelled, stopping publish loop")
		return ctx.Err()
	}
}
