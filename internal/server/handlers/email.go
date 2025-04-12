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
		log.Printf("Session error in VerifyEmailHandler: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Could not process verification. Please try again.")
	}

	var sessionUser models.SessionUser
	user := sess.Get("user")
	if existing, ok := user.(models.SessionUser); ok {
		sessionUser = existing
	}
	sessionUser.Email = email

	sess.Set("user", sessionUser)
	if err := sess.Save(); err != nil {
		log.Printf("Failed to save session in VerifyEmailHandler: %v", err)
		// Render an error page or generic message
		return c.Status(fiber.StatusInternalServerError).SendString("Could not complete sign in. Please try again.")
	}
	delete(session.EmailTokens, token)
	return c.Redirect("/verification-success")
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

func (h *Handlers) CheckAuthStatusHandler(c *fiber.Ctx) error {
	sess, err := h.Store.Get(c)
	if err != nil {
		log.Printf("Session error in CheckAuthStatusHandler: %v", err)
		return c.SendStatus(fiber.StatusOK)
	}

	user := sess.Get("user")
	sessionUser, ok := user.(models.SessionUser)

	if ok && sessionUser.Email != "" {
		var redirectURL string
		if sessionUser.Name != "" {
			redirectURL = "/dashboard"
		} else {
			redirectURL = "/create-account"
		}

		c.Set("HX-Redirect", redirectURL)
		log.Printf("CheckAuthStatus: User %s verified/authenticated. Redirecting via HX-Redirect to %s", sessionUser.Email, redirectURL)
		return c.SendStatus(fiber.StatusOK)
	}
	log.Printf("CheckAuthStatus: Session status is pending verification.")
	return c.SendStatus(fiber.StatusOK)
}
