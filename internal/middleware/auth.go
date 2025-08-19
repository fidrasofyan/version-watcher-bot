package middleware

import (
	"github.com/fidrasofyan/version-watcher-bot/internal/config"
	"github.com/gofiber/fiber/v2"
)

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("X-Telegram-Bot-Api-Secret-Token")
		if token != config.Cfg.WebhookSecretToken {
			return c.Status(200).JSON(fiber.Map{
				"ok":      false,
				"message": "Unauthorized",
			})
		}
		return c.Next()
	}
}
