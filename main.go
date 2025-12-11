package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rocky-ads/site/cache"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/handlers"

	"github.com/gorilla/mux"
)

func main() {
	// Open existing database (assumes it's already been built with cmd/rebuild_db)
	if err := db.Init("project.db"); err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize caches
	fmt.Println("Initializing caches...")
	if err := cache.Init(); err != nil {
		log.Fatalf("Failed to initialize caches: %v", err)
	}
	fmt.Println("Caches initialized successfully")

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
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Println("API endpoints:")
	fmt.Println("  GET  /api/categories/:category/values/:field")
	fmt.Println("  GET  /api/categories/:category/any-values/:field")
	fmt.Println("  POST /api/categories/:category/ad-values/:field")
	fmt.Println("  GET  /api/categories/:category/chains")
	fmt.Println("  GET  /api/categories/:category/first-spec-fields")
	fmt.Println("  GET  /api/categories/:category/last-spec-field")
	fmt.Println("  POST /api/categories/:category/search")
	fmt.Println("  GET  /api/ads/:id/filter-values")

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
