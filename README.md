# Silicon-Brief

A zero-cost automated Telegram channel that aggregates, ranks, and publishes AI/ML/LLM tech news from free sources.

## Architecture

```
GitHub Actions (cron: 0 * * * *)
        |
        v
+----------------------------+
| Go CLI (silicon-brief)     |
|  1. Fetch (RSS/HN/Reddit)  |
|  2. Deduplicate            |
|  3. Score & Rank           |
|  4. Publish to Telegram    |
+----------------------------+
        |
        v
  Firebase Firestore (state)
```

## Setup

### 1. Fork & Clone

```bash
git clone https://github.com/hamidghavidel/silicon-brief.git
cd silicon-brief
```

### 2. Configure Secrets

Go to **Settings > Secrets and variables > Actions** in your GitHub repository and add:

| Secret | Description |
|--------|-------------|
| `TELEGRAM_BOT_TOKEN` | From [@BotFather](https://t.me/botfather) |
| `TELEGRAM_CHANNEL_ID` | Target channel ID (e.g., `-1001234567890` or `@channelname`) |
| `FIREBASE_PROJECT_ID` | Firebase project identifier |
| `FIREBASE_SERVICE_ACCOUNT_JSON` | Base64-encoded service account key |

### 3. Enable GitHub Actions

The workflow `.github/workflows/feed.yml` runs every hour automatically. You can also trigger it manually via **Actions > Hourly Feed > Run workflow**.

## Local Development

```bash
# Set environment variables
export TELEGRAM_BOT_TOKEN="your-token"
export TELEGRAM_CHANNEL_ID="your-channel-id"
export FIREBASE_PROJECT_ID="your-project"
export FIREBASE_SERVICE_ACCOUNT_JSON="$(cat service-account.json)"

# Run
go run ./cmd/silicon-brief

# Test
go test -race -cover ./...
```

## Project Structure

```
.
├── cmd/silicon-brief/main.go      # Entry point
├── internal/
│   ├── config/config.go           # Environment-based config
│   ├── fetcher/                   # News source fetchers
│   │   ├── fetcher.go             # Interface & concurrent fetching
│   │   ├── rss.go                 # RSS/Atom feeds
│   │   ├── hackernews.go          # Hacker News Algolia API
│   │   ├── reddit.go              # Reddit JSON API
│   │   ├── github.go              # GitHub Trending scraper
│   │   └── retry.go               # HTTP retry with backoff
│   ├── scorer/scorer.go           # Ranking algorithm
│   ├── dedup/dedup.go             # URL & fuzzy title deduplication
│   ├── store/firestore.go         # Published post tracking
│   └── publisher/telegram.go      # Telegram channel publisher
├── .github/workflows/feed.yml     # GitHub Actions cron job
└── docs/PLAN.md                   # Design document
```

## Scoring

Stories are ranked by:
- **Base score**: source weight (HN points x 1.2, Reddit upvotes x 1.1, RSS = 50)
- **Recency boost**: `e^(-hours/24)`
- **Keyword boost**: +10 per matched keyword (GPT, LLM, OpenAI, Anthropic, AI, etc.)

## Sources

- OpenAI Blog (RSS)
- Google AI Blog (RSS)
- Anthropic Blog (RSS)
- TechCrunch AI (RSS)
- Hacker News (Algolia API)
- Reddit (r/MachineLearning + r/technology + r/programming)
- GitHub Trending (AI-filtered)

## License

MIT
