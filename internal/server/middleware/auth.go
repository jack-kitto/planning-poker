package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func ProtectedMiddleware(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Session error")
		}

		if sess.Get("user") == nil {
			return c.Redirect("/")
		}

		return c.Next()
	}
}
