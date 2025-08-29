package job

import (
	"context"
	"fmt"
	"slices"
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

type product struct {
	ProductId       int32
	ProductLabel    string
	ProductEolUrl   string
	ProductVersions []productVersion
}

func NotifyUsers(ctx context.Context) error {
	// Populate products and product_versions with the latest data
	datetime, err := PopulateProducts(ctx)
	if err != nil {
		return custom_error.NewError(err)
	}

	// Get distinct	product_id from product_versions that just got populated (i.e. new releases)
	productIds, err := database.Sqlc.GetDistinctProductIdsFromProductVersionsByCreatedAt(ctx, pgtype.Timestamp{Time: *datetime, Valid: true})
	if err != nil {
		return custom_error.NewError(err)
	}

	if len(productIds) == 0 {
		return nil
	}

	// Get products details
	productsWithNewReleases, err := database.Sqlc.GetProductsWithNewReleases(ctx, &database.GetProductsWithNewReleasesParams{
		CreatedAt: pgtype.Timestamp{Time: *datetime, Valid: true},
		Column2:   productIds,
	})
	if err != nil {
		return custom_error.NewError(err)
	}

	products := make([]product, len(productsWithNewReleases))
	for i, p := range productsWithNewReleases {
		var productVersions []productVersion
		if err := sonic.Unmarshal(p.ProductVersions, &productVersions); err != nil {
			return custom_error.NewError(err)
		}
		products[i] = product{
			ProductId:       p.ProductID,
			ProductLabel:    p.ProductLabel,
			ProductEolUrl:   p.ProductEolUrl,
			ProductVersions: productVersions,
		}
	}

	// Get watch lists
	watchLists, err := database.Sqlc.GetWatchListsGroupedByChat(ctx)
	if err != nil {
		return custom_error.NewError(err)
	}

	// Notify users
	for _, wl := range watchLists {
		var productIds []int32
		if err := sonic.Unmarshal(wl.ProductIds, &productIds); err != nil {
			return custom_error.NewError(err)
		}

		filteredProducts := filterProducts(products, productIds)
		if len(filteredProducts) == 0 {
			continue
		}

		// Send notification
		textLimit := 3500
		var textB strings.Builder
		if len(filteredProducts) > 1 {
			textB.WriteString("<b>New Releases Detected</b>\n\n")
		} else {
			textB.WriteString("<b>New Release Detected</b>\n\n")
		}

		for _, p := range filteredProducts {
			// Set title
			textB.WriteString(fmt.Sprintf("# <b>%s</b> - <a href=\"%s\">source</a>\n", p.ProductLabel, p.ProductEolUrl))

			// Set product versions
			for _, pv := range p.ProductVersions {
				textB.WriteString(fmt.Sprintf("Version: <code>%s</code> | Label: %s\n", pv.Version, pv.ReleaseLabel))

				if pv.VersionReleaseDate.Valid {
					textB.WriteString(fmt.Sprintf("• Release: %s\n", pv.VersionReleaseDate.Time.Format("2 Jan 2006")))
				} else {
					textB.WriteString("• Release: -\n")
				}

				if pv.VersionReleaseLink != nil && *pv.VersionReleaseLink != "" {
					textB.WriteString(fmt.Sprintf("• Changelog: <a href=\"%s\">%s</a>\n", *pv.VersionReleaseLink, "link"))
				} else {
					textB.WriteString("• Changelog: -\n")
				}
			}
			textB.WriteString("\n")

			// If text is too long, send it part by part
			if textB.Len() >= textLimit {
				service.SendMessage(ctx, &service.SendMessageParams{
					ChatId:    wl.ChatID,
					ParseMode: service.TelegramParseModeHTML,
					Text:      textB.String(),
					LinkPreviewOptions: &types.TelegramLinkPreviewOptions{
						IsDisabled: true,
					},
				})
				textB.Reset()
			}
		}

		if textB.Len() == 0 {
			return nil
		}

		service.SendMessage(ctx, &service.SendMessageParams{
			ChatId:    wl.ChatID,
			ParseMode: service.TelegramParseModeHTML,
			Text:      textB.String(),
			LinkPreviewOptions: &types.TelegramLinkPreviewOptions{
				IsDisabled: true,
			},
		})
	}

	return nil
}

func filterProducts(products []product, productIds []int32) []product {
	filteredProducts := make([]product, 0, len(productIds))
	for _, p := range products {
		if slices.Contains(productIds, p.ProductId) {
			filteredProducts = append(filteredProducts, p)
		}
	}
	return filteredProducts
}
