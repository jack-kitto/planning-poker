package server

import (
	"encoding/gob"
	"planning-poker/internal/database"
	"planning-poker/internal/server/handlers"
	"planning-poker/internal/server/models"
	"planning-poker/internal/server/routes"
	"planning-poker/internal/server/session"

	"github.com/gofiber/fiber/v2"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	gob.Register(models.User{})
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "planning-poker",
			AppName:      "planning-poker",
		}),

		db: database.New(),
	}

	// Initialize handlers with dependencies
	handlers := handlers.NewHandlers(server.db, session.Store)

	// setup oath providers
	setupProviders(handlers)

	// Register routes
	routes.RegisterFiberRoutes(server.App, handlers)

	return server
}
