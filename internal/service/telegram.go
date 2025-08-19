package service

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/fidrasofyan/version-watcher-bot/internal/config"
	"github.com/fidrasofyan/version-watcher-bot/internal/types"
)

type SendMessageParams struct {
	ChatId             int64                             `json:"chat_id"`
	ParseMode          string                            `json:"parse_mode"`
	Text               string                            `json:"text"`
	LinkPreviewOptions *types.TelegramLinkPreviewOptions `json:"link_preview_options,omitempty"`
}

func SendMessage(params *SendMessageParams) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.Cfg.TelegramBotToken)
	jsonData, err := sonic.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func SetWebhook() error {
	log.Println("Setting webhook...")
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", config.Cfg.TelegramBotToken)
	data := struct {
		Url                string   `json:"url"`
		SecretToken        string   `json:"secret_token"`
		MaxConnections     int      `json:"max_connections"`
		DropPendingUpdates bool     `json:"drop_pending_updates"`
		AllowedUpdates     []string `json:"allowed_updates"`
	}{
		Url:                config.Cfg.WebhookURL + "/webhook",
		SecretToken:        config.Cfg.WebhookSecretToken,
		MaxConnections:     50,
		DropPendingUpdates: true,
		AllowedUpdates:     []string{"message", "callback_query"},
	}
	jsonData, err := sonic.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println("DONE: set webhook")

	return nil
}
