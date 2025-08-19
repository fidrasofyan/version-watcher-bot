package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/fidrasofyan/version-watcher-bot/database"
	"github.com/fidrasofyan/version-watcher-bot/internal/custom_error"
	"github.com/fidrasofyan/version-watcher-bot/internal/types"
	"github.com/jackc/pgx/v5/pgtype"
)

func Start(ctx context.Context, req types.TelegramUpdate) (*types.TelegramResponse, error) {
	exists, err := database.Sqlc.IsUserExists(ctx, req.Message.Chat.Id)
	if err != nil {
		return nil, custom_error.NewError(err)
	}

	if !exists {
		user, err := database.Sqlc.CreateUser(ctx, &database.CreateUserParams{
			ID:        req.Message.Chat.Id,
			Username:  &req.Message.Chat.Username,
			FirstName: &req.Message.Chat.FirstName,
			LastName:  &req.Message.Chat.LastName,
			CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		})
		if err != nil {
			return nil, custom_error.NewError(err)
		}
		fmt.Println("User created:", *user.FirstName)
	}

	return &types.TelegramResponse{
		Method:      "sendMessage",
		ChatId:      req.Message.Chat.Id,
		ParseMode:   "HTML",
		Text:        "Welcome to Version Watcher. Type /help to see the list of available commands.",
		ReplyMarkup: types.DefaultReplyMarkup,
	}, nil
}
