package handlers

import (
	"fmt"
	"log"
	"net/http"
	"planning-poker/internal/server/config"
	"planning-poker/internal/server/utils"

	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) SendEmailHandler(c *fiber.Ctx) error {
	email := c.FormValue("email")
	if email == "" {
		return c.Status(http.StatusBadRequest).SendString("Email is required")
	}

	token := utils.GenerateToken()
	config.EmailTokens[token] = email

	link := fmt.Sprintf("http://localhost:3000/verify-email/%s", token)
	html := fmt.Sprintf("<p>Click <a href='%s'>here</a> to sign in.</p>", link)

	err := utils.SendEmail(email, "Sign In Link", html)
	if err != nil {
		log.Println(err)
		return c.Status(http.StatusInternalServerError).SendString("Failed to send email")
	}

	return c.Render("email_sent", fiber.Map{
		"email": email,
	})
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

	link := fmt.Sprintf("http://localhost:3000/verify-email/%s", token)
	html := fmt.Sprintf("<p>Click <a href='%s'>here</a> to sign in.</p>", link)

	err := utils.SendEmail(email, "Resend Sign In Link", html)
	if err != nil {
		log.Println(err)
		return c.Status(http.StatusInternalServerError).SendString("Failed to resend email")
	}

	return c.Render("email_sent", fiber.Map{
		"email": email,
	})
}
