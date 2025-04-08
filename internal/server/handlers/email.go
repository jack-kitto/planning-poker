package handlers

import (
	"fmt"
	"log"
	"net/http"
	"planning-poker/cmd/web"
	"planning-poker/internal/server/config"
	"planning-poker/internal/server/utils"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func (h *Handlers) SendEmailHandler(c *fiber.Ctx) error {
	email := c.FormValue("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Email is required")
	}

	token := utils.GenerateToken()
	config.EmailTokens[token] = email // Consider thread-safety here if concurrent access is possible.

	link := fmt.Sprintf("http://%s/verify-email/%s", h.Config.BaseURL, token)
	html := fmt.Sprintf("<p>Click <a href='%s'>here</a> to sign in.</p>", link)

	if err := utils.SendEmail(email, "Sign In Link", html); err != nil {
		log.Printf("failed to send email to %s: %v", email, err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to send email")
	}

	return adaptor.HTTPHandler(templ.Handler(web.EmailSentPage(email)))(c)
}

func (h *Handlers) VerifyEmailHandler(c *fiber.Ctx) error {
	token := c.Params("token")
	email, exists := config.EmailTokens[token]
	if !exists {
		return c.Status(http.StatusUnauthorized).SendString("Invalid or expired token")
	}

	sess := h.Store.Get(c)
	sess.Set("user", map[string]string{"Email": email})
	err := sess.Save()
	if err != nil {
		log.Println(err)
		return c.Status(http.StatusInternalServerError).SendString("Failed to log in")
	}

	delete(config.EmailTokens, token)

	return c.Redirect("/dashboard")
}

func (h *Handlers) ResendEmailHandler(c *fiber.Ctx) error {
	email := c.FormValue("email")
	if email == "" {
		return c.Status(http.StatusBadRequest).SendString("Email is required")
	}

	token := utils.GenerateToken()
	config.EmailTokens[token] = email

	link := fmt.Sprintf("http://%s/verify-email/%s", h.Config.BaseURL, token)
	html := fmt.Sprintf("<p>Click <a href='%s'>here</a> to sign in.</p>", link)

	err := utils.SendEmail(email, "Resend Sign In Link", html)
	if err != nil {
		log.Println(err)
		return c.Status(http.StatusInternalServerError).SendString("Failed to resend email")
	}

	return adaptor.HTTPHandler(templ.Handler(web.EmailSentPage(email)))(c)
}
