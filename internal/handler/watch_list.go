package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/fidrasofyan/version-watcher-bot/database"
	"github.com/fidrasofyan/version-watcher-bot/internal/custom_error"
	"github.com/fidrasofyan/version-watcher-bot/internal/service"
	"github.com/fidrasofyan/version-watcher-bot/internal/types"
	"github.com/jackc/pgx/v5/pgtype"
)

type productVersion struct {
	ReleaseLabel       string           `json:"release_label"`
	Version            string           `json:"version"`
	VersionReleaseDate pgtype.Timestamp `json:"version_release_date"`
	VersionReleaseLink *string          `json:"version_release_link"`
}

func WatchList(ctx context.Context, req types.TelegramUpdate) (*types.TelegramResponse, error) {
	watchLists, err := database.Sqlc.GetWatchListsWithProductVersions(ctx, req.Message.Chat.Id)
	if err != nil {
		return nil, custom_error.NewError(err)
	}

	textLimit := 3500
	var textB strings.Builder
	textB.WriteString("<b>Watch List</b>\n")

	switch len(watchLists) {
	case 0:
		textB.WriteString("\n<i>No watch list found</i>")
	case 1:
		textB.WriteString("<i>You watch 1 product</i>\n\n")
	default:
		textB.WriteString(fmt.Sprintf("<i>You watch %d products</i>\n\n", len(watchLists)))
	}

	for _, watchList := range watchLists {
		// Set title
		textB.WriteString(fmt.Sprintf("# <b>%s</b> - <a href=\"%s\">source</a>\n", watchList.ProductLabel, watchList.ProductEolUrl))

		// Set product versions
		productVersions := []productVersion{}
		if len(watchList.ProductVersions) != 0 {
			if err := sonic.Unmarshal(watchList.ProductVersions, &productVersions); err != nil {
				return nil, custom_error.NewError(err)
			}
		}

		for _, pv := range productVersions {
			if pv.VersionReleaseDate.Valid {
				textB.WriteString(fmt.Sprintf("• Latest: %s - %s\n", pv.Version, pv.VersionReleaseDate.Time.Format("2 Jan 2006")))
			} else {
				textB.WriteString("• Latest release: -\n")
			}
		}

		// If text is too long, send it part by part
		if textB.Len() >= textLimit {
			service.SendMessage(&service.SendMessageParams{
				ChatId:             req.Message.Chat.Id,
				ParseMode:          "HTML",
				Text:               textB.String(),
				LinkPreviewOptions: &types.TelegramLinkPreviewOptions{IsDisabled: true},
			})
			textB.Reset()
		}
	}

	if textB.Len() == 0 {
		return nil, nil
	}

	return &types.TelegramResponse{
		Method:      "sendMessage",
		ChatId:      req.Message.Chat.Id,
		ParseMode:   "HTML",
		Text:        textB.String(),
		ReplyMarkup: types.DefaultReplyMarkup,
		LinkPreviewOptions: &types.TelegramLinkPreviewOptions{
			IsDisabled: true,
		},
	}, nil
}
