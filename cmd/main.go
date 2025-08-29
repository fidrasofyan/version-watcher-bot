package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/fidrasofyan/version-watcher-bot/database"
	"github.com/fidrasofyan/version-watcher-bot/internal/config"
	"github.com/fidrasofyan/version-watcher-bot/internal/job"
	"github.com/fidrasofyan/version-watcher-bot/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/robfig/cron/v3"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing command")
	}

	mainCtx, cancel := context.WithCancel(context.Background())

	// Load config
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	log.Printf("Environment: %s - Runtime: %s\n", config.Cfg.AppEnv, runtime.Version())

	// Load database
	err = database.LoadDatabase(mainCtx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Setup signal catching
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh,
		os.Interrupt,    // SIGINT (Ctrl+C)
		syscall.SIGTERM, // stop
		syscall.SIGQUIT, // Ctrl+\
		syscall.SIGHUP,  // terminal hangup
	)
	errCh := make(chan error, 1)

	var httpServer *fiber.App
	var cronJob *cron.Cron

	switch os.Args[1] {
	case "start":
		go func() {
			if config.Cfg.AppEnv == "production" {
				log.Println("Setting webhook and commands...")
				// Set webhook
				err := service.SetWebhook(mainCtx)
				if err != nil {
					errCh <- fmt.Errorf("setting webhook: %v", err)
				}

				// Set commands
				commands := []service.Command{
					{
						Command:     "/start",
						Description: "Start the bot",
					},
					{
						Command:     "/watch",
						Description: "Watch a product",
					},
				}
				err = service.SetMyCommands(mainCtx, commands)
				if err != nil {
					errCh <- fmt.Errorf("setting commands: %v", err)
				}

				log.Println("DONE: setting webhook and commands")
			}

			// Start cron job
			cronJob, err = startCronJob(mainCtx)
			if err != nil {
				errCh <- fmt.Errorf("starting cron job: %v", err)
			}

			// Start HTTP server
			httpServer = startHTTPServer(errCh)
		}()

	case "populate-products":
		// Populate products table
		go func() {
			_, err := job.PopulateProducts(mainCtx)
			if err != nil {
				errCh <- fmt.Errorf("populating products: %v", err)
			}
			quitCh <- syscall.SIGQUIT
		}()

	case "notify-users":
		// Notify users
		go func() {
			err := job.NotifyUsers(mainCtx)
			if err != nil {
				errCh <- fmt.Errorf("notifying users: %v", err)
			}
			quitCh <- syscall.SIGQUIT
		}()

	case "set-webhook":
		// Set webhook
		go func() {
			log.Println("Setting webhook...")
			err := service.SetWebhook(mainCtx)
			if err != nil {
				errCh <- fmt.Errorf("setting webhook: %v", err)
			}
			log.Println("DONE: setting webhook")
			quitCh <- syscall.SIGQUIT
		}()

	default:
		log.Printf("Unknown command: %s", os.Args[1])
		quitCh <- syscall.SIGQUIT
	}

	// Wait for signal
	select {
	case sig := <-quitCh:
		log.Printf("Signal caught: %s", sig)
	case err = <-errCh:
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}

	// Cancel context
	cancel()
	<-mainCtx.Done()

	// Stop cron
	if cronJob != nil {
		log.Println("Stopping cron...")
		ctx := cronJob.Stop()
		<-ctx.Done()
	}

	// Stop HTTP server
	if httpServer != nil {
		log.Println("Stopping HTTP server...")
		err = httpServer.Shutdown()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	}

	// Close database
	log.Println("Closing database...")
	database.Pool.Close()

	log.Println("Goodbye!")
}
