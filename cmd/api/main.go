package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"planning-poker/internal/database"
	"planning-poker/internal/server"
	"planning-poker/internal/server/seed"
	"strconv"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
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
	server := server.New()
	done := make(chan bool, 1)
	seedFlag := flag.Bool("seed", false, "Clean and seed the database")
	flag.Parse()

	if *seedFlag {
		log.Println("Seeding database...")
		db := database.BunDB()
		if err := seed.Seed(db); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		log.Println("Seeding complete.")
		return
	}

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
