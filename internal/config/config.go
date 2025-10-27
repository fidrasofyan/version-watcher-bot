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

func MustLoadConfig() error {
	godotenv.Load()

	// App timezone
	if os.Getenv("APP_TIMEZONE") == "" {
		fmt.Println("missing env variable: APP_TIMEZONE")
		os.Exit(1)
	}
	os.Setenv("TZ", os.Getenv("APP_TIMEZONE"))

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
		fmt.Println("missing APP_ENV")
		os.Exit(1)
	}
	if Cfg.AppEnv != "development" && Cfg.AppEnv != "production" {
		fmt.Printf("invalid APP_ENV: %s\n", Cfg.AppEnv)
		os.Exit(1)
	}
	if Cfg.AppHost == "" {
		fmt.Println("missing APP_HOST")
		os.Exit(1)
	}
	if Cfg.AppPort == "" {
		fmt.Println("missing APP_PORT")
		os.Exit(1)
	}
	if Cfg.TelegramBotToken == "" {
		fmt.Println("missing TELEGRAM_BOT_TOKEN")
		os.Exit(1)
	}
	if Cfg.DatabaseURL == "" {
		fmt.Println("missing DATABASE_URL")
		os.Exit(1)
	}
	if Cfg.WebhookURL == "" {
		fmt.Println("missing WEBHOOK_URL")
		os.Exit(1)
	}
	if Cfg.WebhookSecretToken == "" {
		fmt.Println("missing WEBHOOK_SECRET_TOKEN")
		os.Exit(1)
	}

	return nil
}
