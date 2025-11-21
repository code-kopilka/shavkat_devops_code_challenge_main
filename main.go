package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LogicGateTech/devops-code-challenge/conf"
	route "github.com/LogicGateTech/devops-code-challenge/http"
)

func main() {
	// Load configuration first (before creating logger)
	cfg, err := conf.New()
	if err != nil {
		// Use basic logger for config errors
		log := slog.New(slog.NewTextHandler(os.Stdout, nil))
		log.With("error", err).Error("Error loading configuration")
		os.Exit(1)
	}

	// Create logger based on configuration
	log := conf.NewLogger(cfg)

	// Initialize router
	router, err := route.New()
	if err != nil {
		log.With("error", err).Error("Error initializing router")
		os.Exit(1)
	}

	router.Bootstrap()

	// Create HTTP server with timeouts
	bindAddr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         bindAddr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Server starting", "bind-addr", bindAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.With("error", err).Error("Server failed to start")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Server shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close router resources (database connections, etc.)
	if router != nil {
		if err := router.Close(); err != nil {
			log.With("error", err).Error("Error closing router resources")
		}
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.With("error", err).Error("Server forced to shutdown")
		os.Exit(1)
	}

	log.Info("Server stopped")
}
