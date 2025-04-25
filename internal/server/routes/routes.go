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
	// ==========================================
	// Middleware Configuration
	// ==========================================

	// Apply CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// ==========================================
	// Static Assets
	// ==========================================

	// Serve favicon files
	app.Use("/favicons", filesystem.New(filesystem.Config{
		Root:       http.FS(web.Files),
		PathPrefix: "assets/favicons",
		Browse:     false,
	}))

	// Serve static assets
	app.Use("/assets", filesystem.New(filesystem.Config{
		Root:       http.FS(web.Files),
		PathPrefix: "assets",
		Browse:     false,
	}))

	// ==========================================
	// Public Pages
	// ==========================================

	// Landing and login pages
	app.Get("/", adaptor.HTTPHandler(templ.Handler(pages.LandingPage())))
	app.Get("/login", adaptor.HTTPHandler(templ.Handler(pages.LoginPage())))
	app.Get("/verification-success", adaptor.HTTPHandler(templ.Handler(pages.VerificationSuccessPage())))

	// Health check endpoint
	app.Get("/health", handlers.HealthHandler)

	// ==========================================
	// Authentication & User Management
	// ==========================================

	// OAuth routes
	app.Get("/auth/:provider", handlers.AuthHandler)
	app.Get("/auth/:provider/callback", handlers.AuthCallbackHandler)

	// Email verification
	app.Post("/send-email", handlers.SendEmailHandler)
	app.Get("/check-status", handlers.CheckAuthStatusHandler)
	app.Get("/verify-email/:token", handlers.VerifyEmailHandler)
	app.Post("/resend-email", handlers.ResendEmailHandler)

	// Account creation and management
	app.Get("/create-account", handlers.CreateAccountHandler)
	app.Post("/create-account", handlers.CreateAccountSubmitHandler)

	// User profile
	app.Post("/user", handlers.CreateUserHandler)
	app.Patch("/user", handlers.UpdateUserHandler)
	app.Get("/user", handlers.GetUserHandler)

	// Logout
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

	// Dashboard
	app.Get("/dashboard", middleware.ProtectedMiddleware(handlers.Store), handlers.DashboardHandler)

	// ==========================================
	// Session Management
	// ==========================================

	// Session creation
	app.Post("/create-session", middleware.ProtectedMiddleware(handlers.Store), handlers.CreateSessionHandler)
	app.Post("/test-create-session", handlers.TestSessionHandler)

	// Session viewing and editing
	app.Get("/session/:id", middleware.ProtectedMiddleware(handlers.Store), handlers.SessionPageHandler)
	app.Get("/test/session/:id", middleware.ProtectedMiddleware(handlers.Store), handlers.SessionPageHandler)
	app.Get("/session/:id/edit-title-form", middleware.ProtectedMiddleware(handlers.Store), handlers.EditSessionTitleFormHandler)
	app.Post("/session/:id/edit-title", middleware.ProtectedMiddleware(handlers.Store), handlers.EditSessionNameHandler)
	app.Get("/session/:id/title", middleware.ProtectedMiddleware(handlers.Store), handlers.SessionTitleHandler)

	// ==========================================
	// Story Management
	// ==========================================

	// Story creation and editing
	app.Get("/session/:id/story/create", middleware.ProtectedMiddleware(handlers.Store), handlers.StoryCreatePopupHandler)
	app.Post("/session/story/create", middleware.ProtectedMiddleware(handlers.Store), handlers.CreateStoryHandler)
	app.Get("/session/story/edit/:id", middleware.ProtectedMiddleware(handlers.Store), handlers.StoryEditPopupHandler)
	app.Post("/session/story/edit", middleware.ProtectedMiddleware(handlers.Store), handlers.UpdateStoryHandler)

	// ==========================================
	// API Endpoints
	// ==========================================

	// Data retrieval endpoints
	app.Get("api/session/:id", middleware.ProtectedMiddleware(handlers.Store), handlers.GetSessionHandler)
	app.Get("api/story/:id", middleware.ProtectedMiddleware(handlers.Store), handlers.GetUserStoryHandler)

	// ==========================================
	// WebSocket Connection
	// ==========================================

	// Real-time communication
	app.Get("/websocket", websocket.New(handlers.WebsocketHandler))
}
