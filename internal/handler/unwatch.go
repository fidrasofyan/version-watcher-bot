package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/fidrasofyan/version-watcher-bot/database"
	"github.com/fidrasofyan/version-watcher-bot/internal/custom_error"
	"github.com/fidrasofyan/version-watcher-bot/internal/repository"
	"github.com/fidrasofyan/version-watcher-bot/internal/service"
	"github.com/fidrasofyan/version-watcher-bot/internal/types"
)

func UnwatchStep1(ctx context.Context, req types.TelegramUpdate) (*types.TelegramResponse, error) {
	watchList, err := database.Sqlc.GetWatchList(ctx, req.Message.Chat.Id)
	if err != nil {
		return nil, custom_error.NewError(err)
	}

	textLimit := 3500
	var textB strings.Builder
	textB.WriteString("<b>Watch List</b>\n")

	switch len(watchList) {
	case 0:
		textB.WriteString("\n<i>No watch list found</i>")
	case 1:
		textB.WriteString("<i>You watch 1 product</i>\n\n")
	default:
		textB.WriteString(fmt.Sprintf("<i>You watch %d products</i>\n\n", len(watchList)))

		for _, watchListItem := range watchList {
			// Normalize product name
			productName := strings.ReplaceAll(watchListItem.ProductName, "-", "_")
			textB.WriteString(fmt.Sprintf("• %s - /unwatch_%s\n", watchListItem.ProductLabel, productName))

			// If text is too long, send it part by part
			if textB.Len() >= textLimit {
				service.SendMessage(ctx, &service.SendMessageParams{
					ChatId:    req.Message.Chat.Id,
					ParseMode: "HTML",
					Text:      textB.String(),
					LinkPreviewOptions: &types.TelegramLinkPreviewOptions{
						IsDisabled: true,
					},
				})
				textB.Reset()
			}
		}
	}

	if textB.Len() == 0 {
		return nil, nil
	}

	return &types.TelegramResponse{
		Method:    "sendMessage",
		ChatId:    req.Message.Chat.Id,
		ParseMode: "HTML",
		Text:      textB.String(),
	}, nil
}

type productData struct {
	ID    int32  `json:"id"`
	Label string `json:"label"`
}

func UnwatchStep2(ctx context.Context, req types.TelegramUpdate) (*types.TelegramResponse, error) {
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
			Command: "unwatch_",
			Step:    1,
		})
		if err != nil {
			return nil, custom_error.NewError(err)
		}
	}

	switch chat.Step {
	// Step 1
	case 1:
		productName := strings.Replace(req.Message.Text, "/unwatch_", "", 1)
		productName = strings.ReplaceAll(productName, "_", "-")

		product, err := database.Sqlc.GetWatchedProductByName(ctx, &database.GetWatchedProductByNameParams{
			Name:   productName,
			ChatID: chatId,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Delete chat
				err := repository.TelegramDeleteChat(ctx, chatId)
				if err != nil {
					return nil, custom_error.NewError(err)
				}

				return &types.TelegramResponse{
					Method:      "sendMessage",
					ChatId:      req.Message.Chat.Id,
					ParseMode:   "HTML",
					Text:        "<i>Product not found</i>",
					ReplyMarkup: types.DefaultReplyMarkup,
				}, nil
			}
			return nil, custom_error.NewError(err)
		}

		productData := productData{
			ID:    product.ID,
			Label: product.Label,
		}
		productDataB, err := sonic.Marshal(productData)
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		// Set step
		_, err = repository.TelegramSetChat(ctx, &repository.TelegramSetChatParams{
			ID:      chatId,
			Command: "unwatch_",
			Step:    2,
			Data:    productDataB,
		})
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		return &types.TelegramResponse{
			Method:    "sendMessage",
			ChatId:    req.Message.Chat.Id,
			ParseMode: "HTML",
			Text:      fmt.Sprintf("Are you sure you want to unwatch <b>%s</b>?", product.Label),
			ReplyMarkup: types.TelegramReplyKeyboardMarkup{
				ResizeKeyboard: true,
				Keyboard: [][]string{
					{"Yes", "No"},
				},
			},
		}, nil

	// Step 2
	case 2:
		if req.Message.Text != "Yes" {
			// Delete chat
			err := repository.TelegramDeleteChat(ctx, chatId)
			if err != nil {
				return nil, custom_error.NewError(err)
			}

			return &types.TelegramResponse{
				Method:      "sendMessage",
				ChatId:      req.Message.Chat.Id,
				ParseMode:   "HTML",
				Text:        "<i>Cancelled</i>",
				ReplyMarkup: types.DefaultReplyMarkup,
			}, nil
		}

		// Get chat data
		chat, err := repository.TelegramGetChat(ctx, chatId)
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		// Get product data
		productData := productData{}
		if err := sonic.Unmarshal(chat.Data, &productData); err != nil {
			return nil, custom_error.NewError(err)
		}

		// Delete watch list
		err = database.Sqlc.DeleteWatchList(ctx, &database.DeleteWatchListParams{
			ChatID:    chatId,
			ProductID: productData.ID,
		})
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		// Delete chat
		err = repository.TelegramDeleteChat(ctx, chatId)
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		return &types.TelegramResponse{
			Method:      "sendMessage",
			ChatId:      req.Message.Chat.Id,
			ParseMode:   "HTML",
			Text:        fmt.Sprintf("<b>%s</b> removed from watch list", productData.Label),
			ReplyMarkup: types.DefaultReplyMarkup,
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
