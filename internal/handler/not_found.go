package handler

import (
	"context"

	"github.com/fidrasofyan/version-watcher-bot/internal/custom_error"
	"github.com/fidrasofyan/version-watcher-bot/internal/repository"
	"github.com/fidrasofyan/version-watcher-bot/internal/service"
	"github.com/fidrasofyan/version-watcher-bot/internal/types"
)

func NotFound(ctx context.Context, req types.TelegramUpdate) (*types.TelegramResponse, error) {
	// Is it callback query?
	if req.CallbackQuery.Id != "" {
		// Delete chat
		err := repository.TelegramDeleteChat(ctx, req.CallbackQuery.From.Id)
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		// Answer callback query
		err = service.AnswerCallbackQuery(ctx, &service.AnswerCallbackQueryParams{
			CallbackQueryId: req.CallbackQuery.Id,
		})
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		return &types.TelegramResponse{
			Method:    types.TelegramMethodEditMessageText,
			MessageId: req.CallbackQuery.Message.MessageId,
			ChatId:    req.CallbackQuery.Message.Chat.Id,
			ParseMode: types.TelegramParseModeHTML,
			Text:      "<i>Invalid session</i>",
		}, nil
	}

	return &types.TelegramResponse{
		Method:      types.TelegramMethodSendMessage,
		ChatId:      req.Message.Chat.Id,
		ParseMode:   types.TelegramParseModeHTML,
		Text:        "<i>Unknown command</i>",
		ReplyMarkup: types.DefaultReplyMarkup,
	}, nil
}
