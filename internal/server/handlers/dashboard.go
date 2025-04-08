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

	// Convert the user map to map[string]string
	userMap, ok := user.(map[string]interface{})
	if !ok {
		return c.Status(http.StatusInternalServerError).SendString("Invalid user data")
	}

	convertedUser := make(map[string]string)
	for key, value := range userMap {
		strValue, ok := value.(string)
		if !ok {
			return c.Status(http.StatusInternalServerError).SendString("Invalid user data")
		}
		convertedUser[key] = strValue
	}

	return adaptor.HTTPHandler(templ.Handler(web.DashboardPage(convertedUser)))(c)
}
