package config

// Telegram holds Telegram bot config.
type Telegram struct {
	BotToken       string `env:"BOT_TOKEN"`
	ChannelID      string `env:"CHANNEL_ID"`
	PublishEnabled bool   `env:"PUBLISH_ENABLED" envDefault:"true"`
}
