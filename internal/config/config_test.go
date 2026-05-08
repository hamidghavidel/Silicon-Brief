package config

import (
	"os"
	"testing"
)

func TestLoad_MissingBotToken(t *testing.T) {
	os.Clearenv()
	os.Setenv("TELEGRAM_CHANNEL_ID", "test")
	os.Setenv("FIREBASE_PROJECT_ID", "test")
	os.Setenv("FIREBASE_SERVICE_ACCOUNT_JSON", "test")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing TELEGRAM_BOT_TOKEN")
	}
}

func TestLoad_InvalidBotToken(t *testing.T) {
	os.Clearenv()
	os.Setenv("TELEGRAM_BOT_TOKEN", "bad")
	os.Setenv("TELEGRAM_CHANNEL_ID", "-100")
	os.Setenv("FIREBASE_PROJECT_ID", "test")
	os.Setenv("FIREBASE_SERVICE_ACCOUNT_JSON", "test")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid TELEGRAM_BOT_TOKEN")
	}
}

func TestLoad_Success(t *testing.T) {
	os.Clearenv()
	os.Setenv("TELEGRAM_BOT_TOKEN", "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11")
	os.Setenv("TELEGRAM_CHANNEL_ID", "-1001234567890")
	os.Setenv("FIREBASE_PROJECT_ID", "my-project")
	os.Setenv("FIREBASE_SERVICE_ACCOUNT_JSON", "{}")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TelegramBotToken == "" {
		t.Fatal("expected bot token to be set")
	}
	if cfg.MaxPosts != 15 {
		t.Fatalf("expected MaxPosts=15, got %d", cfg.MaxPosts)
	}
}
