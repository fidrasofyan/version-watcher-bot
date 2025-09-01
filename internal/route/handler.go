package route

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/fidrasofyan/version-watcher-bot/internal/handler"
	"github.com/fidrasofyan/version-watcher-bot/internal/repository"
	"github.com/fidrasofyan/version-watcher-bot/internal/types"
	"github.com/fidrasofyan/version-watcher-bot/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req types.TelegramUpdate
		if err := c.BodyParser(&req); err != nil {
			return utils.NewError(err)
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
					Method:      types.TelegramMethodSendMessage,
					ChatId:      req.Message.Chat.Id,
					ParseMode:   types.TelegramParseModeHTML,
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
			err := repository.TelegramDeleteChat(c.UserContext(), chatId)
			if err != nil {
				return utils.NewError(err)
			}
			return c.Status(200).JSON(types.TelegramResponse{
				Method:      types.TelegramMethodSendMessage,
				ChatId:      req.Message.Chat.Id,
				ParseMode:   types.TelegramParseModeHTML,
				Text:        "<i>Cancelled</i>",
				ReplyMarkup: types.DefaultReplyMarkup,
			})
		}

		// Get chat
		chat, err := repository.TelegramGetChat(c.UserContext(), chatId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return utils.NewError(err)
		}
		if chat != nil {
			// Set command
			command = chat.Command
		}

		switch command {
		// Start
		case "start":
			resp, err := handler.Start(c.UserContext(), req)
			if err != nil {
				return utils.NewError(err)
			}
			if resp == nil {
				return c.Status(200).Send(nil)
			}
			return c.Status(200).JSON(resp)

		// Watch
		case "watch":
			resp, err := handler.Watch(c.UserContext(), req)
			if err != nil {
				return utils.NewError(err)
			}
			if resp == nil {
				return c.Status(200).Send(nil)
			}
			return c.Status(200).JSON(resp)

		// Watch list
		case "watch list":
			resp, err := handler.WatchList(c.UserContext(), req)
			if err != nil {
				return utils.NewError(err)
			}
			if resp == nil {
				return c.Status(200).Send(nil)
			}
			return c.Status(200).JSON(resp)

		// Unwatch
		case "unwatch":
			resp, err := handler.UnwatchStep1(c.UserContext(), req)
			if err != nil {
				return utils.NewError(err)
			}
			if resp == nil {
				return c.Status(200).Send(nil)
			}
			return c.Status(200).JSON(resp)

		// Not found
		default:
			if strings.HasPrefix(command, "unwatch_") {
				resp, err := handler.UnwatchStep2(c.UserContext(), req)
				if err != nil {
					return utils.NewError(err)
				}
				if resp == nil {
					return c.Status(200).Send(nil)
				}
				return c.Status(200).JSON(resp)
			}

			resp, err := handler.NotFound(c.UserContext(), req)
			if err != nil || resp == nil {
				return utils.NewError(err)
			}
			return c.Status(200).JSON(resp)

		}
	}
}
