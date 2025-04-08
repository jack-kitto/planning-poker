// internal/server/routes/routes.go
package routes

import (
	"log"
	"net/http"
	"planning-poker/cmd/web"
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

	// Basic routes
	app.Get("/", adaptor.HTTPHandler(templ.Handler(web.LandingPage())))
	app.Get("/login", adaptor.HTTPHandler(templ.Handler(web.LoginPage())))
	app.Get("/register", adaptor.HTTPHandler(templ.Handler(web.SignUpPage())))
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
	app.Get("/verify-email/:token", handlers.VerifyEmailHandler)
	app.Post("/resend-email", handlers.ResendEmailHandler)

	// Logout route
	app.Get("/logout", func(c *fiber.Ctx) error {
		sess := handlers.Store.Get(c)
		err := sess.Destroy()
		if err != nil {
			log.Println(err)
		}
		return c.Redirect("/")
	})
}
