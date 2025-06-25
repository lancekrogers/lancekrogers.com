package server

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"blockhead.consulting/internal/security"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
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

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// HTMXMiddleware adds HTMX-specific headers
func HTMXMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			// Add HTMX-specific response headers
			w.Header().Set("HX-Push-Url", "false")
		}
		next.ServeHTTP(w, r)
	})
}

// CacheControlMiddleware adds cache control headers
func CacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set cache control headers based on content type
		if strings.HasPrefix(r.URL.Path, "/static/") {
			// Static assets can be cached for longer
			w.Header().Set("Cache-Control", "public, max-age=3600")
		} else if strings.HasPrefix(r.URL.Path, "/api/") {
			// API responses should not be cached
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		} else {
			// HTML pages - short cache
			w.Header().Set("Cache-Control", "public, max-age=300")
		}
		next.ServeHTTP(w, r)
	})
}

// RequestSizeMiddleware limits request body size
func RequestSizeMiddleware(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityLoggingMiddleware logs security-relevant events
func SecurityLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		clientIP := security.ExtractClientIP(r)

		// Log security-relevant events
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
			log.Printf("SECURITY: %s %s from %s User-Agent: %s",
				r.Method, r.URL.Path, clientIP, r.UserAgent())
		}

		// Detect potential attacks
		if security.DetectSuspiciousActivity(r) {
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
