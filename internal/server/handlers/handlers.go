package handlers

import (
	"planning-poker/internal/database"
	"planning-poker/internal/server/config"

	"github.com/gofiber/fiber/v2/middleware/session"
)

type Handlers struct {
	DB     database.Service
	Store  *session.Store
	Config *config.Config
}

func NewHandlers(db database.Service, store *session.Store) *Handlers {
	cfg, e := config.Load()
	if e != nil {
		panic(e)
	}
	return &Handlers{
		DB:     db,
		Store:  store,
		Config: cfg,
	}
}
