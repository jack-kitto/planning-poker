package handlers

import (
	"planning-poker/internal/database"
	"planning-poker/internal/server/config"

	"github.com/gofiber/session/v2"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

type Handlers struct {
	DB     database.Service
	Store  *session.Session
	Config *config.Config
}

func NewHandlers(db database.Service, store *session.Session) *Handlers {
	cfg, e := config.Load()
	goth.UseProviders(
		google.New(
			cfg.GoogleOAuthClientID,
			cfg.GoogleOAuthClientSecret,
			cfg.GoogleOAuthClientSecret,
		),
		github.New(
			cfg.GitHubOAuthClientID,
			cfg.GitHubOAuthClientSecret,
			cfg.GitHubOAuthCallbackUrl,
		),
	)
	if e != nil {
		panic(e)
	}
	return &Handlers{
		DB:     db,
		Store:  store,
		Config: cfg,
	}
}
