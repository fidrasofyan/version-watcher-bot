package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fidrasofyan/version-watcher-bot/internal/job"
	"github.com/robfig/cron/v3"
)

func startCronJob(ctx context.Context) (*cron.Cron, error) {
	c := cron.New()

	errCh := make(chan error, 1)
	// Listen for errors
	go func() {
		for err := range errCh {
			log.Printf("Cron: error: %v", err)
		}
	}()

	_, err := c.AddFunc("*/15 * * * *", job.NewNotifyUsers(ctx, errCh))
	if err != nil {
		return nil, fmt.Errorf("error adding function: %v", err)
	}

	c.Start()
	log.Println("Cron job started")
	return c, nil
}
