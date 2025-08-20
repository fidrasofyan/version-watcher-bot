package job

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/fidrasofyan/version-watcher-bot/database"
	"github.com/fidrasofyan/version-watcher-bot/internal/config"
	"github.com/fidrasofyan/version-watcher-bot/internal/custom_error"
	"github.com/jackc/pgx/v5/pgtype"
)

type Product struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Category string `json:"category"`
	URI      string `json:"uri"`
}

type FetchProductsResponse struct {
	SchemaVersion string    `json:"schema_version"`
	GeneratedAt   string    `json:"generated_at"`
	Total         int       `json:"total"`
	Result        []Product `json:"result"`
}

type LatestRelease struct {
	Name *string `json:"name"`
	Date *string `json:"date"`
	Link *string `json:"link"`
}

type Custom struct {
	APIVersion *string `json:"apiVersion"`
}

type Release struct {
	Name        string         `json:"name"`
	Codename    *string        `json:"codename"`
	Label       string         `json:"label"`
	ReleaseDate *string        `json:"releaseDate"`
	Latest      *LatestRelease `json:"latest"`
	Custom      *Custom        `json:"custom"`
}

type FetchProductDetailResult struct {
	Name     string    `json:"name"`
	Label    string    `json:"label"`
	Category string    `json:"category"`
	Releases []Release `json:"releases"`
}

type FetchProductDetailResponse struct {
	SchemaVersion string                   `json:"schema_version"`
	GeneratedAt   string                   `json:"generated_at"`
	LastModified  string                   `json:"last_modified"`
	Result        FetchProductDetailResult `json:"result"`
}

func PopulateProducts(ctx context.Context) (*time.Time, error) {
	// Fetch products
	log.Println("Fetching products...")
	fetchProductsResp, err := http.Get("https://endoflife.date/api/v1/products")
	if err != nil {
		return nil, custom_error.NewError(err)
	}
	defer fetchProductsResp.Body.Close()

	var pr FetchProductsResponse
	if err := sonic.ConfigDefault.NewDecoder(fetchProductsResp.Body).Decode(&pr); err != nil {
		return nil, custom_error.NewError(err)
	}
	log.Println("DONE: fetched products:", pr.Total)

	// Start transaction
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return nil, custom_error.NewError(err)
	}
	defer tx.Rollback(ctx) // Always defer rollback (will do nothing if already committed)

	qtx := database.Sqlc.WithTx(tx)
	datetime := time.Now()

	// Populate products
	log.Println("Populating products...")
	for _, p := range pr.Result {
		err = qtx.UpsertProduct(ctx, &database.UpsertProductParams{
			Name:      p.Name,
			Label:     p.Label,
			Category:  p.Category,
			Uri:       p.URI,
			CreatedAt: pgtype.Timestamp{Time: datetime, Valid: true},
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, custom_error.NewError(err)
		}
	}
	log.Println("DONE: products populated")

	// Populate product_versions based on watch_lists
	log.Println("Populating product_versions...")
	watchedProducts, err := qtx.GetWatchedProducts(ctx)
	if err != nil {
		return nil, custom_error.NewError(err)
	}
	log.Println("Watched products:", len(watchedProducts))

	for _, wp := range watchedProducts {
		// Throttle
		time.Sleep(100 * time.Millisecond)

		// Fetch product
		if config.Cfg.AppEnv == "development" {
			log.Printf("Fetching product %s...", wp.Name)
		}
		fetchProductResp, err := http.Get(wp.Uri.(string))
		if err != nil {
			return nil, custom_error.NewError(err)
		}
		defer fetchProductResp.Body.Close()

		var pr FetchProductDetailResponse
		if err := sonic.ConfigDefault.NewDecoder(fetchProductResp.Body).Decode(&pr); err != nil {
			return nil, custom_error.NewError(err)
		}

		for _, release := range pr.Result.Releases {
			// Insert product_version
			var version string
			var versionReleaseDate time.Time
			var versionReleaseDateValid bool
			var versionReleaseLink *string

			if release.Latest != nil {
				version = *release.Latest.Name
				versionReleaseDate, _ = time.Parse("2006-01-02", *release.Latest.Date)
				versionReleaseDateValid = true
				versionReleaseLink = release.Latest.Link
			} else {
				version = "-"
				versionReleaseDateValid = false
			}

			if release.Custom != nil {
				if release.Custom.APIVersion != nil {
					version = *release.Custom.APIVersion
				}
			}

			releaseDate, _ := time.Parse("2006-01-02", *release.ReleaseDate)

			err = qtx.CreateProductVersion(ctx, &database.CreateProductVersionParams{
				ProductID:          wp.ID,
				ReleaseName:        release.Name,
				ReleaseCodename:    release.Codename,
				ReleaseLabel:       release.Label,
				ReleaseDate:        pgtype.Timestamp{Time: releaseDate, Valid: true},
				Version:            version,
				VersionReleaseDate: pgtype.Timestamp{Time: versionReleaseDate, Valid: versionReleaseDateValid},
				VersionReleaseLink: versionReleaseLink,
				CreatedAt:          pgtype.Timestamp{Time: datetime, Valid: true},
			})
			if err != nil {
				return nil, custom_error.NewError(err)
			}
		}
	}
	log.Println("DONE: product_versions populated")

	return &datetime, tx.Commit(ctx)
}
