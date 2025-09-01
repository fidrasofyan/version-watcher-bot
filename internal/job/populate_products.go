package job

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/fidrasofyan/version-watcher-bot/database"
	"github.com/fidrasofyan/version-watcher-bot/internal/config"
	"github.com/fidrasofyan/version-watcher-bot/internal/utils"
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

var httpClient = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	},
	Timeout: 10 * time.Second,
}

func PopulateProducts(ctx context.Context) (*time.Time, error) {
	// Set timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Fetch products
	log.Println("Fetching products...")

	fetchProductsReq, err := http.NewRequestWithContext(
		ctxWithTimeout,
		"GET",
		"https://endoflife.date/api/v1/products",
		nil,
	)
	if err != nil {
		return nil, utils.NewError(err)
	}

	fetchProductsRes, err := httpClient.Do(fetchProductsReq)
	if err != nil {
		return nil, utils.NewError(err)
	}

	var pr FetchProductsResponse
	if err := sonic.ConfigDefault.NewDecoder(fetchProductsRes.Body).Decode(&pr); err != nil {
		return nil, utils.NewError(err)
	}
	fetchProductsRes.Body.Close()
	log.Println("DONE: fetched products:", pr.Total)

	// Start transaction
	tx, err := database.Pool.Begin(ctxWithTimeout)
	if err != nil {
		return nil, utils.NewError(err)
	}
	defer tx.Rollback(ctxWithTimeout) // Always defer rollback (will do nothing if already committed)

	qtx := database.Sqlc.WithTx(tx)
	datetime := time.Now()

	// Populate products
	log.Println("Populating products...")
	for _, p := range pr.Result {
		// Check context
		select {
		case <-ctxWithTimeout.Done():
			return nil, utils.NewError(ctxWithTimeout.Err())
		default:
		}

		err = qtx.UpsertProduct(ctxWithTimeout, &database.UpsertProductParams{
			Name:      p.Name,
			Label:     p.Label,
			Category:  p.Category,
			ApiUrl:    p.URI,
			EolUrl:    strings.Replace(p.URI, "/api/v1/products/", "/", 1),
			CreatedAt: pgtype.Timestamp{Time: datetime, Valid: true},
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NewError(err)
		}
	}
	log.Println("DONE: products populated")

	// Populate product_versions based on watch_lists
	log.Println("Populating product_versions...")

	watchedProducts, err := qtx.GetWatchedProducts(ctxWithTimeout)
	if err != nil {
		return nil, utils.NewError(err)
	}
	log.Println("Watched products:", len(watchedProducts))

	for _, wp := range watchedProducts {
		// Check context
		select {
		case <-ctxWithTimeout.Done():
			return nil, utils.NewError(ctxWithTimeout.Err())
		default:
		}

		// Throttle
		time.Sleep(100 * time.Millisecond)

		// Fetch product
		if config.Cfg.AppEnv == "development" {
			log.Printf("Fetching product %s...", wp.Name)
		}

		fetchProductReq, err := http.NewRequestWithContext(
			ctxWithTimeout,
			"GET",
			wp.ApiUrl.(string),
			nil,
		)
		if err != nil {
			return nil, utils.NewError(err)
		}

		fetchProductRes, err := httpClient.Do(fetchProductReq)
		if err != nil {
			return nil, utils.NewError(err)
		}

		var pr FetchProductDetailResponse
		if err := sonic.ConfigDefault.NewDecoder(fetchProductRes.Body).Decode(&pr); err != nil {
			return nil, utils.NewError(err)
		}
		fetchProductRes.Body.Close()

		for _, release := range pr.Result.Releases {
			// Check context
			select {
			case <-ctxWithTimeout.Done():
				return nil, utils.NewError(ctxWithTimeout.Err())
			default:
			}

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

			err = qtx.CreateProductVersion(ctxWithTimeout, &database.CreateProductVersionParams{
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
				return nil, utils.NewError(err)
			}
		}
	}
	log.Println("DONE: product_versions populated")

	return &datetime, tx.Commit(ctxWithTimeout)
}
