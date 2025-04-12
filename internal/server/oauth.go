package server

import (
	"planning-poker/internal/server/handlers"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

func setupProviders(handlers *handlers.Handlers) {
	goth.UseProviders(
		google.New(
			handlers.Config.GoogleOAuthClientID,
			handlers.Config.GoogleOAuthClientSecret,
			handlers.Config.GoogleOAuthClientSecret,
		),
		github.New(
			handlers.Config.GitHubOAuthClientID,
			handlers.Config.GitHubOAuthClientSecret,
			handlers.Config.GitHubOAuthCallbackUrl,
		),
	)
}
