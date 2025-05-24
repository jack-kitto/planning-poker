package handlers

import (
	"net/http"
	"planning-poker/cmd/web/pages"
	"planning-poker/internal/server/models"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func (h *Handlers) DashboardHandler(c *fiber.Ctx) error {
	sess, err := h.Store.Get(c)
	if err != nil {
		return err
	}

	user := sess.Get("user")
	if user == nil {
		return c.Redirect("/login")
	}

	sessionUser, ok := user.(models.User)
	if !ok {
		return c.Status(http.StatusInternalServerError).SendString("Invalid user data")
	}
	sessions, err := h.DB.GetSessionsForUser(sessionUser.ID)
	if err != nil {
		return err
	}
	return adaptor.HTTPHandler(templ.Handler(pages.DashboardPage(&sessionUser, sessions)))(c)
}
