package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"planning-poker/internal/server"
	"strconv"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

func gracefulShutdown(fiberServer *server.FiberServer, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := fiberServer.ShutdownWithContext(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")
	done <- true
}

func main() {
	// Initialize Goth with OAuth providers
	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
			os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
			os.Getenv("GOOGLE_OAUTH_CALLBACK_URL"),
		),
		github.New(
			os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
			os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
			os.Getenv("GITHUB_OAUTH_CALLBACK_URL"),
		),
	)

	server := server.New()
	server.RegisterFiberRoutes()

	done := make(chan bool, 1)

	go func() {
		port, _ := strconv.Atoi(os.Getenv("PORT"))
		err := server.Listen(fmt.Sprintf(":%d", port))
		if err != nil {
			panic(fmt.Sprintf("http server error: %s", err))
		}
	}()

	go gracefulShutdown(server, done)

	<-done
	log.Println("Graceful shutdown complete.")
}
