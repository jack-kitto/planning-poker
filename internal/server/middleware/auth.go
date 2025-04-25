package middleware

import (
	"planning-poker/internal/server/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func ProtectedMiddleware(store *session.Store) fiber.Handler {
	config, err := config.Load()
	if err != nil {
		panic(err)
	}
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Session error")
		}

		if sess.Get("user") == nil && config.AUTH_BYPASS == "false" {
			return c.Redirect("/")
		}

		return c.Next()
	}
}
