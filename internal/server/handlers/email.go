package handlers

import (
	"fmt"
	"log"
	"net/http"
	"planning-poker/cmd/web"
	"planning-poker/internal/server/config"
	"planning-poker/internal/server/utils"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/resend/resend-go/v2"
)

// generateEmailTemplate generates the HTML email template with the provided link.
func generateEmailTemplate(link string) string {
	year := time.Now().Year()
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Your Magic Login Link</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { max-width: 150px; margin-bottom: 20px; }
        .button { 
            display: inline-block; 
            padding: 12px 24px; 
            background-color: #4F46E5; 
            color: white; 
            text-decoration: none; 
            border-radius: 4px; 
            font-weight: bold; 
            margin: 20px 0;
        }
        .footer { 
            margin-top: 30px; 
            font-size: 12px; 
            color: #666; 
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Welcome to Planning Poker!</h1>
    </div>
    
    <p>Hello,</p>
    
    <p>We received a request to sign in to Planning Poker using this email address. Click the button below to sign in.</p>
    
    <div style="text-align: center;">
        <a href="%s" class="button">Sign In Now</a>
    </div>
    
    <p>If you didn't request this link, you can safely ignore this email.</p>
    
    <div class="footer">
        <p>This link will expire in 24 hours and can only be used once.</p>
        <p>© %d Planning Poker. All rights reserved.</p>
    </div>
</body>
</html>
`, link, year)
}

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

	link := fmt.Sprintf("%s/verify-email/%s", h.Config.BaseURL, token)

	// Send the email
	err := h.sendEmail(email, "Verify Your Email", link)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to send email")
	}

	return adaptor.HTTPHandler(templ.Handler(web.EmailSentPage(email)))(c)
}
