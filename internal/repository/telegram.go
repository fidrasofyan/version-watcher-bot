package repository

import (
	"context"
	"time"

	"github.com/fidrasofyan/version-watcher-bot/database"
	"github.com/fidrasofyan/version-watcher-bot/internal/custom_error"
	"github.com/jackc/pgx/v5/pgtype"
)

func TelegramGetChat(ctx context.Context, id int64) (*database.Chat, error) {
	chat, err := database.Sqlc.GetChat(ctx, id)
	if err != nil {
		return nil, custom_error.NewError(err)
	}

	return chat, nil
}

type TelegramSetChatParams struct {
	ID      int64
	Command string
	Step    int16
	Data    []byte
}

func TelegramSetChat(ctx context.Context, arg *TelegramSetChatParams) (*database.Chat, error) {
	datetime := time.Now()

	chatExists, err := database.Sqlc.IsChatExists(ctx, arg.ID)
	if err != nil {
		return nil, custom_error.NewError(err)
	}

	if chatExists {
		chat, err := database.Sqlc.UpdateChat(ctx, &database.UpdateChatParams{
			Command:   arg.Command,
			Step:      arg.Step,
			Data:      arg.Data,
			UpdatedAt: pgtype.Timestamp{Time: datetime, Valid: true},
			ID:        arg.ID,
		})
		if err != nil {
			return nil, custom_error.NewError(err)
		}

		return chat, nil
	}

	chat, err := database.Sqlc.CreateChat(ctx, &database.CreateChatParams{
		ID:        arg.ID,
		Command:   arg.Command,
		Step:      arg.Step,
		Data:      arg.Data,
		CreatedAt: pgtype.Timestamp{Time: datetime, Valid: true},
	})
	if err != nil {
		return nil, custom_error.NewError(err)
	}

	return chat, nil
}

func TelegramDeleteChat(ctx context.Context, id int64) error {
	err := database.Sqlc.DeleteChat(ctx, id)
	if err != nil {
		return custom_error.NewError(err)
	}

	return nil
}
