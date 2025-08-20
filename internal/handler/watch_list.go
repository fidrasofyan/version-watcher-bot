package handler

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/fidrasofyan/version-watcher-bot/database"
	"github.com/fidrasofyan/version-watcher-bot/internal/custom_error"
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
	watchLists, err := database.Sqlc.GetWatchLists(ctx, req.Message.Chat.Id)
	if err != nil {
		return nil, custom_error.NewError(err)
	}

	text := "<b>Watch List</b>\n"

	if len(watchLists) == 0 {
		text += "\n<i>No watch list found</i>"
	} else {
		if len(watchLists) == 1 {
			text += "<i>You watch 1 product</i>\n\n"
		} else {
			text += fmt.Sprintf("<i>You watch %d products</i>\n\n", len(watchLists))
		}
	}

	for _, watchList := range watchLists {
		// Set title
		text += fmt.Sprintf("# <b>%s</b> - <a href=\"%s\">source</a>\n", watchList.ProductLabel, watchList.ProductEolUrl)

		// Set product versions
		productVersions := []productVersion{}
		if len(watchList.ProductVersions) != 0 {
			if err := sonic.Unmarshal(watchList.ProductVersions, &productVersions); err != nil {
				return nil, custom_error.NewError(err)
			}
		}

		for _, pv := range productVersions {
			text += fmt.Sprintf("Version: <code>%s</code> | Label: %s\n", pv.Version, pv.ReleaseLabel)

			if pv.VersionReleaseDate.Valid == true {
				text += fmt.Sprintf("• Latest release: %s\n", pv.VersionReleaseDate.Time.Format("2 Jan 2006"))
			} else {
				text += "• Latest release: -\n"
			}

			if pv.VersionReleaseLink != nil && *pv.VersionReleaseLink != "" {
				text += fmt.Sprintf("• Changelog: <a href=\"%s\">link</a>\n", *pv.VersionReleaseLink)
			} else {
				text += "• Changelog: -\n"
			}
		}
		text += "\n"
	}

	// Limit message length
	if len(text) > 4000 {
		text = text[:4000]
		text += "\n\n<i>--- Part of the message has been truncated ---</i>"
	}

	return &types.TelegramResponse{
		Method:      "sendMessage",
		ChatId:      req.Message.Chat.Id,
		ParseMode:   "HTML",
		Text:        text,
		ReplyMarkup: types.DefaultReplyMarkup,
		LinkPreviewOptions: &types.TelegramLinkPreviewOptions{
			IsDisabled: true,
		},
	}, nil
}
