package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/zhisme/tinylist/internal/config"
	"github.com/zhisme/tinylist/internal/db"
	"github.com/zhisme/tinylist/internal/handlers/private"
	"github.com/zhisme/tinylist/internal/handlers/public"
	"github.com/zhisme/tinylist/internal/mailer"
	authmw "github.com/zhisme/tinylist/internal/middleware"
	"github.com/zhisme/tinylist/internal/worker"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize mailer (unconfigured - settings loaded from DB)
	mail := mailer.New()

	// Load SMTP settings from database
	loadSMTPFromDB(database, mail)

	// Initialize campaign worker
	campaignWorker := worker.NewCampaignWorker(database, mail, cfg.Sending, cfg.Server.PublicURL)

	// Initialize router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy"}`)
	})

	// Public API routes
	subscribeHandler := public.NewSubscribeHandler(database, mail, cfg.Server.PublicURL)
	verifyHandler := public.NewVerifyHandler(database)
	unsubscribeHandler := public.NewUnsubscribeHandler(database)

	r.Route("/api", func(r chi.Router) {
		r.Post("/subscribe", subscribeHandler.Subscribe)
		r.Get("/verify/{token}", verifyHandler.Verify)
		r.Get("/unsubscribe/{token}", unsubscribeHandler.Unsubscribe)
	})

	// Private API routes (protected by Basic Auth)
	subscriberHandler := private.NewSubscriberHandler(database, mail, cfg.Server.PublicURL)
	campaignHandler := private.NewCampaignHandler(database, campaignWorker, mail)
	settingsHandler := private.NewSettingsHandler(database, mail)
	r.Route("/api/private", func(r chi.Router) {
		r.Use(authmw.BasicAuth(cfg.Auth))
		r.Mount("/subscribers", subscriberHandler.Routes())
		r.Mount("/campaigns", campaignHandler.Routes())
		r.Mount("/settings", settingsHandler.Routes())
	})

	log.Printf("Basic Auth enabled for /api/private (user: %s)", cfg.Auth.Username)

	// Server configuration
	port := cfg.Server.Port
  // TODO: maybe move to config yaml
	if envPort := os.Getenv("PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", &port)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting TinyList server on %s:%d", cfg.Server.Host, port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// loadSMTPFromDB loads SMTP settings from database and reconfigures the mailer
func loadSMTPFromDB(database *db.DB, mail *mailer.Mailer) {
	settings, err := database.GetAllSettings()
	if err != nil {
		log.Printf("Warning: failed to load settings from DB: %v", err)
		return
	}

	// Check if SMTP is configured in DB
	host := settings["smtp_host"]
	if host == "" {
		log.Println("SMTP not configured - configure via admin UI Settings page")
		return
	}

	port := 587
	if portStr := settings["smtp_port"]; portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	tls := settings["smtp_tls"] == "true"

	mail.Reconfigure(
		host,
		port,
		settings["smtp_username"],
		settings["smtp_password"],
		settings["smtp_from_email"],
		settings["smtp_from_name"],
		tls,
	)

	log.Println("SMTP settings loaded from database")
}
