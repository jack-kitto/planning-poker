package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/session/v2"
)

func ProtectedMiddleware(store *session.Session) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess := store.Get(c)

		if sess.Get("user") == nil {
			return c.Redirect("/")
		}

		return c.Next()
	}
}
