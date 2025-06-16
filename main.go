package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"blockhead.consulting/internal/bio"
	"blockhead.consulting/internal/blog"
	"blockhead.consulting/internal/config"
	"blockhead.consulting/internal/contact"
	"blockhead.consulting/internal/email"
	"blockhead.consulting/internal/events"
	"blockhead.consulting/internal/security"
	"blockhead.consulting/internal/storage/git"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type BlogPost struct {
	Slug        string
	Title       string
	Date        time.Time
	Summary     string
	Content     template.HTML
	ReadingTime int
	Tags        []string
	FileName    string
}

type BlogFrontmatter struct {
	Title       string    `yaml:"title"`
	Date        time.Time `yaml:"date"`
	Summary     string    `yaml:"summary"`
	Tags        []string  `yaml:"tags"`
	ReadingTime int       `yaml:"readingTime"`
}

type TimeSlot struct {
	ID        string    `json:"id"`
	Date      string    `json:"date"`
	Time      string    `json:"time"`
	Available bool      `json:"available"`
	Booked    bool      `json:"booked"`
	BookedBy  string    `json:"bookedBy,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type BookingRequest struct {
	SlotID      string `json:"slotId"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Company     string `json:"company"`
	ServiceType string `json:"serviceType"`
	Message     string `json:"message"`
}

// Legacy structures - these are now in internal/security package

type SiteConfig struct {
	CalendarEnabled bool
	BlogEnabled     bool
	SiteName        string
	Environment     string
	HeroStyle       string // "professional" or "cyberpunk"
	ConsoleLogging  bool   // Enable/disable JavaScript console logging
	CSPNonce        string // Content Security Policy nonce for inline scripts
}

// Input validation patterns
// Validation patterns moved to internal/security package

//go:embed templates
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

//go:embed content/blog/*.md content/blog.yml
var blogFS embed.FS

var (
	templates       *template.Template
	blogPosts       []BlogPost
	blogService     blog.Service
	bioService      bio.Service
	contactService  contact.Service
	emailService    email.Service
	gitStorageService git.Service
	timeSlots       = make(map[string]*TimeSlot)
	bookingsFile    = "data/bookings.json"
	securityConfig  *security.Config
	siteConfig      *SiteConfig
	configService   config.Service
	appConfig       *config.SiteConfig
	workConfig      *config.WorkConfig
)

func init() {
	// Create directories if they don't exist
	dirs := []string{"templates", "static", "data", "content/blog"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Warning: Could not create directory %s: %v", dir, err)
		}
	}

	// Load templates from embedded filesystem with layout inheritance
	var err error
	
	// Define template functions
	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"replaceAll": strings.ReplaceAll,
		"slug": func(s string) string {
			s = strings.ToLower(s)
			s = strings.ReplaceAll(s, " ", "-")
			s = strings.ReplaceAll(s, ".", "")
			s = strings.ReplaceAll(s, "&", "")
			return s
		},
	}
	
	templates = template.New("main").Funcs(funcMap)
	
	// Parse all HTML templates with multiple patterns
	// Order matters - pages must be parsed last
	patterns := []string{
		"templates/layouts/partials/*.html",
		"templates/fragments/*.html", 
		"templates/layouts/*.html",
		"templates/pages/*.html",
	}
	
	for _, pattern := range patterns {
		if _, err = templates.ParseFS(templateFS, pattern); err != nil {
			// Some patterns might not match any files, that's OK
			log.Printf("Warning: pattern %s matched no files: %v", pattern, err)
		}
	}
	
	log.Printf("Loaded %d templates", len(templates.Templates()))

	// Initialize security configuration first (needed for CSP nonce)
	initializeSecurity()

	// Initialize configuration (uses security config)
	initializeConfig()

	// Load blog posts (after config is initialized)
	if err := initializeBlogService(); err != nil {
		log.Fatalf("Failed to initialize blog service: %v", err)
	}

	// Load existing bookings
	loadBookings()

	// Initialize available time slots for the next 30 days
	initializeTimeSlots()
}

func main() {
	r := mux.NewRouter()

	// Routes
	r.HandleFunc("/", homeHandler).Methods("GET")
	
	// HTMX content-only routes
	r.HandleFunc("/content/home", homeContentHandler).Methods("GET")
	
	// Blog routes (conditional based on config)
	if siteConfig.BlogEnabled {
		r.HandleFunc("/blog", blogHandler).Methods("GET")
		r.HandleFunc("/blog/{slug}", blogPostHandler).Methods("GET")
		r.HandleFunc("/content/blog", blogContentHandler).Methods("GET")
	}
	
	// About routes
	r.HandleFunc("/about", aboutHandler).Methods("GET")
	r.HandleFunc("/content/about", aboutContentHandler).Methods("GET")
	
	// Work experience routes
	r.HandleFunc("/work", workHandler).Methods("GET")
	r.HandleFunc("/content/work", workContentHandler).Methods("GET")
	
	r.HandleFunc("/contact", contactHandler).Methods("POST")
	
	// Health check endpoint for Docker/monitoring
	r.HandleFunc("/health", healthHandler).Methods("GET")

	// Calendar routes (conditional based on config)
	if siteConfig.CalendarEnabled {
		r.HandleFunc("/calendar", calendarHandler).Methods("GET")
		r.HandleFunc("/content/calendar", calendarContentHandler).Methods("GET")
		r.HandleFunc("/api/slots", slotsHandler).Methods("GET")
		r.HandleFunc("/api/book", bookingHandler).Methods("POST")
	}

	// Static files - serve from embedded filesystem
	staticFiles, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Failed to create static file sub-filesystem: %v", err)
	}
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))))

	// Admin endpoints (protect these in production!)
	r.HandleFunc("/admin/slots", adminSlotsHandler).Methods("GET", "POST")
	
	// Security middleware stack (order matters!)
	r.Use(security.SecurityMiddleware(securityConfig))
	r.Use(loggingMiddleware)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	// Create server with timeouts
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		log.Printf("Visit http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server exited gracefully")
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)

		// Log static file requests specifically
		if strings.HasPrefix(r.URL.Path, "/static/") {
			filePath := "./static/" + strings.TrimPrefix(r.URL.Path, "/static/")
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				log.Printf("WARNING: Static file not found: %s", filePath)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get brief bio
	var bioBrief *bio.Bio
	if bioService != nil {
		var err error
		bioBrief, err = bioService.GetBrief(ctx)
		if err != nil {
			log.Printf("Warning: Failed to load brief bio: %v", err)
		}
	}
	
	// Determine title based on config availability
	title := "Blockhead Consulting - Enterprise Blockchain & AI Infrastructure"
	if appConfig != nil && appConfig.Site.Name != "" {
		title = appConfig.Site.Name + " - " + appConfig.Site.Tagline
	}
	
	data := struct {
		Title    string
		Page     string
		Config   *SiteConfig
		AppConfig *config.SiteConfig
		BioBrief *bio.Bio
	}{
		Title:    title,
		Page:     "home",
		Config:   siteConfig,
		AppConfig: appConfig,
		BioBrief: bioBrief,
	}

	// Use ExecuteTemplate directly with the specific page template
	if err := templates.ExecuteTemplate(w, "home-full.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func blogHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title      string
		Page       string
		Posts      []BlogPost
		Config     *SiteConfig
		AppConfig  *config.SiteConfig
		WorkConfig *config.WorkConfig
		BlogConfig *blog.BlogConfig
	}{
		Title:      "Blog - Blockhead Consulting",
		Page:       "blog",
		Posts:      blogPosts,
		Config:     siteConfig,
		AppConfig:  appConfig,
		WorkConfig: workConfig,
		BlogConfig: blogService.GetBlogConfig(),
	}

	if err := templates.ExecuteTemplate(w, "page-blog.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func blogPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	ctx := r.Context()
	
	// Debug: Check if blogService is nil
	if blogService == nil {
		log.Printf("ERROR: blogService is nil in blogPostHandler")
		http.NotFound(w, r)
		return
	}
	
	// Use the blog service to get the post
	servicePost, err := blogService.GetBySlug(ctx, slug)
	log.Printf("DEBUG: GetBySlug('%s') returned: post=%v, err=%v", slug, servicePost != nil, err)
	
	if err != nil || servicePost == nil {
		log.Printf("ERROR: Blog post '%s' not found", slug)
		http.NotFound(w, r)
		return
	}
	
	// Convert service post to legacy BlogPost structure for template compatibility
	post := &BlogPost{
		Slug:        servicePost.Slug,
		Title:       servicePost.Title,
		Date:        servicePost.Date,
		Summary:     servicePost.Summary,
		Content:     template.HTML(servicePost.Content),
		ReadingTime: servicePost.ReadingTime,
		Tags:        servicePost.Tags,
		FileName:    servicePost.FileName,
	}

	data := struct {
		Title     string
		Page      string
		Post      *BlogPost
		Config    *SiteConfig
		AppConfig *config.SiteConfig
	}{
		Title:     post.Title + " - Blockhead Consulting",
		Page:      "blog",
		Post:      post,
		Config:    siteConfig,
		AppConfig: appConfig,
	}

	if err := templates.ExecuteTemplate(w, "blog-post.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func calendarHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title     string
		Page      string
		Config    *SiteConfig
		AppConfig *config.SiteConfig
	}{
		Title:     "Book a Consultation - Blockhead Consulting",
		Page:      "calendar",
		Config:    siteConfig,
		AppConfig: appConfig,
	}

	if err := templates.ExecuteTemplate(w, "page-calendar.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func slotsHandler(w http.ResponseWriter, r *http.Request) {
	// Get available slots for the next 30 days
	var availableSlots []*TimeSlot
	for _, slot := range timeSlots {
		if slot.Available && !slot.Booked {
			availableSlots = append(availableSlots, slot)
		}
	}

	// Sort by date and time
	sort.Slice(availableSlots, func(i, j int) bool {
		if availableSlots[i].Date == availableSlots[j].Date {
			return availableSlots[i].Time < availableSlots[j].Time
		}
		return availableSlots[i].Date < availableSlots[j].Date
	})

	// Limit to first 20 slots for demo
	if len(availableSlots) > 20 {
		availableSlots = availableSlots[:20]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availableSlots)
}

func bookingHandler(w http.ResponseWriter, r *http.Request) {
	var req BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("SECURITY: Invalid JSON in booking request from %s", security.ExtractClientIP(r))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Enhanced validation
	if err := validateBookingRequest(&req); err != nil {
		log.Printf("SECURITY: Invalid booking request from %s: %v", security.ExtractClientIP(r), err)
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Check if slot exists and is available
	slot, exists := timeSlots[req.SlotID]
	if !exists || !slot.Available || slot.Booked {
		http.Error(w, "Slot not available", http.StatusConflict)
		return
	}

	// Book the slot
	slot.Booked = true
	slot.BookedBy = req.Email

	// Save bookings
	saveBookings()

	// Send confirmation (implement email sending later)
	log.Printf("BOOKING: New booking from %s - Slot: %s, Email: %s, Service: %s", 
		security.ExtractClientIP(r), req.SlotID, req.Email, req.ServiceType)

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Booking confirmed! You'll receive a confirmation email shortly.",
	})
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Create contact request from form data
	req := &contact.ContactRequest{
		Name:    strings.TrimSpace(r.FormValue("name")),
		Email:   strings.TrimSpace(r.FormValue("email")),
		Company: strings.TrimSpace(r.FormValue("company")),
		Message: strings.TrimSpace(r.FormValue("message")),
	}

	// Process contact form using the contact service
	if contactService != nil {
		log.Printf("CONTACT: Processing form from %s <%s>", req.Name, req.Email)
		_, err := contactService.ProcessContactForm(ctx, req, r)
		if err != nil {
			log.Printf("CONTACT: Error processing form: %v", err)
			// Handle validation errors with HTMX-friendly response
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<div class="alert error">%s</div>`, err.Error())
			return
		}
		log.Printf("CONTACT: Form processed successfully")
	} else {
		// Fallback for when contact service is not available
		log.Printf("Contact form (fallback): Name=%s, Email=%s, Message=%s", req.Name, req.Email, req.Message)
	}

	// Return HTMX success response
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<div class="alert success">Message sent successfully! I'll get back to you within 24 hours.</div>`)
}

func adminSlotsHandler(w http.ResponseWriter, r *http.Request) {
	// Get admin credentials from environment variables
	expectedUser := os.Getenv("ADMIN_USERNAME")
	expectedPass := os.Getenv("ADMIN_PASSWORD")
	
	// Ensure credentials are configured
	if expectedUser == "" || expectedPass == "" {
		log.Printf("SECURITY: Admin credentials not configured in environment variables")
		http.Error(w, "Admin interface is not configured", http.StatusServiceUnavailable)
		return
	}
	
	// Basic auth check
	username, password, ok := r.BasicAuth()
	if !ok || subtle.ConstantTimeCompare([]byte(username), []byte(expectedUser)) != 1 ||
		subtle.ConstantTimeCompare([]byte(password), []byte(expectedPass)) != 1 {
		w.Header().Set("WWW-Authenticate", `Basic realm="Admin"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		// Handle slot updates
		// Implementation depends on your needs
	}

	// Return admin interface
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Admin Interface</h1><p>Slots: %d</p>", len(timeSlots))
}

// Health check handler for Docker and monitoring
func healthHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version": "1.0.0",
		"services": map[string]string{
			"blog": "ok",
			"templates": "ok",
		},
	}
	
	// Check if blog service is working
	if blogService != nil {
		ctx := r.Context()
		posts := blogService.GetAll(ctx)
		status["services"].(map[string]string)["blog"] = fmt.Sprintf("ok (%d posts)", len(posts))
	}
	
	// Check if contact service is working
	if contactService != nil {
		status["services"].(map[string]string)["contact"] = "ok"
	}
	
	// Check if email service is working
	if emailService != nil {
		status["services"].(map[string]string)["email"] = "ok"
	}
	
	// Check if git storage is working
	if gitStorageService != nil {
		status["services"].(map[string]string)["storage"] = "ok"
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// Security initialization
func initializeSecurity() {
	// Generate CSP nonce
	nonce, err := security.GenerateNonce()
	if err != nil {
		log.Fatal("Failed to generate CSP nonce:", err)
	}

	// Initialize rate limiter with default config
	rateLimiterConfig := security.DefaultRateLimiterConfig()
	rateLimiter := security.NewRateLimiter(rateLimiterConfig)

	securityConfig = &security.Config{
		CSPNonce:    nonce,
		RateLimiter: rateLimiter,
		ValidFileTypes: map[string]bool{
			".css":  true,
			".js":   true,
			".png":  true,
			".jpg":  true,
			".jpeg": true,
			".gif":  true,
			".svg":  true,
			".ico":  true,
			".woff": true,
			".woff2": true,
		},
		MaxUploadSize:  10 << 20, // 10MB
		SessionTimeout: 30 * time.Minute,
		MaxRequestSize: 1 << 20, // 1MB
	}
}

// Legacy rate limiter methods - replaced by security package
/*
func (rl *RateLimiter) isAllowed(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientIP]

	if !exists {
		client = &ClientInfo{
			requests: []time.Time{now},
		}
		rl.clients[clientIP] = client
		return true
	}

	// Check if client is currently blocked
	if client.blocked && now.Before(client.blockUntil) {
		return false
	}

	// Reset block status if block period expired
	if client.blocked && now.After(client.blockUntil) {
		client.blocked = false
		client.requests = []time.Time{}
	}

	// Remove old requests outside the window
	cutoff := now.Add(-rl.window)
	validRequests := []time.Time{}
	for _, req := range client.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}

	// Check if under rate limit
	if len(validRequests) >= rl.maxReqs {
		client.blocked = true
		client.blockUntil = now.Add(1 * time.Minute) // Block for 1 minute
		client.requests = validRequests
		return false
	}

	// Add current request
	client.requests = append(validRequests, now)
	return true
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window * 2) // Keep some history

		for ip, client := range rl.clients {
			// Remove completely stale clients
			if len(client.requests) == 0 {
				delete(rl.clients, ip)
				continue
			}

			// Clean old requests
			validRequests := []time.Time{}
			for _, req := range client.requests {
				if req.After(cutoff) {
					validRequests = append(validRequests, req)
				}
			}
			client.requests = validRequests

			// Remove clients with no recent activity
			if len(validRequests) == 0 && (!client.blocked || now.After(client.blockUntil)) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Security middleware
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content Security Policy
		cspPolicy := fmt.Sprintf(`
			default-src 'self';
			script-src 'self' 'nonce-%s' https://unpkg.com;
			style-src 'self' 'unsafe-inline';
			img-src 'self' data: https:;
			font-src 'self' data:;
			connect-src 'self';
			frame-ancestors 'none';
			base-uri 'self';
			form-action 'self';
		`, securityConfig.CSPNonce)
		
		// Remove newlines and extra spaces from CSP
		cspPolicy = strings.ReplaceAll(cspPolicy, "\n", "")
		cspPolicy = strings.ReplaceAll(cspPolicy, "\t", "")
		cspPolicy = regexp.MustCompile(`\s+`).ReplaceAllString(cspPolicy, " ")
		cspPolicy = strings.TrimSpace(cspPolicy)

		w.Header().Set("Content-Security-Policy", cspPolicy)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

		next.ServeHTTP(w, r)
	})
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := security.ExtractClientIP(r)
		
		if !securityConfig.RateLimiter.IsAllowed(clientIP) {
			log.Printf("SECURITY: Rate limit exceeded for IP %s", clientIP)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func requestSizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, securityConfig.MaxRequestSize)
		next.ServeHTTP(w, r)
	})
}

func securityLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		clientIP := security.ExtractClientIP(r)
		
		// Log security-relevant events
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
			log.Printf("SECURITY: %s %s from %s User-Agent: %s", 
				r.Method, r.URL.Path, clientIP, r.UserAgent())
		}

		// Detect potential attacks
		if detectSuspiciousActivity(r) {
			log.Printf("SECURITY: Suspicious activity detected from %s: %s %s", 
				clientIP, r.Method, r.URL.Path)
		}

		next.ServeHTTP(w, r)
		
		duration := time.Since(start)
		if duration > 5*time.Second {
			log.Printf("SECURITY: Slow request detected: %s %s took %v", 
				r.Method, r.URL.Path, duration)
		}
	})
}
*/

// Helper functions
// getClientIP moved to security.ExtractClientIP

// detectSuspiciousActivity moved to security.DetectSuspiciousActivity

// Enhanced input validation
func validateBookingRequest(req *BookingRequest) error {
	if req.SlotID == "" || req.Name == "" || req.Email == "" || req.ServiceType == "" {
		return fmt.Errorf("missing required fields")
	}

	if !security.SlotIDRegex.MatchString(req.SlotID) {
		return fmt.Errorf("invalid slot ID format")
	}

	if !security.NameRegex.MatchString(req.Name) {
		return fmt.Errorf("invalid name format")
	}

	if !security.EmailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	if req.Company != "" && !security.CompanyRegex.MatchString(req.Company) {
		return fmt.Errorf("invalid company format")
	}

	if req.Message != "" {
		if len(req.Message) > 2000 {
			return fmt.Errorf("message too long (max 2000 characters)")
		}
		if !security.MessageRegex.MatchString(req.Message) {
			return fmt.Errorf("invalid message format")
		}
	}

	validServiceTypes := map[string]bool{
		"crypto-infrastructure": true,
		"ai-claude":            true,
		"both":                 true,
		"other":                true,
	}

	if !validServiceTypes[req.ServiceType] {
		return fmt.Errorf("invalid service type")
	}

	return nil
}

// Configuration initialization
func initializeConfig() {
	// Load .env file if it exists (ignore errors - file may not exist)
	godotenv.Load()
	
	// Initialize config service
	configLogger := log.New(os.Stdout, "[config] ", log.LstdFlags)
	configService = config.NewService(configLogger)
	
	// Load configuration from YAML file
	var err error
	appConfig, err = configService.LoadConfig("content/site.yml")
	if err != nil {
		log.Printf("Warning: Failed to load site.yml, using environment variables: %v", err)
		// Fallback to environment-based configuration
		initializeLegacyConfig()
		return
	}
	
	// Create legacy SiteConfig for backward compatibility
	environment := getEnv("ENVIRONMENT", "development")
	siteConfig = &SiteConfig{
		CalendarEnabled: getEnvBool("CALENDAR_ENABLED", appConfig.Features.CalendarEnabled),
		BlogEnabled:     getEnvBool("BLOG_ENABLED", appConfig.Features.BlogEnabled),
		SiteName:        appConfig.Site.Name,
		Environment:     environment,
		HeroStyle:       getEnv("HERO_STYLE", "professional"),
		ConsoleLogging:  getEnvBool("CONSOLE_LOGGING", environment == "development"),
		CSPNonce:        securityConfig.CSPNonce,
	}
	
	// Validate hero style
	if siteConfig.HeroStyle != "professional" && siteConfig.HeroStyle != "cyberpunk" {
		log.Printf("Warning: Invalid HERO_STYLE '%s', defaulting to 'professional'", siteConfig.HeroStyle)
		siteConfig.HeroStyle = "professional"
	}
	
	log.Printf("CONFIG: Loaded from site.yml - %s", appConfig.Site.Name)
	log.Printf("CONFIG: Calendar enabled: %v", siteConfig.CalendarEnabled)
	log.Printf("CONFIG: Blog enabled: %v", siteConfig.BlogEnabled)
	log.Printf("CONFIG: Environment: %s", siteConfig.Environment)
	log.Printf("CONFIG: Hero style: %s", siteConfig.HeroStyle)
	
	// Load work configuration
	workConfig, err = configService.LoadWorkConfig("content/work.yml")
	if err != nil {
		log.Printf("Warning: Failed to load work.yml: %v", err)
		workConfig = nil
	} else {
		log.Printf("CONFIG: Loaded work configuration with %d sections", 3)
	}
}

// Legacy configuration fallback
func initializeLegacyConfig() {
	calendarEnabled := true // Default to enabled
	
	// Check environment variable
	if envCalendar := os.Getenv("CALENDAR_ENABLED"); envCalendar != "" {
		if parsed, err := strconv.ParseBool(envCalendar); err == nil {
			calendarEnabled = parsed
		}
	}
	
	// Blog enabled (default true, can be disabled for development)
	blogEnabled := true
	if blogEnv := os.Getenv("BLOG_ENABLED"); blogEnv != "" {
		if parsed, err := strconv.ParseBool(blogEnv); err == nil {
			blogEnabled = parsed
		}
	}
	
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}
	
	siteName := os.Getenv("SITE_NAME")
	if siteName == "" {
		siteName = "Blockhead Consulting"
	}
	
	heroStyle := os.Getenv("HERO_STYLE")
	if heroStyle == "" {
		heroStyle = "professional" // Default to professional mode
	}
	
	siteConfig = &SiteConfig{
		CalendarEnabled: calendarEnabled,
		BlogEnabled:     blogEnabled,
		SiteName:        siteName,
		Environment:     environment,
		HeroStyle:       heroStyle,
		ConsoleLogging:  getEnvBool("CONSOLE_LOGGING", environment == "development"),
		CSPNonce:        securityConfig.CSPNonce,
	}
	
	log.Printf("CONFIG: Using legacy environment variables")
	log.Printf("CONFIG: Calendar enabled: %v", siteConfig.CalendarEnabled)
	log.Printf("CONFIG: Blog enabled: %v", siteConfig.BlogEnabled)
	log.Printf("CONFIG: Environment: %s", siteConfig.Environment)
	log.Printf("CONFIG: Hero style: %s", siteConfig.HeroStyle)
	log.Printf("CONFIG: Console logging: %v", siteConfig.ConsoleLogging)
}

func initializeBlogService() error {
	// Skip if blog is disabled
	if !siteConfig.BlogEnabled {
		log.Printf("Blog disabled - skipping blog service initialization")
		return nil
	}
	
	// Create logger
	logger := log.New(os.Stdout, "[blog] ", log.LstdFlags)
	
	// Create event bus
	eventBus := events.NewInMemoryEventBus(5, logger)
	
	// Create blog service
	blogService = blog.NewService(blogFS, logger, eventBus)
	
	// Initialize email service
	emailConfig := &email.EmailConfig{
		SMTPHost:    getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:    getEnvInt("SMTP_PORT", 587),
		Username:    getEnv("SMTP_USERNAME", ""),
		Password:    getEnv("SMTP_PASSWORD", ""),
		FromAddress: getEnv("SMTP_FROM_ADDRESS", ""),
		FromName:    getEnv("SMTP_FROM_NAME", "Blockhead Consulting"),
		TLSEnabled:  getEnvBool("SMTP_TLS_ENABLED", true),
	}
	emailService = email.NewService(emailConfig)
	
	// Log email configuration status
	if emailConfig.SMTPHost != "" && emailConfig.Username != "" && emailConfig.Password != "" {
		log.Printf("EMAIL: Service initialized with SMTP host: %s", emailConfig.SMTPHost)
	} else {
		log.Printf("EMAIL: Service initialized but missing configuration")
		if emailConfig.SMTPHost == "" {
			log.Printf("EMAIL: Missing SMTP_HOST")
		}
		if emailConfig.Username == "" {
			log.Printf("EMAIL: Missing SMTP_USERNAME")
		}
		if emailConfig.Password == "" {
			log.Printf("EMAIL: Missing SMTP_PASSWORD")
		}
	}
	
	// Initialize Git storage service
	gitConfig := git.StorageConfig{
		RepoPath:      getEnv("GIT_REPO_PATH", "./data/messages"),
		EncryptionKey: getEnv("GIT_ENCRYPTION_KEY", ""),
		RemoteURL:     getEnv("GIT_REMOTE_URL", ""),
		PushOnWrite:   getEnvBool("GIT_PUSH_ON_WRITE", false),
		Branch:        getEnv("GIT_BRANCH", "main"),
		CommitAuthor:  getEnv("GIT_COMMIT_AUTHOR", "Blockhead Consulting Bot"),
		CommitEmail:   getEnv("GIT_COMMIT_EMAIL", "bot@blockhead.consulting"),
	}
	
	var err error
	gitStorageService, err = git.NewService(gitConfig, logger, eventBus)
	if err != nil {
		log.Printf("Warning: Git storage service initialization failed: %v", err)
		// Continue without git storage for development
	}
	
	// Initialize contact service
	contactLogger := log.New(os.Stdout, "[contact] ", log.LstdFlags)
	adminEmail := getEnv("ADMIN_EMAIL", "")
	if adminEmail != "" {
		log.Printf("CONTACT: Admin email configured: %s", adminEmail)
	} else {
		log.Printf("CONTACT: WARNING - No ADMIN_EMAIL configured, contact form emails will not be sent!")
	}
	contactService = contact.NewService(gitStorageService, eventBus, emailService, adminEmail, contactLogger)
	
	// Initialize bio service
	bioLogger := log.New(os.Stdout, "[bio] ", log.LstdFlags)
	bioService = bio.NewService(bioLogger)
	
	// Start services
	ctx := context.Background()
	if err := eventBus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}
	
	// Start the blog service (cast to registry.Service interface)
	if starter, ok := blogService.(interface{ Start(context.Context) error }); ok {
		if err := starter.Start(ctx); err != nil {
			return fmt.Errorf("failed to start blog service: %w", err)
		}
	}
	
	// Load posts into the legacy blogPosts variable for compatibility
	posts := blogService.GetAll(ctx)
	blogPosts = make([]BlogPost, len(posts))
	for i, p := range posts {
		blogPosts[i] = BlogPost{
			Slug:        p.Slug,
			Title:       p.Title,
			Date:        p.Date,
			Summary:     p.Summary,
			Content:     p.Content,
			ReadingTime: p.ReadingTime,
			Tags:        p.Tags,
			FileName:    p.FileName,
		}
	}
	
	log.Printf("Blog service initialized with %d posts", len(blogPosts))
	return nil
}

// Deprecated: loadBlogPosts is replaced by initializeBlogService
func loadBlogPosts() {
	blogPosts = []BlogPost{}
	
	// Skip loading blog posts if disabled (for faster development)
	if !siteConfig.BlogEnabled {
		log.Printf("Blog disabled - skipping blog post loading")
		return
	}
	
	// Load markdown files from embedded filesystem
	files, err := fs.ReadDir(blogFS, "content/blog")
	if err != nil {
		log.Printf("Warning: Could not read blog directory: %v", err)
		return
	}
	
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		
		post, err := loadMarkdownPost(file.Name())
		if err != nil {
			log.Printf("Warning: Could not load blog post %s: %v", file.Name(), err)
			continue
		}
		
		blogPosts = append(blogPosts, *post)
	}
	
	// Sort posts by date (newest first)
	sort.Slice(blogPosts, func(i, j int) bool {
		return blogPosts[i].Date.After(blogPosts[j].Date)
	})
	
	log.Printf("Loaded %d blog posts", len(blogPosts))
}

func loadMarkdownPost(filename string) (*BlogPost, error) {
	// Read markdown file from embedded filesystem
	content, err := fs.ReadFile(blogFS, "content/blog/"+filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}
	
	// Parse frontmatter and content
	frontmatter, markdownContent, err := parseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("could not parse frontmatter: %w", err)
	}
	
	// Convert markdown to HTML
	htmlContent := markdownToHTML(markdownContent)
	
	// Generate slug from filename
	slug := strings.TrimSuffix(filename, ".md")
	
	// Calculate reading time if not provided
	readingTime := frontmatter.ReadingTime
	if readingTime == 0 {
		readingTime = calculateReadingTime(string(markdownContent))
	}
	
	return &BlogPost{
		Slug:        slug,
		Title:       frontmatter.Title,
		Date:        frontmatter.Date,
		Summary:     frontmatter.Summary,
		Content:     template.HTML(htmlContent),
		ReadingTime: readingTime,
		Tags:        frontmatter.Tags,
		FileName:    filename,
	}, nil
}

func parseFrontmatter(content []byte) (*BlogFrontmatter, []byte, error) {
	// Check if content starts with frontmatter delimiter
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return nil, nil, fmt.Errorf("missing frontmatter delimiter")
	}
	
	// Find the end of frontmatter
	endDelimiter := []byte("\n---\n")
	endIndex := bytes.Index(content[4:], endDelimiter)
	if endIndex == -1 {
		return nil, nil, fmt.Errorf("missing frontmatter end delimiter")
	}
	
	// Extract frontmatter and content
	frontmatterBytes := content[4 : endIndex+4]
	markdownContent := content[endIndex+8:] // Skip past "\n---\n"
	
	// Parse YAML frontmatter
	var frontmatter BlogFrontmatter
	if err := yaml.Unmarshal(frontmatterBytes, &frontmatter); err != nil {
		return nil, nil, fmt.Errorf("could not parse YAML: %w", err)
	}
	
	return &frontmatter, markdownContent, nil
}

func markdownToHTML(mdContent []byte) string {
	// Configure markdown parser
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	
	// Configure HTML renderer with syntax highlighting
	htmlFlags := mdhtml.CommonFlags | mdhtml.HrefTargetBlank
	opts := mdhtml.RendererOptions{
		Flags: htmlFlags,
		RenderNodeHook: chromaRenderHook,
	}
	renderer := mdhtml.NewRenderer(opts)
	
	// Convert markdown to HTML
	return string(markdown.ToHTML(mdContent, p, renderer))
}

// chromaRenderHook provides syntax highlighting for code blocks and Mermaid diagram support
func chromaRenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	if code, ok := node.(*ast.CodeBlock); ok && entering {
		// Get the language from the code block info
		language := ""
		if code.Info != nil {
			language = string(code.Info)
		}
		
		// Handle Mermaid diagrams
		if language == "mermaid" {
			// Try server-side rendering first, fallback to client-side
			if svgContent := renderMermaidServerSide(string(code.Literal)); svgContent != "" {
				// Server-side rendering successful
				w.Write([]byte(`<div class="mermaid-container mermaid-ssr">`))
				w.Write([]byte(svgContent))
				w.Write([]byte("</div>"))
			} else {
				// Fallback to client-side rendering
				w.Write([]byte(`<div class="mermaid-container mermaid-csr"><div class="mermaid">`))
				w.Write(code.Literal)
				w.Write([]byte("</div></div>"))
			}
			return ast.GoToNext, true
		}
		
		// Get lexer for the language
		lexer := lexers.Get(language)
		if lexer == nil {
			lexer = lexers.Fallback
		}
		
		// Configure formatter with cyberpunk theme
		formatter := html.New(html.WithClasses(true), html.TabWidth(2))
		style := styles.Get("monokai")
		if style == nil {
			style = styles.Fallback
		}
		
		// Create iterator from code content
		iterator, err := lexer.Tokenise(nil, string(code.Literal))
		if err != nil {
			// Fallback to plain text
			w.Write([]byte("<pre><code>"))
			w.Write(code.Literal)
			w.Write([]byte("</code></pre>"))
			return ast.GoToNext, true
		}
		
		// Format the code
		err = formatter.Format(w, style, iterator)
		if err != nil {
			// Fallback to plain text
			w.Write([]byte("<pre><code>"))
			w.Write(code.Literal)
			w.Write([]byte("</code></pre>"))
		}
		
		return ast.GoToNext, true
	}
	
	return ast.GoToNext, false
}

// Cache for server-side rendered Mermaid diagrams
var mermaidCache = make(map[string]string)
var mermaidCacheMutex sync.RWMutex

// renderMermaidServerSide attempts to render Mermaid diagram on server-side
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func renderMermaidServerSide(mermaidCode string) string {
	// Generate cache key from content hash
	hash := md5.Sum([]byte(mermaidCode))
	cacheKey := hex.EncodeToString(hash[:])
	
	// Check cache first
	mermaidCacheMutex.RLock()
	if cached, exists := mermaidCache[cacheKey]; exists {
		mermaidCacheMutex.RUnlock()
		return cached
	}
	mermaidCacheMutex.RUnlock()
	
	// Create temporary directory for Mermaid CLI
	tempDir := "data/mermaid-temp"
	os.MkdirAll(tempDir, 0755)
	
	inputFile := filepath.Join(tempDir, cacheKey+".mmd")
	outputFile := filepath.Join(tempDir, cacheKey+".svg")
	
	// Write Mermaid input file
	if err := os.WriteFile(inputFile, []byte(mermaidCode), 0644); err != nil {
		log.Printf("Failed to write Mermaid input file: %v", err)
		return ""
	}
	
	// Try to execute mmdc (Mermaid CLI) with timeout - using simple dark theme for now
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "mmdc", 
		"-i", inputFile, 
		"-o", outputFile,
		"-t", "dark",
		"-b", "transparent")
	
	// Capture both stdout and stderr for debugging
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		log.Printf("Mermaid CLI failed (falling back to client-side):")
		log.Printf("  Error: %v", err)
		log.Printf("  Command: mmdc -i %s -o %s -t dark -b transparent", inputFile, outputFile)
		log.Printf("  Stdout: %s", stdout.String())
		log.Printf("  Stderr: %s", stderr.String())
		log.Printf("  Input file exists: %v", fileExists(inputFile))
		log.Printf("  Output dir exists: %v", fileExists(filepath.Dir(outputFile)))
		// Clean up temp files
		os.Remove(inputFile)
		os.Remove(outputFile)
		return ""
	}
	
	// Read generated SVG
	svgContent, err := os.ReadFile(outputFile)
	if err != nil {
		log.Printf("Failed to read generated SVG: %v", err)
		os.Remove(inputFile)
		os.Remove(outputFile)
		return ""
	}
	
	// Clean up temp files
	os.Remove(inputFile)
	os.Remove(outputFile)
	
	svgString := string(svgContent)
	
	// Cache the result
	mermaidCacheMutex.Lock()
	mermaidCache[cacheKey] = svgString
	mermaidCacheMutex.Unlock()
	
	return svgString
}

func calculateReadingTime(text string) int {
	// Average reading speed: 200 words per minute
	words := len(strings.Fields(text))
	minutes := int(math.Ceil(float64(words) / 200.0))
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}

func initializeTimeSlots() {
	// Create available slots for next 30 days
	// Monday-Friday, 10am-5pm, 1-hour slots
	now := time.Now()
	for i := 0; i < 30; i++ {
		date := now.AddDate(0, 0, i)
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue
		}

		dateStr := date.Format("2006-01-02")
		for hour := 10; hour < 17; hour++ {
			timeStr := fmt.Sprintf("%02d:00", hour)
			slotID := fmt.Sprintf("%s-%s", dateStr, timeStr)

			if _, exists := timeSlots[slotID]; !exists {
				timeSlots[slotID] = &TimeSlot{
					ID:        slotID,
					Date:      dateStr,
					Time:      timeStr,
					Available: true,
					Booked:    false,
					CreatedAt: now,
				}
			}
		}
	}
}

func loadBookings() {
	data, err := os.ReadFile(bookingsFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error loading bookings: %v", err)
		}
		return
	}

	var slots map[string]*TimeSlot
	if err := json.Unmarshal(data, &slots); err != nil {
		log.Printf("Error parsing bookings: %v", err)
		return
	}

	// Merge with existing slots
	for id, slot := range slots {
		timeSlots[id] = slot
	}
}

func saveBookings() {
	// Create data directory if it doesn't exist
	os.MkdirAll(filepath.Dir(bookingsFile), 0755)

	data, err := json.MarshalIndent(timeSlots, "", "  ")
	if err != nil {
		log.Printf("Error marshaling bookings: %v", err)
		return
	}

	if err := os.WriteFile(bookingsFile, data, 0644); err != nil {
		log.Printf("Error saving bookings: %v", err)
	}
}

// HTMX content-only handlers
func homeContentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get brief bio
	var bioBrief *bio.Bio
	if bioService != nil {
		var err error
		bioBrief, err = bioService.GetBrief(ctx)
		if err != nil {
			log.Printf("Warning: Failed to load brief bio: %v", err)
		}
	}
	
	data := struct {
		Config   *SiteConfig
		AppConfig *config.SiteConfig
		BioBrief *bio.Bio
	}{
		Config:   siteConfig,
		AppConfig: appConfig,
		BioBrief: bioBrief,
	}

	w.Header().Set("Content-Type", "text/html")
	
	if err := templates.ExecuteTemplate(w, "home-content", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func blogContentHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Posts      []BlogPost
		Config     *SiteConfig
		AppConfig  *config.SiteConfig
		WorkConfig *config.WorkConfig
		BlogConfig *blog.BlogConfig
	}{
		Posts:      blogPosts,
		Config:     siteConfig,
		AppConfig:  appConfig,
		WorkConfig: workConfig,
		BlogConfig: blogService.GetBlogConfig(),
	}

	w.Header().Set("Content-Type", "text/html")
	
	if err := templates.ExecuteTemplate(w, "blog-content", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get full bio
	var fullBio *bio.Bio
	if bioService != nil {
		var err error
		fullBio, err = bioService.GetFull(ctx)
		if err != nil {
			log.Printf("Warning: Failed to load full bio: %v", err)
		}
	}
	
	data := struct {
		Title     string
		Page      string
		Config    *SiteConfig
		AppConfig *config.SiteConfig
		Bio       *bio.Bio
	}{
		Title:     "About Lance Rogers - Blockhead Consulting",
		Page:      "about",
		Config:    siteConfig,
		AppConfig: appConfig,
		Bio:       fullBio,
	}

	// Use ExecuteTemplate directly with the specific page template
	if err := templates.ExecuteTemplate(w, "about-full.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func aboutContentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get full bio
	var fullBio *bio.Bio
	if bioService != nil {
		var err error
		fullBio, err = bioService.GetFull(ctx)
		if err != nil {
			log.Printf("Warning: Failed to load full bio: %v", err)
		}
	}
	
	data := struct {
		Config    *SiteConfig
		AppConfig *config.SiteConfig
		Bio       *bio.Bio
	}{
		Config:    siteConfig,
		AppConfig: appConfig,
		Bio:       fullBio,
	}

	w.Header().Set("Content-Type", "text/html")
	
	if err := templates.ExecuteTemplate(w, "about-content", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func calendarContentHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Config    *SiteConfig
		AppConfig *config.SiteConfig
	}{
		Config:    siteConfig,
		AppConfig: appConfig,
	}
	
	w.Header().Set("Content-Type", "text/html")
	
	if err := templates.ExecuteTemplate(w, "calendar-content", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func workHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title      string
		Page       string
		Config     *SiteConfig
		AppConfig  *config.SiteConfig
		WorkConfig *config.WorkConfig
	}{
		Title:      "Work Experience - Blockhead Consulting",
		Page:       "work",
		Config:     siteConfig,
		AppConfig:  appConfig,
		WorkConfig: workConfig,
	}

	// Use ExecuteTemplate directly with the specific page template
	if err := templates.ExecuteTemplate(w, "work-full.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func workContentHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Config     *SiteConfig
		AppConfig  *config.SiteConfig
		WorkConfig *config.WorkConfig
	}{
		Config:     siteConfig,
		AppConfig:  appConfig,
		WorkConfig: workConfig,
	}

	w.Header().Set("Content-Type", "text/html")
	
	if err := templates.ExecuteTemplate(w, "work-content", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Environment variable helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
