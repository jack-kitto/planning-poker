package server

import (
	"planning-poker/internal/database"
	"planning-poker/internal/server/config"
	"planning-poker/internal/server/handlers"
	"planning-poker/internal/server/routes"

	"github.com/gofiber/fiber/v2"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "planning-poker",
			AppName:      "planning-poker",
		}),

		db: database.New(),
	}

	// Initialize handlers with dependencies
	handlers := handlers.NewHandlers(server.db, config.Store)

	// Register routes
	routes.RegisterFiberRoutes(server.App, handlers)

	return server
}
