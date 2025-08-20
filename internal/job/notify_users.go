package job

import (
	"context"
	"fmt"
	"slices"

	"github.com/bytedance/sonic"
	"github.com/fidrasofyan/version-watcher-bot/database"
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

func NewNotifyUsers(ctx context.Context, errCh chan<- error) func() {
	return func() {
		// Populate products and product_versions with the latest data
		datetime, err := PopulateProducts(ctx)
		if err != nil {
			errCh <- fmt.Errorf("notify users: error populating products: %v", err)
			return
		}

		// Get distinct	product_id from product_versions that just got populated (i.e. new releases)
		productIds, err := database.Sqlc.GetDistinctProductIdsFromProductVersionsByCreatedAt(ctx, pgtype.Timestamp{Time: *datetime, Valid: true})
		if err != nil {
			errCh <- fmt.Errorf("notify users: error getting product versions: %v", err)
			return
		}

		if len(productIds) == 0 {
			return
		}

		// Get products details
		productsWithNewReleases, err := database.Sqlc.GetProductsWithNewReleases(ctx, &database.GetProductsWithNewReleasesParams{
			CreatedAt: pgtype.Timestamp{Time: *datetime, Valid: true},
			Column2:   productIds,
		})
		if err != nil {
			errCh <- fmt.Errorf("notify users: error getting products with new releases: %v", err)
			return
		}

		products := make([]product, len(productsWithNewReleases))
		for i, p := range productsWithNewReleases {
			var productVersions []productVersion
			if err := sonic.Unmarshal(p.ProductVersions, &productVersions); err != nil {
				errCh <- fmt.Errorf("notify users: error unmarshalling product versions: %v", err)
				return
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
			errCh <- fmt.Errorf("notify users: error getting watch lists: %v", err)
			return
		}

		// Notify users
		for _, wl := range watchLists {
			var productIds []int32
			if err := sonic.Unmarshal(wl.ProductIds, &productIds); err != nil {
				errCh <- fmt.Errorf("notify users: error unmarshalling product ids: %v", err)
				return
			}

			filteredProducts := filterProducts(products, productIds)
			if len(filteredProducts) == 0 {
				continue
			}

			// fmt.Println("Chat ID:", wl.ChatID)
			// fmt.Println("Filtered products:", filteredProducts)
			// fmt.Println("Total products:", len(filteredProducts))
			// fmt.Println("\n")

			// Send notification
			var text string
			if len(filteredProducts) > 1 {
				text += "<b>New Releases Detected</b>\n\n"
			} else {
				text += "<b>New Release Detected</b>\n\n"
			}

			for _, p := range filteredProducts {
				// Set title
				text += fmt.Sprintf("# <b>%s</b> - <a href=\"%s\">source</a>\n", p.ProductLabel, p.ProductEolUrl)

				// Set product versions
				for _, pv := range p.ProductVersions {
					text += fmt.Sprintf("Version: <code>%s</code> | Label: %s\n", pv.Version, pv.ReleaseLabel)

					if pv.VersionReleaseDate.Valid {
						text += fmt.Sprintf("• Release: %s\n", pv.VersionReleaseDate.Time.Format("2 Jan 2006"))
					} else {
						text += "• Release: -\n"
					}

					if pv.VersionReleaseLink != nil && *pv.VersionReleaseLink != "" {
						text += fmt.Sprintf("• Changelog: <a href=\"%s\">%s</a>\n", *pv.VersionReleaseLink, "link")
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

			service.SendMessage(&service.SendMessageParams{
				ChatId:    wl.ChatID,
				ParseMode: "HTML",
				Text:      text,
				LinkPreviewOptions: &types.TelegramLinkPreviewOptions{
					IsDisabled: true,
				},
			})
		}
	}
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
