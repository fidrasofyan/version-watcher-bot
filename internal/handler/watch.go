package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fidrasofyan/version-watcher-bot/database"
	"github.com/fidrasofyan/version-watcher-bot/internal/custom_error"
	"github.com/fidrasofyan/version-watcher-bot/internal/repository"
	"github.com/fidrasofyan/version-watcher-bot/internal/service"
	"github.com/fidrasofyan/version-watcher-bot/internal/types"
	"github.com/jackc/pgx/v5/pgtype"
)

const command = "watch"

func Watch(ctx context.Context, req types.TelegramUpdate) (*types.TelegramResponse, error) {
	var chatId int64

	// Is it callback query?
	if req.CallbackQuery.Data != "" {
		chatId = req.CallbackQuery.From.Id
	} else {
		chatId = req.Message.Chat.Id
	}

	// Get chat
	chat, err := repository.TelegramGetChat(ctx, chatId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, custom_error.NewError(err)
	}

	if chat == nil {
		// Create new chat
		chat, err = repository.TelegramSetChat(ctx, &repository.TelegramSetChatParams{
			ID:      chatId,
			Command: command,
			Step:    1,
		})
		if err != nil {
			return nil, custom_error.NewError(err)
		}
	}

	switch chat.Step {
	// Step 1
	case 1:
		// Set step
		_, err := repository.TelegramSetChat(ctx, &repository.TelegramSetChatParams{
			ID:      chatId,
			Command: command,
			Step:    2,
		})
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		return &types.TelegramResponse{
			Method:    "sendMessage",
			ChatId:    chatId,
			ParseMode: "HTML",
			Text: strings.Join([]string{
				"What do you want to watch?",
				"\n<i>E.g. Ubuntu, Nginx</i>",
			}, "\n"),
			ReplyMarkup: types.TelegramInlineKeyboardMarkup{
				InlineKeyboard: [][]types.TelegramInlineKeyboardButton{
					{
						{
							Text:         "❌ Cancel",
							CallbackData: "cancel",
						},
					},
				},
			},
		}, nil

	// Step 2
	case 2:
		if req.CallbackQuery.Data == "cancel" {
			// Delete chat
			err := repository.TelegramDeleteChat(ctx, chatId)
			if err != nil {
				return nil, custom_error.NewError(err)
			}

			// Answer callback query
			err = service.AnswerCallbackQuery(&service.AnswerCallbackQueryParams{
				CallbackQueryId: req.CallbackQuery.Id,
			})
			if err != nil {
				return nil, custom_error.NewError(err)
			}

			return &types.TelegramResponse{
				Method:    "editMessageText",
				MessageId: req.CallbackQuery.Message.MessageId,
				ChatId:    chatId,
				ParseMode: "HTML",
				Text:      "<i>Canceled</i>",
			}, nil
		}

		if len(req.Message.Text) < 2 {
			return &types.TelegramResponse{
				Method:    "sendMessage",
				ChatId:    chatId,
				ParseMode: "HTML",
				Text:      "<i>Keyword must be at least 2 characters</i>",
				ReplyMarkup: types.TelegramInlineKeyboardMarkup{
					InlineKeyboard: [][]types.TelegramInlineKeyboardButton{
						{
							{
								Text:         "❌ Cancel",
								CallbackData: "cancel",
							},
						},
					},
				},
			}, nil
		}

		products, err := database.Sqlc.GetProductsByLabel(ctx, fmt.Sprint("%", req.Message.Text, "%"))
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		if len(products) == 0 {
			return &types.TelegramResponse{
				Method:    "sendMessage",
				ChatId:    chatId,
				ParseMode: "HTML",
				Text:      "<i>No products found. Type another keyword...</i>",
				ReplyMarkup: types.TelegramInlineKeyboardMarkup{
					InlineKeyboard: [][]types.TelegramInlineKeyboardButton{
						{
							{
								Text:         "❌ Cancel",
								CallbackData: "cancel",
							},
						},
					},
				},
			}, nil
		}

		inlineKeyboard := make([][]types.TelegramInlineKeyboardButton, len(products)+1) // +1 for cancel button

		for i, product := range products {
			inlineKeyboard[i] = []types.TelegramInlineKeyboardButton{
				{
					Text:         product.Label,
					CallbackData: fmt.Sprint(product.ID),
				},
			}
		}

		inlineKeyboard[len(products)] = []types.TelegramInlineKeyboardButton{
			{
				Text: "❌ Cancel", CallbackData: "cancel",
			},
		}

		// Set step
		_, err = repository.TelegramSetChat(ctx, &repository.TelegramSetChatParams{
			ID:      chatId,
			Command: command,
			Step:    3,
		})
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		return &types.TelegramResponse{
			Method:    "sendMessage",
			ChatId:    chatId,
			ParseMode: "HTML",
			Text:      "Choose product:",
			ReplyMarkup: types.TelegramInlineKeyboardMarkup{
				InlineKeyboard: inlineKeyboard,
			},
		}, nil

	// Step 3
	case 3:
		// It must be callback query
		if req.CallbackQuery.Data == "" {
			// Delete chat
			err := repository.TelegramDeleteChat(ctx, chatId)
			if err != nil {
				return nil, custom_error.NewError(err)
			}

			return &types.TelegramResponse{
				Method:    "sendMessage",
				ChatId:    chatId,
				ParseMode: "HTML",
				Text:      "<i>Invalid command</i>",
			}, nil
		}

		if req.CallbackQuery.Data == "cancel" {
			// Delete chat
			err := repository.TelegramDeleteChat(ctx, chatId)
			if err != nil {
				return nil, custom_error.NewError(err)
			}

			// Answer callback query
			err = service.AnswerCallbackQuery(&service.AnswerCallbackQueryParams{
				CallbackQueryId: req.CallbackQuery.Id,
			})
			if err != nil {
				return nil, custom_error.NewError(err)
			}

			return &types.TelegramResponse{
				Method:    "editMessageText",
				MessageId: req.CallbackQuery.Message.MessageId,
				ChatId:    chatId,
				ParseMode: "HTML",
				Text:      "<i>Canceled</i>",
			}, nil
		}

		productId64, err := strconv.ParseInt(req.CallbackQuery.Data, 10, 32)
		if err != nil {
			return nil, custom_error.NewError(err)
		}
		productId := int32(productId64)

		product, err := database.Sqlc.GetProductById(ctx, productId)
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		// Is it already in watch list?
		isWatchListExists, err := database.Sqlc.IsWatchListExists(ctx, &database.IsWatchListExistsParams{
			ChatID:    chatId,
			ProductID: productId,
		})
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		if isWatchListExists {
			// Delete chat
			err = repository.TelegramDeleteChat(ctx, chatId)
			if err != nil {
				return nil, custom_error.NewError(err)
			}

			// Answer callback query
			err = service.AnswerCallbackQuery(&service.AnswerCallbackQueryParams{
				CallbackQueryId: req.CallbackQuery.Id,
			})
			if err != nil {
				return nil, custom_error.NewError(err)
			}

			return &types.TelegramResponse{
				Method:    "editMessageText",
				MessageId: req.CallbackQuery.Message.MessageId,
				ChatId:    chatId,
				ParseMode: "HTML",
				Text:      fmt.Sprintf("<i>❌ %s is already in watch list</i>", product.Label),
			}, nil
		}

		// Add to watch list
		_, err = database.Sqlc.CreateWatchList(ctx, &database.CreateWatchListParams{
			ChatID:    chatId,
			ProductID: productId,
			CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		})
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		// Delete chat
		err = repository.TelegramDeleteChat(ctx, chatId)
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		// Answer callback query
		err = service.AnswerCallbackQuery(&service.AnswerCallbackQueryParams{
			CallbackQueryId: req.CallbackQuery.Id,
		})
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		var textB strings.Builder
		textB.WriteString(fmt.Sprintf("✅ %s added to watch list\n\n", product.Label))
		textB.WriteString("<i>*You'll be notified when a new version is released</i>")

		return &types.TelegramResponse{
			Method:    "editMessageText",
			MessageId: req.CallbackQuery.Message.MessageId,
			ChatId:    chatId,
			ParseMode: "HTML",
			Text:      textB.String(),
		}, nil

	// Unhandled step
	default:
		// Delete step
		err := repository.TelegramDeleteChat(ctx, chatId)
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		return &types.TelegramResponse{
			Method:      "sendMessage",
			ChatId:      req.Message.Chat.Id,
			ParseMode:   "HTML",
			Text:        "<i>Unhandled step</i>",
			ReplyMarkup: types.DefaultReplyMarkup,
		}, nil
	}

}
