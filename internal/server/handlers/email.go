package handlers

import (
	"fmt"
	"log"
	"net/http"
	"planning-poker/cmd/web/pages"
	"planning-poker/internal/server/config"
	"planning-poker/internal/server/utils"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/resend/resend-go/v2"
)

// sendEmail sends an email using the Resend API.
func (h *Handlers) sendEmail(to, subject, link string) error {
	client := resend.NewClient(h.Config.EmailServerPassword)

	html := generateEmailTemplate(link)

	params := &resend.SendEmailRequest{
		From:    h.Config.EmailFrom,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	return nil
}

func (h *Handlers) SendEmailHandler(c *fiber.Ctx) error {
	email := c.FormValue("email")
	sess, err := h.Store.Get(c)
	if err != nil {
		return err
	}
	if email == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Email is required")
	}

	token := utils.GenerateToken()
	config.EmailTokens[token] = email // Consider thread-safety here if concurrent access is possible.

	link := fmt.Sprintf("%s/verify-email/%s", h.Config.BaseURL, token)

	// Send the email asynchronously
	go func() {
		err := h.sendEmail(email, "Verify Your Email", link)
		if err != nil {
			log.Printf("Failed to send email: %v", err)
		}
	}()

	return adaptor.HTTPHandler(templ.Handler(pages.EmailSentPage(email)))(c)
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
		return c.Status(fiber.StatusBadRequest).SendString("Email is required")
	}

	token := utils.GenerateToken()
	config.EmailTokens[token] = email // Consider thread-safety here if concurrent access is possible.

	link := fmt.Sprintf("%s/verify-email/%s", h.Config.BaseURL, token)

	// Send the email asynchronously
	go func() {
		err := h.sendEmail(email, "Verify Your Email", link)
		if err != nil {
			log.Printf("Failed to send email: %v", err)
		}
	}()

	return adaptor.HTTPHandler(templ.Handler(pages.EmailSentPage(email)))(c)
}
