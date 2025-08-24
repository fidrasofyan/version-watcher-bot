package service

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

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

var httpClient = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	},
	Timeout: 10 * time.Second,
}

func SendMessage(ctx context.Context, params *SendMessageParams) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.Cfg.TelegramBotToken)
	jsonData, err := sonic.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

type AnswerCallbackQueryParams struct {
	CallbackQueryId string  `json:"callback_query_id"`
	Text            *string `json:"text,omitempty"`
	ShowAlert       *bool   `json:"show_alert,omitempty"`
}

func AnswerCallbackQuery(ctx context.Context, params *AnswerCallbackQueryParams) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", config.Cfg.TelegramBotToken)
	jsonData, err := sonic.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func SetWebhook(ctx context.Context) error {
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

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

type Command struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

func SetMyCommands(ctx context.Context, commands []Command) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setMyCommands", config.Cfg.TelegramBotToken)
	data := struct {
		Commands []Command `json:"commands"`
	}{
		Commands: commands,
	}
	jsonData, err := sonic.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
