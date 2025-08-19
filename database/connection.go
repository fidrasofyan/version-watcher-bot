package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/fidrasofyan/version-watcher-bot/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var Pool *pgxpool.Pool
var Sqlc *Queries

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

func LoadDatabase() error {
	ctx := context.TODO()
	var err error

	// Open database
	pgxConfig, err := pgxpool.ParseConfig(config.Cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("error parsing database URL: %v", err)
	}
	pgxConfig.MaxConns = 20

	Pool, err = pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v\n", err)
	}
	config := Pool.Config()

	// Ping
	if err := Pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Initialize SQLC
	Sqlc = New(Pool)
	log.Printf("Database connection established - MaxConn: %d", config.MaxConns)

	// Apply migrations
	if err := applyMigrations(); err != nil {
		return fmt.Errorf("failed to apply migrations: %v", err)
	}

	return nil
}

func applyMigrations() error {
	// Setup database
	db, err := sql.Open("pgx", config.Cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	// Ping
	err = db.Ping()
	if err != nil {
		return err
	}

	// Apply migrations
	goose.SetBaseFS(embeddedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}
