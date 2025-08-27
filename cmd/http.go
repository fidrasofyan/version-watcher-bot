package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bytedance/sonic"
	"github.com/fidrasofyan/version-watcher-bot/internal/config"
	"github.com/fidrasofyan/version-watcher-bot/internal/middleware"
	"github.com/fidrasofyan/version-watcher-bot/internal/repository"
	"github.com/fidrasofyan/version-watcher-bot/internal/route"
	"github.com/fidrasofyan/version-watcher-bot/internal/service"
	"github.com/fidrasofyan/version-watcher-bot/internal/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/timeout"
)

func startHTTPServer(errCh chan<- error) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "Version Watcher Bot",
		Prefork:               false,
		DisableStartupMessage: true,
		JSONEncoder:           sonic.Marshal,
		JSONDecoder:           sonic.Unmarshal,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Error: %v", err)

			text := "<i>Something went wrong</i>"
			if errors.Is(err, fiber.ErrRequestTimeout) {
				text = "<i>Request timeout</i>"
			}

			var body types.TelegramUpdate
			if err := c.BodyParser(&body); err != nil {
				return err
			}

			// Delete chat
			_ = repository.TelegramDeleteChat(c.Context(), body.CallbackQuery.From.Id)

			if body.CallbackQuery.Id != "" {
				// Answer callback query
				_ = service.AnswerCallbackQuery(c.Context(), &service.AnswerCallbackQueryParams{
					CallbackQueryId: body.CallbackQuery.Id,
				})

				return c.Status(200).JSON(types.TelegramResponse{
					Method:    types.TelegramMethodEditMessageText,
					MessageId: body.CallbackQuery.Message.MessageId,
					ChatId:    body.CallbackQuery.Message.Chat.Id,
					ParseMode: types.TelegramParseModeHTML,
					Text:      text,
				})
			}

			return c.Status(200).JSON(types.TelegramResponse{
				Method:      types.TelegramMethodSendMessage,
				ChatId:      body.Message.Chat.Id,
				ParseMode:   types.TelegramParseModeHTML,
				Text:        text,
				ReplyMarkup: types.DefaultReplyMarkup,
			})
		},
	})

	// Middlewares
	app.Use(recover.New())
	if config.Cfg.AppEnv == "development" {
		app.Use(logger.New())
	}

	// Routes
	app.Post(
		"/webhook",
		middleware.Protected(),
		timeout.NewWithContext(route.Handler(), 10*time.Second),
	)

	// Not found
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"message": "Not found",
		})
	})

	go func() {
		log.Printf("Server is running on http://%s:%s", config.Cfg.AppHost, config.Cfg.AppPort)
		err := app.Listen(config.Cfg.AppHost + ":" + config.Cfg.AppPort)
		if err != nil {
			errCh <- fmt.Errorf("failed to start HTTP server: %v", err)
		}
	}()

	return app
}
