package handlers

import (
	"log"
	"net/http"
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
	userData := sess.Get("user")
	user, ok := userData.(models.User)
	if !ok || user.Email == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No email in session")
	}
	_, err = h.DB.UpdateUser(name, user.Email)
	if err != nil {
		return c.Status(http.StatusUnauthorized).SendString("Invalid or expired token")
	}
	user.Name = name
	sess.Set("user", user)
	if err := sess.Save(); err != nil {
		log.Printf("Failed to save session in VerifyEmailHandler: %v", err)
		// Render an error page or generic message
		return c.Status(fiber.StatusInternalServerError).SendString("Could not complete sign up. Please try again.")
	}
	return c.Redirect("/dashboard")
}
