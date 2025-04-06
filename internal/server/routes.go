package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"planning-poker/cmd/web"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/session/v2"

	"github.com/shareed2k/goth_fiber"
)

var (
	store       = session.New()
	emailTokens = make(map[string]string)
)

func (s *FiberServer) RegisterFiberRoutes() {
	// Apply CORS middleware
	s.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: false, // credentials require explicit origins
		MaxAge:           300,
	}))

	// Basic routes
	s.Get("/", adaptor.HTTPHandler(templ.Handler(web.LandingPage())))
	s.Get("/login", adaptor.HTTPHandler(templ.Handler(web.LoginPage())))
	s.Get("/register", adaptor.HTTPHandler(templ.Handler(web.SignUpPage())))
	s.Get("/dashboard", s.protectedMiddleware(), s.dashboardHandler)

	// Health check
	s.Get("/health", s.healthHandler)

	// WebSocket
	s.Get("/websocket", websocket.New(s.websocketHandler))

	s.Use("/assets", filesystem.New(filesystem.Config{
		Root:       http.FS(web.Files),
		PathPrefix: "assets",
		Browse:     false,
	}))
	// OAuth routes
	s.Get("/auth/:provider", s.authHandler)
	s.Get("/auth/:provider/callback", s.authCallbackHandler)
	s.Post("/send-email", s.sendEmailHandler)
	s.Get("/verify-email/:token", s.verifyEmailHandler)
	s.Post("/resend-email", s.resendEmailHandler)

	// Logout route
	s.Get("/logout", func(c *fiber.Ctx) error {
		sess := store.Get(c)
		err := sess.Destroy()
		if err != nil {
			log.Println(err)
		}
		return c.Redirect("/")
	})
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	return c.JSON(s.db.Health())
}

func (s *FiberServer) websocketHandler(con *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			_, _, err := con.ReadMessage()
			if err != nil {
				cancel()
				log.Println("Receiver Closing", err)
				break
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			payload := fmt.Sprintf("server timestamp: %d", time.Now().UnixNano())
			if err := con.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
				log.Printf("could not write to socket: %v", err)
				return
			}
			time.Sleep(time.Second * 2)
		}
	}
}

func (s *FiberServer) authHandler(c *fiber.Ctx) error {
	err := goth_fiber.BeginAuthHandler(c)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (s *FiberServer) authCallbackHandler(c *fiber.Ctx) error {
	user, err := goth_fiber.CompleteUserAuth(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).SendString(fmt.Sprintf("Authentication failed: %v", err))
	}

	sess := store.Get(c)
	sess.Set("user", map[string]string{
		"Name":     user.Name,
		"Email":    user.Email,
		"Provider": user.Provider,
	})
	save_err := sess.Save()
	if save_err != nil {
		log.Println(save_err)
	}

	return c.Redirect("/dashboard")
}

func (s *FiberServer) protectedMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess := store.Get(c)

		if sess.Get("user") == nil {
			return c.Redirect("/")
		}

		return c.Next()
	}
}

func (s *FiberServer) dashboardHandler(c *fiber.Ctx) error {
	sess := store.Get(c)

	user := sess.Get("user")
	if user == nil {
		return c.Status(http.StatusUnauthorized).SendString("Unauthorized")
	}

	return adaptor.HTTPHandler(templ.Handler(web.DashboardPage(user.(map[string]string))))(c)
}

func generateToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (s *FiberServer) sendEmailHandler(c *fiber.Ctx) error {
	email := c.FormValue("email")
	if email == "" {
		return c.Status(http.StatusBadRequest).SendString("Email is required")
	}

	token := generateToken()
	emailTokens[token] = email

	link := fmt.Sprintf("http://localhost:3000/verify-email/%s", token)
	html := fmt.Sprintf("<p>Click <a href='%s'>here</a> to sign in.</p>", link)

	err := sendEmail(email, "Sign In Link", html)
	if err != nil {
		log.Println(err)
		return c.Status(http.StatusInternalServerError).SendString("Failed to send email")
	}

	return c.Render("email_sent", fiber.Map{
		"email": email,
	})
}

func (s *FiberServer) verifyEmailHandler(c *fiber.Ctx) error {
	token := c.Params("token")
	email, exists := emailTokens[token]
	if !exists {
		return c.Status(http.StatusUnauthorized).SendString("Invalid or expired token")
	}

	// Log the user in (store session)
	sess := store.Get(c)
	sess.Set("user", map[string]string{"Email": email})
	err := sess.Save()
	if err != nil {
		log.Println(err)
		return c.Status(http.StatusInternalServerError).SendString("Failed to log in")
	}

	// Remove the token after use
	delete(emailTokens, token)

	return c.Redirect("/dashboard")
}

func (s *FiberServer) resendEmailHandler(c *fiber.Ctx) error {
	email := c.FormValue("email")
	if email == "" {
		return c.Status(http.StatusBadRequest).SendString("Email is required")
	}

	// Generate a new token and send the email
	token := generateToken()
	emailTokens[token] = email

	link := fmt.Sprintf("http://localhost:3000/verify-email/%s", token)
	html := fmt.Sprintf("<p>Click <a href='%s'>here</a> to sign in.</p>", link)

	err := sendEmail(email, "Resend Sign In Link", html)
	if err != nil {
		log.Println(err)
		return c.Status(http.StatusInternalServerError).SendString("Failed to resend email")
	}

	return c.Render("email_sent", fiber.Map{
		"email": email,
	})
}
