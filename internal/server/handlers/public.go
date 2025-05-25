package handlers

import (
	"log"
	"planning-poker/cmd/web/pages"
	"planning-poker/internal/server/models"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func (h *Handlers) LoginPageHandler(c *fiber.Ctx) error {
	sess, err := h.Store.Get(c)
	if err != nil {
		return adaptor.HTTPHandler(templ.Handler(pages.LandingPage()))(c)
	} else {
		log.Printf("[SessionPageHandler] found session")
	}
	userData := sess.Get("user")
	user, ok := userData.(models.User)
	if ok {
		log.Printf("[SessionPageHandler] found user in session. Hi %s", user.Name)
		return c.Redirect("/dashboard")
	}

	return adaptor.HTTPHandler(templ.Handler(pages.LoginPage()))(c)
}

func (h *Handlers) LandingPageHandler(c *fiber.Ctx) error {
	sess, err := h.Store.Get(c)
	if err != nil {
		return adaptor.HTTPHandler(templ.Handler(pages.LandingPage()))(c)
	} else {
		log.Printf("[SessionPageHandler] found session")
	}
	userData := sess.Get("user")
	user, ok := userData.(models.User)
	if ok {
		log.Printf("[SessionPageHandler] found user in session. Hi %s", user.Name)
		return c.Redirect("/dashboard")
	}

	return adaptor.HTTPHandler(templ.Handler(pages.LandingPage()))(c)
}

func (h *Handlers) VerificationSuccessHandler(c *fiber.Ctx) error {
	sess, err := h.Store.Get(c)
	if err != nil {
		return c.Redirect("/")
	} else {
		log.Printf("[SessionPageHandler] found session")
	}
	userData := sess.Get("user")
	user, ok := userData.(models.User)
	if ok {
		log.Printf("[SessionPageHandler] found user in session. Hi %s", user.Name)
		return adaptor.HTTPHandler(templ.Handler(pages.VerificationSuccessPage()))(c)
	}

	return c.Redirect("/dashboard")
}
