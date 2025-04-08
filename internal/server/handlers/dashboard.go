package handlers

import (
	"net/http"
	"planning-poker/cmd/web"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func (h *Handlers) DashboardHandler(c *fiber.Ctx) error {
	sess := h.Store.Get(c)

	user := sess.Get("user")
	if user == nil {
		return c.Status(http.StatusUnauthorized).SendString("Unauthorized")
	}

	return adaptor.HTTPHandler(templ.Handler(web.DashboardPage(user.(map[string]string))))(c)
}
