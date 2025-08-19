package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	AppHost            string
	AppPort            string
	TelegramBotToken   string
	DatabaseURL        string
	WebhookURL         string
	WebhookSecretToken string
}

var Cfg *Config

func LoadConfig() error {
	godotenv.Load()

	Cfg = &Config{
		AppEnv:             os.Getenv("APP_ENV"),
		AppHost:            os.Getenv("APP_HOST"),
		AppPort:            os.Getenv("APP_PORT"),
		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		WebhookURL:         os.Getenv("WEBHOOK_URL"),
		WebhookSecretToken: os.Getenv("WEBHOOK_SECRET_TOKEN"),
	}

	// Validate
	if Cfg.AppEnv == "" {
		return fmt.Errorf("missing APP_ENV")
	}
	if Cfg.AppEnv != "development" && Cfg.AppEnv != "production" {
		return fmt.Errorf("invalid APP_ENV: %s", Cfg.AppEnv)
	}
	if Cfg.AppHost == "" {
		return fmt.Errorf("missing APP_HOST")
	}
	if Cfg.AppPort == "" {
		return fmt.Errorf("missing APP_PORT")
	}
	if Cfg.TelegramBotToken == "" {
		return fmt.Errorf("missing TELEGRAM_BOT_TOKEN")
	}
	if Cfg.DatabaseURL == "" {
		return fmt.Errorf("missing DATABASE_URL")
	}
	if Cfg.WebhookURL == "" {
		return fmt.Errorf("missing WEBHOOK_URL")
	}
	if Cfg.WebhookSecretToken == "" {
		return fmt.Errorf("missing WEBHOOK_SECRET_TOKEN")
	}

	return nil
}
