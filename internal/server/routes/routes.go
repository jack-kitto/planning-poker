package routes

import (
	"net/http"
	"planning-poker/cmd/web"
	"planning-poker/cmd/web/pages"
	"planning-poker/internal/server/handlers"
	"planning-poker/internal/server/middleware"

	"github.com/a-h/templ"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

func RegisterFiberRoutes(app *fiber.App, handlers *handlers.Handlers) {
	// Apply CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: false,
		MaxAge:           300,
	}))

	app.Use("/favicons", filesystem.New(filesystem.Config{
		Root:       http.FS(web.Files),
		PathPrefix: "assets/favicons",
		Browse:     false,
	}))
	// Basic routes
	app.Get("/", adaptor.HTTPHandler(templ.Handler(pages.LandingPage())))
	app.Get("/login", adaptor.HTTPHandler(templ.Handler(pages.LoginPage())))
	app.Get("/register", adaptor.HTTPHandler(templ.Handler(pages.SignUpPage())))
	app.Get("/dashboard", middleware.ProtectedMiddleware(handlers.Store), handlers.DashboardHandler)

	// Health check
	app.Get("/health", handlers.HealthHandler)

	// WebSocket
	app.Get("/websocket", websocket.New(handlers.WebsocketHandler))

	app.Use("/assets", filesystem.New(filesystem.Config{
		Root:       http.FS(web.Files),
		PathPrefix: "assets",
		Browse:     false,
	}))

	// OAuth routes
	app.Get("/auth/:provider", handlers.AuthHandler)
	app.Get("/auth/:provider/callback", handlers.AuthCallbackHandler)
	app.Post("/send-email", handlers.SendEmailHandler)

	app.Get("/check-status", handlers.CheckAuthStatusHandler)
	app.Post("/create-account", handlers.CreateAccountSubmitHandler)
	app.Get("/create-account", handlers.CreateAccountHandler)
	app.Get("/verify-email/:token", handlers.VerifyEmailHandler)
	app.Get("/verification-success", adaptor.HTTPHandler(templ.Handler(pages.VerificationSuccessPage())))
	app.Post("/resend-email", handlers.ResendEmailHandler)

	// Logout route
	app.Get("/logout", func(c *fiber.Ctx) error {
		sess, err := handlers.Store.Get(c)
		if err != nil {
			return err
		}
		err = sess.Destroy()
		if err != nil {
			return err
		}
		return c.Redirect("/")
	})
}
