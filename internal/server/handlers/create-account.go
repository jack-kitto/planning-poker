package handlers

import (
	"fmt"
	"planning-poker/cmd/web/pages"
	"planning-poker/internal/server/models"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func (h *Handlers) CreateAccountHandler(c *fiber.Ctx) error {
	sess, err := h.Store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Session error")
	}

	user := sess.Get("user")
	sessionUser, ok := user.(models.User)
	if !ok || sessionUser.Email == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No email in session")
	}

	return adaptor.HTTPHandler(templ.Handler(pages.CreateAccount(sessionUser.Email)))(c)
}

func (h *Handlers) CreateAccountSubmitHandler(c *fiber.Ctx) error {
	name := c.FormValue("name")
	sess, err := h.Store.Get(c)
	if err != nil {
		return err
	}
	user := sess.Get("user")
	if user == nil {
		return nil
	}
	existing, ok := user.(models.User)
	if ok {
		existing.Name = name
		fmt.Println("Setting name to:")
		fmt.Println(name)
		sess.Set("user", existing)
		if err := sess.Save(); err != nil {
			return err
		}
	} else {
		fmt.Println("No existing user found")
	}
	return c.Redirect("/dashboard")
}
