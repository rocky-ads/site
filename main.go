package main

import (
	"net/http"

	"github.com/rocky-ads/site/cache"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/handlers"
	"github.com/rocky-ads/site/logger"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize logger
	if err := logger.Init("info", "text", ""); err != nil {
		logger.Fatal("Failed to initialize logger", "error", err)
	}

	// Open existing database (assumes it's already been built with cmd/rebuild_db)
	if err := db.Init("project.db"); err != nil {
		logger.Fatal("Failed to open database", "error", err)
	}
	defer db.Close()

	// Initialize caches
	logger.Info("Initializing caches...")
	if err := cache.Init(); err != nil {
		logger.Fatal("Failed to initialize caches", "error", err)
	}
	logger.Info("Caches initialized successfully")

	// Set up router
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Category-specific routes
	categoryRouter := api.PathPrefix("/categories/{category}").Subrouter()
	categoryRouter.HandleFunc("/values/{field}", handlers.GetAllValuesHandler).Methods("GET")
	categoryRouter.HandleFunc("/any-values/{field}", handlers.GetAnyValuesHandler).Methods("GET")
	categoryRouter.HandleFunc("/ad-values/{field}", handlers.GetAdValuesHandler).Methods("POST")
	categoryRouter.HandleFunc("/chains", handlers.GetChainsHandler).Methods("GET")
	categoryRouter.HandleFunc("/first-spec-fields", handlers.GetFirstSpecFieldsHandler).Methods("GET")
	categoryRouter.HandleFunc("/last-spec-field", handlers.GetLastSpecFieldHandler).Methods("GET")
	categoryRouter.HandleFunc("/search", handlers.SearchHandler).Methods("POST")

	// Ad routes
	api.HandleFunc("/ads/{id}/filter-values", handlers.GetAdFilterValuesHandler).Methods("GET")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Start server
	port := ":8080"
	logger.Info("Server starting", "port", port)
	logger.Info("API endpoints:")
	logger.Info("  GET  /api/categories/:category/values/:field")
	logger.Info("  GET  /api/categories/:category/any-values/:field")
	logger.Info("  POST /api/categories/:category/ad-values/:field")
	logger.Info("  GET  /api/categories/:category/chains")
	logger.Info("  GET  /api/categories/:category/first-spec-fields")
	logger.Info("  GET  /api/categories/:category/last-spec-field")
	logger.Info("  POST /api/categories/:category/search")
	logger.Info("  GET  /api/ads/:id/filter-values")

	if err := http.ListenAndServe(port, r); err != nil {
		logger.Fatal("Server failed to start", "error", err)
	}
}
