package handlers

import (
	"fmt"
	"log"
	"net/http"
	"planning-poker/cmd/web/pages"
	"planning-poker/internal/server/models"
	"planning-poker/internal/server/session"
	"planning-poker/internal/server/utils"

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
	user := models.SessionUser{Email: email}
	sess.Set("user", user)
	if email == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Email is required")
	}

	token := utils.GenerateToken()
	session.EmailTokens[token] = email

	link := fmt.Sprintf("%s/verify-email/%s", h.Config.BaseURL, token)

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
	email, exists := session.EmailTokens[token]
	if !exists {
		return c.Status(http.StatusUnauthorized).SendString("Invalid or expired token")
	}

	sess, err := h.Store.Get(c)
	if err != nil {
		return err
	}
	user := sess.Get("user")
	sessionUser := models.SessionUser{Email: email}
	if existing, ok := user.(models.SessionUser); ok {
		sessionUser.Name = existing.Name
	}

	sess.Set("user", sessionUser)
	err = sess.Save()
	if err != nil {
		log.Println(err)
		return c.Status(http.StatusInternalServerError).SendString("Failed to log in")
	}

	delete(session.EmailTokens, token)

	return c.Redirect("/create-account")
}

func (h *Handlers) ResendEmailHandler(c *fiber.Ctx) error {
	email := c.FormValue("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Email is required")
	}

	token := utils.GenerateToken()
	session.EmailTokens[token] = email // Consider thread-safety here if concurrent access is possible.

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
