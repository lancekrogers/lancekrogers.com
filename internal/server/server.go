package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Config holds server configuration
type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// NewServer creates a new HTTP server
func NewServer(cfg Config, router *mux.Router) *http.Server {
	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}

// Start starts the server
func Start(srv *http.Server, logger *log.Logger) {
	go func() {
		logger.Printf("Server starting on %s", srv.Addr)
		logger.Printf("Visit http://localhost%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()
}

// Shutdown gracefully shuts down the server
func Shutdown(ctx context.Context, srv *http.Server, logger *log.Logger) error {
	logger.Println("Shutting down server...")

	if err := srv.Shutdown(ctx); err != nil {
		logger.Printf("Server forced to shutdown: %v", err)
		return err
	}

	logger.Println("Server exited gracefully")
	return nil
}
