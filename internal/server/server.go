package server

import (
	"github.com/gofiber/fiber/v2"

	"planning-poker/internal/database"
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

	return server
}
