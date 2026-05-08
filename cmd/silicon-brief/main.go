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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	cfg, err := config.Configure()
	if err != nil {
		log.Fatal(err)
	}

	st, err := store.New("silicon-brief.db")
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}
	defer st.Close()

	pub := publisher.New(cfg.Telegram.BotToken, cfg.Telegram.ChannelID)
	if !cfg.Telegram.PublishEnabled {
		log.Println("Telegram publishing is disabled")
	}

	// Build tier-specific fetchers
	tierFetchers := buildTierFetchers(cfg.Sources)

	// Launch each tier concurrently
	tierResults := make(chan tierResult, 3)
	for tier, fetchers := range tierFetchers {
		if len(fetchers) == 0 {
			continue
		}
		go func(t int, f []fetcher.Fetcher) {
			count := processTier(ctx, t, f, pub, st, cfg)
			tierResults <- tierResult{tier: t, count: count}
		}(tier, fetchers)
	}

	// Collect results
	totalPublished := 0
	tiersProcessed := 0
	for range tierFetchers {
		res := <-tierResults
		log.Printf("Tier %d finished. Published %d stories.", res.tier, res.count)
		totalPublished += res.count
		tiersProcessed++
	}

	log.Printf("All tiers complete. Total published: %d new stories.", totalPublished)
	return nil
}

type tierResult struct {
	tier  int
	count int
}

// buildTierFetchers groups fetchers by their source tier.
func buildTierFetchers(sources []config.Source) map[int][]fetcher.Fetcher {
	m := map[int][]fetcher.Fetcher{}
	for _, src := range sources {
		var f fetcher.Fetcher
		switch src.Type {
		case "rss":
			f = fetcher.NewRSSFetcher(src.Name, src.URL)
		case "hackernews":
			f = fetcher.NewHNFetcher(src.URL)
		case "reddit":
			f = fetcher.NewRedditFetcher(src.URL)
		case "github":
			f = fetcher.NewGitHubFetcher(src.URL)
		case "claudeblog":
			f = fetcher.NewClaudeBlogFetcher(src.URL)
		}
		if f != nil {
			tier := src.Tier
			if tier == 0 {
				tier = 2
			}
			m[tier] = append(m[tier], f)
		}
	}
	return m
}

// processTier fetches, deduplicates, scores, and publishes stories for a single tier.
func processTier(ctx context.Context, tier int, fetchers []fetcher.Fetcher, pub *publisher.Publisher, st *store.Store, cfg *config.Config) int {
	log.Printf("[Tier %d] Fetching...", tier)
	allStories, err := fetcher.FetchAll(ctx, fetchers)
	if err != nil {
		log.Printf("[Tier %d] Fetch error: %v", tier, err)
		return 0
	}
	log.Printf("[Tier %d] Fetched %d stories", tier, len(allStories))

	d := dedup.NewDeduplicator()
	uniqueStories := d.Dedup(allStories)
	log.Printf("[Tier %d] After dedup: %d stories", tier, len(uniqueStories))

	var storiesToPublish []fetcher.Story
	if tier == 1 {
		// Tier 1: filter by age, cap max posts, no ranking
		cutoff := time.Now().Add(-time.Duration(cfg.Feed.Tier1MaxAgeHours) * time.Hour)
		for _, story := range uniqueStories {
			if story.PublishedAt.After(cutoff) {
				storiesToPublish = append(storiesToPublish, story)
			}
		}
		if len(storiesToPublish) > cfg.Feed.Tier1MaxPosts {
			storiesToPublish = storiesToPublish[:cfg.Feed.Tier1MaxPosts]
		}
		log.Printf("[Tier %d] Filtered to %d stories (max %dh, cap %d)", tier, len(storiesToPublish), cfg.Feed.Tier1MaxAgeHours, cfg.Feed.Tier1MaxPosts)
	} else {
		// Tier 2/3: score, rank, take top N
		var scored []scorer.ScoredStory
		for _, story := range uniqueStories {
			scored = append(scored, scorer.Score(story, cfg.Feed.Keywords))
		}
		ranked := scorer.Rank(scored)
		topStories := scorer.TopN(ranked, cfg.Feed.MaxPosts)
		for _, s := range topStories {
			storiesToPublish = append(storiesToPublish, s.Story)
		}
		log.Printf("[Tier %d] Top %d stories selected", tier, len(storiesToPublish))
	}

	count := 0
	for _, story := range storiesToPublish {
		if done, err := publishIfNew(ctx, pub, st, story, cfg.Feed.Keywords, cfg.Telegram.PublishEnabled); err != nil {
			log.Printf("[Tier %d] Error publishing: %v", tier, err)
			continue
		} else if done {
			count++
			if err := waitDelay(ctx, cfg.Feed.PublishDelay); err != nil {
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
