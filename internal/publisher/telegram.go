package publisher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hamidghavidel/silicon-brief/internal/scorer"
)

// Publisher sends messages to Telegram.
type Publisher struct {
	botToken  string
	channelID string
	apiURL    string
}

// New creates a new Telegram publisher.
func New(botToken, channelID string) *Publisher {
	return &Publisher{
		botToken:  botToken,
		channelID: channelID,
		apiURL:    "https://api.telegram.org/bot",
	}
}

// Publish sends a story to the Telegram channel.
func (p *Publisher) Publish(ctx context.Context, story scorer.ScoredStory) error {
	message := formatMessage(story)
	apiEndpoint := fmt.Sprintf("%s%s/sendMessage", p.apiURL, p.botToken)

	data := url.Values{}
	data.Set("chat_id", p.channelID)
	data.Set("text", message)
	data.Set("parse_mode", "HTML")
	data.Set("disable_web_page_preview", "false")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("telegram: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram: unexpected status %d", resp.StatusCode)
	}

	return nil
}

func formatMessage(story scorer.ScoredStory) string {
	return fmt.Sprintf(
		"<b>%s</b>\n\n"+
		"📰 <a href=\"%s\">%s</a>\n"+
		"🏷 Source: %s | Score: %.1f\n",
		story.Title,
		story.URL,
		story.Title,
		story.Source,
		story.FinalScore,
	)
}
