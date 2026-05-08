# Silicon-Brief: Plan & Design Document

## Overview

Silicon-Brief is a zero-cost automated Telegram channel that aggregates, ranks, and publishes AI/ML/LLM tech news from free sources.

## Goals

- Aggregate AI news from RSS, Hacker News, Reddit, and GitHub Trending
- Rank posts using a custom scoring algorithm
- Publish top ~10-20 posts immediately to a Telegram channel
- Run every hour at zero cost

## Architecture

```
GitHub Actions (cron: 0 * * * *)
        │
        ▼
┌─────────────────────┐
│   Go CLI Binary     │
│  (silicon-brief)    │
│                     │
│  1. Fetch Stage     │
│     ├── RSS feeds   │
│     ├── HN API      │
│     ├── Reddit JSON │
│     └── GitHub Trend│
│                     │
│  2. Merge/Dedup     │
│     ├── URL match   │
│     └── Title fuzzy │
│                     │
│  3. Score/Rank      │
│     └── Top N       │
│                     │
│  4. Publish         │
│     └── Telegram API│
└─────────────────────┘
        │
        ▼
   Firebase Firestore (state)
```

## File Structure

```
/
├── .github/workflows/feed.yml
├── cmd/silicon-brief/main.go
├── internal/
│   ├── config/config.go
│   ├── fetcher/
│   │   ├── fetcher.go
│   │   ├── rss.go
│   │   ├── hackernews.go
│   │   ├── reddit.go
│   │   └── github.go
│   ├── scorer/scorer.go
│   ├── dedup/dedup.go
│   ├── store/firestore.go
│   └── publisher/telegram.go
├── go.mod
└── README.md
```

## Scoring Formula

```go
base := 0.0
switch source {
case "hackernews":
    base = float64(hnPoints) * 1.20
case "reddit":
    base = float64(upvotes) * 1.10
case "rss":
    base = 50.0
}

recency := math.Exp(-hoursAgo / 24.0)

keywords := []string{"GPT", "LLM", "OpenAI", "Anthropic"}
keywordBoost := 0.0
for _, kw := range keywords {
    if strings.Contains(strings.ToLower(title), strings.ToLower(kw)) {
        keywordBoost += 10.0
    }
}

score := base + recency + keywordBoost
```

## Sources

| Source | URL/Endpoint | Cost |
|--------|-------------|------|
| RSS Feeds | OpenAI, Google AI, Meta AI, Anthropic, TechCrunch, Ars Technica | Free |
| Hacker News | https://hn.algolia.com/api/v1/search_by_date | Free |
| Reddit | https://www.reddit.com/r/MachineLearning+technology+programming/new.json | Free |
| GitHub Trending | https://github.com/trending (scraped) | Free |

## Deduplication Strategy

1. Exact URL match → merge scores, keep earliest publish time
2. Fuzzy title match (Jaro-Winkler distance > 0.85) → merge

## State Management

Firebase Firestore (Spark plan - free tier):
- Collection: `published_posts`
- Document ID: URL hash or normalized URL
- Fields: `url`, `title`, `source`, `score`, `publishedAt`, `createdAt`

## GitHub Actions Workflow

```yaml
name: Hourly Feed
on:
  schedule:
    - cron: '0 * * * *'
jobs:
  feed:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go run ./cmd/silicon-brief
        env:
          TELEGRAM_BOT_TOKEN: ${{ secrets.TELEGRAM_BOT_TOKEN }}
          TELEGRAM_CHANNEL_ID: ${{ secrets.TELEGRAM_CHANNEL_ID }}
          FIREBASE_PROJECT_ID: ${{ secrets.FIREBASE_PROJECT_ID }}
          FIREBASE_SERVICE_ACCOUNT_JSON: ${{ secrets.FIREBASE_SERVICE_ACCOUNT_JSON }}
```

## Required Secrets

- `TELEGRAM_BOT_TOKEN` - From @BotFather
- `TELEGRAM_CHANNEL_ID` - Target channel ID
- `FIREBASE_PROJECT_ID` - Firebase project identifier
- `FIREBASE_SERVICE_ACCOUNT_JSON` - Base64-encoded service account key

## Tech Stack

- **Language**: Go (matches existing project config)
- **Concurrency**: Goroutines for parallel fetching
- **HTTP Client**: Standard library `net/http`
- **RSS Parsing**: `github.com/mmcdole/gofeed`
- **Fuzzy Matching**: `github.com/xrash/smetrics`
- **Firebase**: `firebase.google.com/go`
- **Telegram**: `github.com/go-telegram-bot-api/telegram-bot-api`

## Verification Plan

1. Run locally with test secrets
2. Verify all fetchers return data
3. Verify deduplication merges correctly
4. Verify scoring produces expected rankings
5. Verify Telegram message format
6. Deploy and monitor GitHub Actions logs
7. Confirm channel receives posts
