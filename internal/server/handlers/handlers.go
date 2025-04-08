package handlers

import (
	"planning-poker/internal/database"

	"github.com/gofiber/session/v2"
)

type Handlers struct {
	DB    database.Service
	Store *session.Session
}

func NewHandlers(db database.Service, store *session.Session) *Handlers {
	return &Handlers{
		DB:    db,
		Store: store,
	}
}
