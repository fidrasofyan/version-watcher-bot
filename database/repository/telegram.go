package repository

import (
	"context"
	"time"

	"github.com/fidrasofyan/version-watcher-bot/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func TelegramGetChat(ctx context.Context, id int64) (*database.Chat, error) {
	chat, err := database.Sqlc.GetChat(ctx, id)
	return chat, err
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
		return nil, err
	}

	if chatExists {
		chat, err := database.Sqlc.UpdateChat(ctx, &database.UpdateChatParams{
			Command:   arg.Command,
			Step:      arg.Step,
			Data:      arg.Data,
			UpdatedAt: pgtype.Timestamp{Time: datetime, Valid: true},
			ID:        arg.ID,
		})
		return chat, err
	}

	chat, err := database.Sqlc.CreateChat(ctx, &database.CreateChatParams{
		ID:        arg.ID,
		Command:   arg.Command,
		Step:      arg.Step,
		Data:      arg.Data,
		CreatedAt: pgtype.Timestamp{Time: datetime, Valid: true},
	})
	return chat, err
}

func TelegramDeleteChat(ctx context.Context, id int64) error {
	return database.Sqlc.DeleteChat(ctx, id)
}
