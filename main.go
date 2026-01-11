package main

import (
	"os"
	"time"

	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/field"
	"github.com/rocky-ads/site/handler"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/service/sms"
	"github.com/sasha-s/go-deadlock"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	flogger "github.com/gofiber/fiber/v2/middleware/logger"
)

// setupApp configures the Fiber app with middleware and routes
func setupApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: handler.ErrorHandler,
		BodyLimit:    config.ServerBodyLimit, // Total request body size (for multiple files)
		ReadTimeout:  30 * time.Second,       // Prevent long-running requests
		//WriteTimeout: 30 * time.Second,       // Prevent long-running responses
	})

	// Must be early in middleware chain
	app.Use(handler.ConfigureHelmet())

	app.Use(limiter.New(limiter.Config{
		Max:        config.ServerRateLimitMax,
		Expiration: config.ServerRateLimitExp,
	}))

	app.Use(handler.JWTMiddleware)
	app.Use(handler.CSRFMiddleware)

	app.Use(flogger.New(flogger.Config{
		Output: logger.Writer(),
		Format: "${status} | ${latency} | ${ip} | ${method} | ${path}\n",
	}))

	// Public routes

	app.Static("/", "./static")

	app.Get("/", handler.HomeHandler)
	app.Get("/login", handler.LoginHandler)
	app.Get("/logout", handler.LogoutHandler)
	app.Get("/register", handler.RegisterHandler)
	app.Get("/health", handler.HandleHealth)
	app.Get("/ad/:id", handler.AdHandler)
	app.Get("/ad/:id/image/:index/:size", handler.ImageHandler)

	// Auth routes

	auth := app.Group("/auth", handler.AuthRequired)

	auth.Get("/sse", handler.SSEHandler)

	auth.Get("/ad/new", handler.NewAdHandler)
	auth.Delete("/ad/:id/delete", handler.DeleteAdHandler)
	auth.Post("/ad/:id/restore", handler.RestoreAdHandler)
	auth.Get("/ad/:id/new-conversation", handler.MessageModalHandler)

	auth.Get("/user/menu", handler.UserMenuHandler)
	auth.Get("/user/myads", handler.UserMyAdsHandler)
	auth.Get("/user/myads/tab/:tab", handler.UserMyAdsTabHandler)
	auth.Get("/user/messages", handler.UserMessagesHandler)
	auth.Get("/user/settings", handler.UserSettingsHandler)
	auth.Get("/user/about", handler.UserAboutHandler)
	auth.Get("/welcome", handler.WelcomeHandler)

	auth.Get("/conversation/:id", handler.ConversationModalHandler)
	auth.Post("/conversation/:id/send", handler.SendConversationMessageHandler)
	auth.Post("/conversation/:id/rock/throw", handler.ThrowRockHandler)
	auth.Delete("/conversation/:id/rock/unthrow", handler.UnthrowRockHandler)

	auth.Post("/bookmark/:id", handler.BookmarkHandler)
	auth.Delete("/bookmark/:id", handler.BookmarkHandler)

	// Admin routes

	admin := app.Group("/admin", handler.AdminRequired)
	admin.Get("/dashboard", handler.AdminDashboardHandler)
	admin.Get("/tab/:tab", handler.AdminTabHandler)
	admin.Get("/users", handler.AdminUsersHandler)
	admin.Post("/user/:id/delete", handler.AdminUserDeleteHandler)
	admin.Post("/user/:id/restore", handler.AdminUserRestoreHandler)
	admin.Post("/user/:id/promote", handler.AdminUserPromoteHandler)
	admin.Post("/user/:id/demote", handler.AdminUserDemoteHandler)

	// API routes

	api := app.Group("/api")

	api.Post("/login", handler.LoginSubmitHandler)
	api.Post("/register/step1", handler.RegistrationRateLimiter, handler.RegisterStep1Handler)
	api.Post("/register/step2", handler.RegisterStep2Handler)
	api.Post("/register/step3", handler.RegisterStep3Handler)
	api.Post("/sms/webhook", handler.SMSWebhookHandler)
	api.Get("/view/:view", handler.ViewHandler)
	api.Get("/image-nav/:id", handler.ImageNavigationHandler)
	api.Get("/image-full/:id", handler.ImageFullScreenHandler)
	api.Get("/show-filters", handler.ShowFiltersHandler)
	api.Get("/category-select", handler.CategorySelectHandler)
	api.Get("/modal-remove/:name", handler.ModalRemoveHandler)
	api.Get("/search/", handler.SearchPageHandler)
	api.Get("/ad/:id/share", handler.AdShareHandler)
	api.Get("/ad/share/copy", handler.AdShareCopyHandler)

	categoryRouter := api.Group("/category/:category")
	categoryRouter.Get("/values/:field", handler.GetAllValuesHandler)
	categoryRouter.Get("/any-values/:field", handler.GetAnyValuesHandler)
	categoryRouter.Post("/ad-values/:field", handler.GetAdValuesHandler)
	categoryRouter.Get("/chains", handler.GetChainsHandler)
	categoryRouter.Get("/first-spec-fields", handler.GetFirstSpecFieldsHandler)
	categoryRouter.Get("/last-spec-field", handler.GetLastSpecFieldHandler)
	categoryRouter.Post("/search", handler.SearchHandler)
	categoryRouter.Get("/switch", handler.SwitchCategoryHandler)

	api.Get("/ads/:id/filter-values", handler.GetAdFilterValuesHandler)

	return app
}

func main() {
	deadlock.Opts.Disable = false                    // default = false
	deadlock.Opts.DeadlockTimeout = 10 * time.Second // detect very long waits
	deadlock.Opts.PrintAllCurrentGoroutines = true   // very helpful
	deadlock.Opts.LogBuf = os.Stderr

	if err := logger.Init(config.LogLevel, config.LogFormat,
		config.LogFile); err != nil {
		logger.Fatal("Failed to initialize logger", "error", err)
	}

	config.SecurityCheck()

	if err := db.Init("project.db"); err != nil {
		logger.Fatal("Failed to open database", "error", err)
	}
	defer db.Close()

	if err := field.Init(); err != nil {
		logger.Fatal("Failed to initialize fields", "error", err)
	}

	if err := ad.Init(); err != nil {
		logger.Fatal("Failed to initialize ads", "error", err)
	}

	sms.StartSMSWorker()

	app := setupApp()

	port := ":" + config.ServerPort
	logger.Info("Server starting", "port", port)
	logger.Info("API endpoints:")
	logger.Info("  GET  /api/category/:category/values/:field")
	logger.Info("  GET  /api/category/:category/any-values/:field")
	logger.Info("  POST /api/category/:category/ad-values/:field")
	logger.Info("  GET  /api/category/:category/chains")
	logger.Info("  GET  /api/category/:category/first-spec-fields")
	logger.Info("  GET  /api/category/:category/last-spec-field")
	logger.Info("  POST /api/category/:category/search")
	logger.Info("  GET  /api/ads/:id/filter-values")

	if err := app.Listen(port); err != nil {
		logger.Fatal("Server failed to start", "error", err)
	}
}
