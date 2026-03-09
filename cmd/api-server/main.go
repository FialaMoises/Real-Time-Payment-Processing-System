package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/real-time-payments/api"
	"github.com/yourusername/real-time-payments/config"
	"github.com/yourusername/real-time-payments/internal/auth"
	"github.com/yourusername/real-time-payments/pkg/database"
	"github.com/yourusername/real-time-payments/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Log.Level)
	logger.Info().Msg("Starting Real-Time Payment Processing System...")

	// Connect to database
	db, err := database.NewPostgresConnection(
		cfg.Database.URL,
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Initialize JWT service
	jwtService := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiration)

	// Setup router
	router := api.NewRouter(db.DB, jwtService)
	engine := router.Setup()

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info().
			Str("port", cfg.Server.Port).
			Str("env", cfg.Server.Env).
			Msg("Server is running")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited")
}
