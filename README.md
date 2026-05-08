# Silicon-Brief

A zero-cost automated Telegram channel that aggregates, ranks, and publishes AI/ML/LLM tech news from free sources.

## Architecture

```
Systemd Timer (cron: 0 * * * *)
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
  SQLite (state)
```

## Setup

### 1. Clone & Build

```bash
git clone https://github.com/hamidghavidel/silicon-brief.git
cd silicon-brief
go build -o silicon-brief ./cmd/silicon-brief
```

### 2. Configure Environment

Set these environment variables (e.g., in `~/.bashrc` or the systemd service file):

| Variable | Description |
|----------|-------------|
| `TELEGRAM_BOT_TOKEN` | From [@BotFather](https://t.me/botfather) |
| `TELEGRAM_CHANNEL_ID` | Target channel ID (e.g., `-1001234567890` or `@channelname`) |

### 3. Deploy to VPS

```bash
# Copy binary
sudo cp silicon-brief /usr/local/bin/
sudo chmod +x /usr/local/bin/silicon-brief

# Create config directory
sudo mkdir -p /etc/silicon-brief
sudo cp service-account.json /etc/silicon-brief/

# Copy and enable systemd service
sudo cp silicon-brief.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable silicon-brief

# Check status
sudo systemctl status silicon-brief
```

### 4. Docker Deployment (Alternative)

```bash
# Create .env file
cat > .env <<EOF
TELEGRAM_BOT_TOKEN=your-token
TELEGRAM_CHANNEL_ID=your-channel-id
EOF

# Build and run
docker-compose up --build -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### 5. Schedule with Cron

```bash
# Edit crontab
crontab -e

# Add line to run every hour
0 * * * * docker-compose -f /path/to/docker-compose.yml up --abort-on-container-exit >> /var/log/silicon-brief.log 2>&1
```

## Local Development

```bash
# Set environment variables
export TELEGRAM_BOT_TOKEN="your-token"
export TELEGRAM_CHANNEL_ID="your-channel-id"

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
│   ├── store/sqlite.go            # Published post tracking
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
