package route

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/fidrasofyan/version-watcher-bot/internal/custom_error"
	"github.com/fidrasofyan/version-watcher-bot/internal/handler"
	"github.com/fidrasofyan/version-watcher-bot/internal/repository"
	"github.com/fidrasofyan/version-watcher-bot/internal/service"
	"github.com/fidrasofyan/version-watcher-bot/internal/types"
	"github.com/gofiber/fiber/v2"
)

func Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req types.TelegramUpdate
		if err := c.BodyParser(&req); err != nil {
			return custom_error.NewError(err)
		}

		var chatId int64
		var command string

		// Is it callback query?
		if req.CallbackQuery.Id != "" {
			// Set chat id
			chatId = req.CallbackQuery.From.Id
		} else {
			// Set chat id
			chatId = req.Message.Chat.Id

			// Only text message is supported
			if req.Message.Text == "" {
				return c.Status(200).JSON(types.TelegramResponse{
					Method:      "sendMessage",
					ChatId:      req.Message.Chat.Id,
					ParseMode:   "HTML",
					Text:        "<i>Only text command is supported</i>",
					ReplyMarkup: types.DefaultReplyMarkup,
				})
			}

			// Set command
			command = strings.TrimSpace(strings.ToLower(
				req.Message.Text,
			))
			// Limit command length
			if len(command) > 30 {
				command = command[:30]
			}
			// Remove leading slashes
			command = strings.TrimLeft(command, "/")
		}

		// Is it "cancel" command?
		if command == "cancel" {
			// Delete chat
			err := repository.TelegramDeleteChat(c.Context(), chatId)
			if err != nil {
				return custom_error.NewError(err)
			}
			return c.Status(200).JSON(types.TelegramResponse{
				Method:      "sendMessage",
				ChatId:      req.Message.Chat.Id,
				ParseMode:   "HTML",
				Text:        "<i>Cancelled</i>",
				ReplyMarkup: types.DefaultReplyMarkup,
			})
		}

		// Get chat
		chat, err := repository.TelegramGetChat(c.Context(), chatId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return custom_error.NewError(err)
		}
		if chat != nil {
			// Set command
			command = chat.Command
		}

		switch command {
		// Start
		case "start":
			resp, err := handler.Start(c.Context(), req)
			if err != nil || resp == nil {
				return custom_error.NewError(err)
			}
			return c.Status(200).JSON(resp)

		// Watch
		case "watch":
			resp, err := handler.Watch(c.Context(), req)
			if err != nil || resp == nil {
				return custom_error.NewError(err)
			}
			return c.Status(200).JSON(resp)

		// Watch list
		case "watch list":
			resp, err := handler.WatchList(c.Context(), req)
			if err != nil || resp == nil {
				return custom_error.NewError(err)
			}
			return c.Status(200).JSON(resp)

		// Not found
		default:
			// Is it callback query?
			if req.CallbackQuery.Id != "" {
				// Delete chat
				err := repository.TelegramDeleteChat(c.Context(), chatId)
				if err != nil {
					return custom_error.NewError(err)
				}

				// Answer callback query
				err = service.AnswerCallbackQuery(&service.AnswerCallbackQueryParams{
					CallbackQueryId: req.CallbackQuery.Id,
				})
				if err != nil {
					return custom_error.NewError(err)
				}

				return c.Status(200).JSON(types.TelegramResponse{
					Method:    "editMessageText",
					MessageId: req.CallbackQuery.Message.MessageId,
					ChatId:    chatId,
					ParseMode: "HTML",
					Text:      "<i>Invalid session</i>",
				})
			}

			return c.Status(200).JSON(&types.TelegramResponse{
				Method:      "sendMessage",
				ChatId:      req.Message.Chat.Id,
				ParseMode:   "HTML",
				Text:        "<i>Unknown command</i>",
				ReplyMarkup: types.DefaultReplyMarkup,
			})
		}
	}
}
