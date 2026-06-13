package main

import (
	"os"
	"time"

	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/handler"
	"github.com/rocky-ads/site/internal/imagestore"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/service/sms"
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
	app.Get("/error", handler.ErrorPageHandler)
	app.Get("/about", handler.AboutHandler)
	app.Get("/faq", handler.FAQHandler)
	app.Get("/faq/:section", handler.FAQHandler)
	app.Get("/ad/:id", handler.AdHandler)
	app.Get("/ad/:id/image/:index/:size", handler.ImageHandler)

	// Auth routes

	auth := app.Group("/auth", handler.AuthRequired)

	auth.Get("/sse", handler.SSEHandler)

	auth.Get("/ad/new", handler.NewAdHandler)
	auth.Get("/ad/new/price-field", handler.NewAdPriceFieldHandler)
	auth.Get("/ad/new/suggestions", handler.SuggestionsHandler)
	auth.Post("/ad/new", handler.CreateAdHandler)
	auth.Get("/ad/:id/edit", handler.EditAdHandler)
	auth.Post("/ad/:id/edit", handler.UpdateAdHandler)
	auth.Get("/ad/:id/edit/suggestions", handler.EditSuggestionsHandler)
	auth.Delete("/ad/:id/delete", handler.DeleteAdHandler)
	auth.Post("/ad/:id/restore", handler.RestoreAdHandler)
	auth.Get("/ad/:id/new-conversation", handler.MessageModalHandler)
	auth.Post("/ad/:id/send", handler.SendMessageHandler)
	auth.Get("/ad/:id/egg/:ordinal", handler.AdEggConversationHandler)

	auth.Get("/user/menu", handler.UserMenuHandler)
	auth.Get("/user/myads", handler.UserMyAdsHandler)
	auth.Get("/user/myads/tab/:tab", handler.UserMyAdsTabHandler)
	auth.Get("/user/messages", handler.UserMessagesHandler)
	auth.Get("/user/settings", handler.UserSettingsHandler)
	auth.Post("/user/settings/notifications", handler.NotificationsToggleHandler)
	auth.Post("/user/settings/password", handler.ChangePasswordHandler)
	auth.Post("/user/settings/delete", handler.DeleteAccountHandler)
	auth.Get("/user/:id", handler.UserProfileHandler)
	auth.Get("/user/:id/summary", handler.UserSummaryHandler)
	auth.Get("/user/:id/egg/:ordinal", handler.UserEggConversationHandler)
	auth.Get("/welcome", handler.WelcomeHandler)

	auth.Get("/conversation/:id", handler.ConversationModalHandler)
	auth.Get("/conversation/:id/egg-opinion", handler.EggOpinionHandler)
	auth.Post("/conversation/:id/send", handler.SendConversationMessageHandler)
	auth.Post("/conversation/:id/egg/throw", handler.ThrowEggHandler)
	auth.Delete("/conversation/:id/egg/unthrow", handler.UnthrowEggHandler)

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
	api.Get("/hide-filters", handler.HideFiltersHandler)
	api.Get("/category-select", handler.CategorySelectHandler)
	api.Get("/modal-remove/:name", handler.ModalRemoveHandler)
	api.Get("/search/", handler.SearchPageHandler)
	api.Get("/ad/:id/share", handler.AdShareHandler)
	api.Get("/ad/share/copy", handler.AdShareCopyHandler)

	api.Get("/category/:category/switch", handler.SwitchCategoryHandler)

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

	if err := db.Init(config.DatabaseURL); err != nil {
		logger.Fatal("Failed to open database", "error", err)
	}
	defer db.Close()
	host, database := db.ConnectionTarget(config.DatabaseURL)
	logger.Info("Database connected", "host", host, "database", database)

	if err := db.CheckSchema(); err != nil {
		logger.Fatal("Database not ready", "error", err)
	}

	if err := ad.LoadCategories(); err != nil {
		logger.Fatal("Failed to initialize ads", "error", err)
	}

	if err := sms.Init(); err != nil {
		logger.Fatal("Failed to initialize SMS service", "error", err)
	}
	sms.StartSMSWorker()

	imageStore, err := imagestore.NewDefault()
	if err != nil {
		logger.Fatal("Failed to initialize image store", "error", err)
	}
	handler.SetAdImageStore(imageStore)
	logger.Info("Image store configured",
		"bucket", config.MinIOBucketName, "url", config.MinIOAPIURL)

	app := setupApp()

	port := ":" + config.ServerPort
	logger.Info("Server starting", "port", port)

	if err := app.Listen(port); err != nil {
		logger.Fatal("Server failed to start", "error", err)
	}
}
