package app

import (
	"context"
	"embed"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blockhead.consulting/internal/handlers"
	"blockhead.consulting/internal/server"
	"github.com/gorilla/mux"
)

// App represents the main application structure
type App struct {
	dependencies *Dependencies
	logger       *log.Logger
}

// New creates a new App instance with embedded file systems
func New(templateFS, staticFS, blogFS embed.FS) *App {
	logger := log.New(os.Stdout, "[app] ", log.LstdFlags)
	return &App{
		dependencies: NewDependencies(templateFS, staticFS, blogFS),
		logger:       logger,
	}
}

// Initialize initializes all application dependencies
func (a *App) Initialize() error {
	return a.dependencies.Initialize()
}

// CreateHandler creates the main HTTP handler with all dependencies
func (a *App) CreateHandler() error {
	logger := log.New(os.Stdout, "[server] ", log.LstdFlags)

	// Convert legacy config to handlers config
	handlerSiteConfig := &handlers.SiteConfig{
		CalendarEnabled: a.dependencies.SiteConfig.CalendarEnabled,
		BlogEnabled:     a.dependencies.SiteConfig.BlogEnabled,
		SiteName:        a.dependencies.SiteConfig.SiteName,
		Environment:     a.dependencies.SiteConfig.Environment,
		HeroStyle:       a.dependencies.SiteConfig.HeroStyle,
		ConsoleLogging:  a.dependencies.SiteConfig.ConsoleLogging,
		CSPNonce:        a.dependencies.SiteConfig.CSPNonce,
	}

	a.dependencies.Handler = handlers.New(
		a.dependencies.TemplateManager.GetTemplates(),
		a.dependencies.BlogService,
		a.dependencies.BioService,
		a.dependencies.ContactService,
		a.dependencies.EmailService,
		a.dependencies.GitStorageService,
		a.dependencies.SecurityConfig,
		handlerSiteConfig,
		a.dependencies.ConfigService,
		a.dependencies.AppConfig,
		a.dependencies.WorkConfig,
		a.dependencies.BookingService,
		a.dependencies.GitHubService,
		a.dependencies.MermaidRenderer,
		logger,
	)

	return nil
}

// CreateRouter creates and configures the HTTP router
func (a *App) CreateRouter() (*mux.Router, error) {
	r := mux.NewRouter()

	// Convert legacy config to handlers config for routes
	handlerSiteConfig := &handlers.SiteConfig{
		CalendarEnabled: a.dependencies.SiteConfig.CalendarEnabled,
		BlogEnabled:     a.dependencies.SiteConfig.BlogEnabled,
		SiteName:        a.dependencies.SiteConfig.SiteName,
		Environment:     a.dependencies.SiteConfig.Environment,
		HeroStyle:       a.dependencies.SiteConfig.HeroStyle,
		ConsoleLogging:  a.dependencies.SiteConfig.ConsoleLogging,
		CSPNonce:        a.dependencies.SiteConfig.CSPNonce,
	}

	// Setup routes
	if err := server.SetupRoutes(r, a.dependencies.Handler, handlerSiteConfig, a.dependencies.SecurityConfig, a.dependencies.StaticFS, a.dependencies.BlogService); err != nil {
		return nil, err
	}

	return r, nil
}

// Run starts the application and blocks until shutdown
func (a *App) Run() error {
	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	// Create router
	router, err := a.CreateRouter()
	if err != nil {
		return err
	}

	// Create server configuration
	serverConfig := server.Config{
		Port:         port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create and start server
	srv := server.NewServer(serverConfig, router)
	logger := log.New(os.Stdout, "[server] ", log.LstdFlags)
	server.Start(srv, logger)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx, srv, logger); err != nil {
		return err
	}

	return nil
}