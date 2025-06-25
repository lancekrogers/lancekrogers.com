package server

import (
	"io/fs"
	"net/http"

	"blockhead.consulting/internal/blog"
	"blockhead.consulting/internal/handlers"
	"blockhead.consulting/internal/security"
	"github.com/gorilla/mux"
)

// SetupRoutes configures all server routes
func SetupRoutes(r *mux.Router, handler *handlers.Handler, siteConfig *handlers.SiteConfig, securityConfig *security.Config, staticFS fs.FS, blogService blog.Service) error {
	// Routes
	r.HandleFunc("/", handler.HomeHandler).Methods("GET")

	// HTMX content-only routes
	r.HandleFunc("/content/home", handler.HomeContentHandler).Methods("GET")

	// Blog routes (conditional based on config)
	if siteConfig.BlogEnabled {
		r.HandleFunc("/blog", handler.BlogHandler).Methods("GET")
		r.HandleFunc("/blog/{slug}", handler.BlogPostHandler).Methods("GET")
		r.HandleFunc("/content/blog", handler.BlogContentHandler).Methods("GET")

		// API routes for enhanced blog features
		r.HandleFunc("/api/blog/search", handler.BlogSearchHandler).Methods("GET")
		r.HandleFunc("/api/blog/posts", handler.BlogPaginationHandler).Methods("GET")
		r.HandleFunc("/api/blog/tags", handler.BlogTagsHandler).Methods("GET")
		r.HandleFunc("/blog/{slug}/related", handler.BlogRelatedHandler).Methods("GET")
		r.HandleFunc("/blog/{slug}/navigation", handler.BlogNavigationHandler).Methods("GET")
		r.HandleFunc("/feed.xml", handler.BlogRSSHandler).Methods("GET")
		r.HandleFunc("/blog/feed.xml", handler.BlogRSSHandler).Methods("GET")
	}

	// About routes
	r.HandleFunc("/about", handler.AboutHandler).Methods("GET")
	r.HandleFunc("/content/about", handler.AboutContentHandler).Methods("GET")

	// Work experience routes - redirect to about page
	r.HandleFunc("/work", handler.WorkRedirectHandler).Methods("GET")
	r.HandleFunc("/content/work", handler.WorkContentRedirectHandler).Methods("GET")

	r.HandleFunc("/contact", handler.ContactHandler).Methods("POST")

	// Health check endpoint for Docker/monitoring
	r.HandleFunc("/health", handler.HealthHandler).Methods("GET")

	// Calendar routes (conditional based on config)
	if siteConfig.CalendarEnabled {
		r.HandleFunc("/calendar", handler.CalendarHandler).Methods("GET")
		r.HandleFunc("/content/calendar", handler.CalendarContentHandler).Methods("GET")
		r.HandleFunc("/api/slots", handler.SlotsHandler).Methods("GET")
		r.HandleFunc("/api/book", handler.BookingHandler).Methods("POST")
	}

	// Static files - serve from embedded filesystem
	staticFiles, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))))

	// Standard browser-requested files served from root
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/robots.txt")
	}).Methods("GET")

	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/favicon.ico")
	}).Methods("GET")

	r.HandleFunc("/apple-touch-icon.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/apple-touch-icon.png")
	}).Methods("GET")

	r.HandleFunc("/apple-touch-icon-precomposed.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/apple-touch-icon-precomposed.png")
	}).Methods("GET")

	r.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/sitemap.xml")
	}).Methods("GET")

	// Admin endpoints (protect these in production!)
	r.HandleFunc("/admin/slots", handler.AdminSlotsHandler).Methods("GET", "POST")

	// Security middleware stack (order matters!)
	r.Use(security.SecurityMiddleware(securityConfig))
	r.Use(LoggingMiddleware)

	// Path redirect middleware for blog posts (only if blog is enabled)
	if siteConfig.BlogEnabled && blogService != nil {
		r.Use(PathRedirectMiddleware(blogService))
	}

	return nil
}
