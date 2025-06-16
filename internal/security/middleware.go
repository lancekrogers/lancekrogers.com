package security

import (
	"fmt"
	"net/http"
	"strings"
)

// HeadersMiddleware adds security headers to all responses
func HeadersMiddleware(config *SecurityHeaders) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultSecurityHeaders()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Content Security Policy
			if config.CSPNonce != "" {
				cspPolicy := fmt.Sprintf(`
					default-src 'self';
					script-src 'self' 'nonce-%s' https://unpkg.com;
					style-src 'self' 'unsafe-inline';
					img-src 'self' data: https:;
					font-src 'self' data:;
					connect-src 'self';
					frame-ancestors 'none';
				`, config.CSPNonce)
				
				// Clean up the CSP policy (remove extra whitespace and newlines)
				cspPolicy = strings.ReplaceAll(cspPolicy, "\n", "")
				cspPolicy = strings.ReplaceAll(cspPolicy, "\t", "")
				cspPolicy = strings.Join(strings.Fields(cspPolicy), " ")
				
				w.Header().Set("Content-Security-Policy", cspPolicy)
			}

			// HTTP Strict Transport Security
			if config.EnableHSTS {
				hstsValue := fmt.Sprintf("max-age=%d; includeSubDomains", config.HSTSMaxAge)
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// X-Content-Type-Options
			if config.EnableContentTypeNoSniff {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}

			// X-Frame-Options
			if config.EnableFrameOptions {
				w.Header().Set("X-Frame-Options", config.FrameOptions)
			}

			// X-XSS-Protection
			if config.EnableXSSProtection {
				w.Header().Set("X-XSS-Protection", "1; mode=block")
			}

			// Additional security headers
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			
			// Permissions-Policy: configurable policy for blocking sensitive features
			if config.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", config.PermissionsPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware applies rate limiting based on client IP
func RateLimitMiddleware(rateLimiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := ExtractClientIP(r)
			
			if !rateLimiter.IsAllowed(clientIP) {
				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// InputValidationMiddleware provides basic input validation and sanitization
func InputValidationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size to prevent DoS
			r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB limit

			// Check for common attack patterns in URL path and query
			fullURL := r.URL.String()
			suspiciousPatterns := []string{
				"../", "..\\", "..",
				"<script", "</script",
				"javascript:", "vbscript:",
				"onload=", "onerror=",
				"eval(", "alert(",
			}

			lowerURL := strings.ToLower(fullURL)
			for _, pattern := range suspiciousPatterns {
				if strings.Contains(lowerURL, pattern) {
					http.Error(w, "Invalid request", http.StatusBadRequest)
					return
				}
			}

			// Check User-Agent header for basic bot detection
			userAgent := r.Header.Get("User-Agent")
			if userAgent == "" {
				// Allow empty user agents but log them
				// In production, you might want to be more strict
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityMiddleware combines multiple security middleware into one
func SecurityMiddleware(config *Config) func(http.Handler) http.Handler {
	// Use consulting-friendly headers by default, override CSP nonce from config
	headerConfig := ConsultingWebsiteHeaders()
	headerConfig.CSPNonce = config.CSPNonce

	return func(next http.Handler) http.Handler {
		var handler http.Handler = next

		// Apply middleware in reverse order (innermost first)
		handler = HeadersMiddleware(headerConfig)(handler)
		handler = InputValidationMiddleware()(handler)
		
		if config.RateLimiter != nil {
			handler = RateLimitMiddleware(config.RateLimiter)(handler)
		}

		return handler
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing headers
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// DetectSuspiciousActivity checks for common attack patterns in requests
func DetectSuspiciousActivity(r *http.Request) bool {
	// Check for common attack patterns
	suspiciousPatterns := []string{
		"<script", "</script>", "javascript:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=",
		"../", "..\\", "/etc/passwd", "/proc/",
		"'or'1'='1", "union select", "drop table",
		"<?php", "<?=", "<%", "%>",
	}

	userAgent := strings.ToLower(r.UserAgent())
	url := strings.ToLower(r.URL.String())
	
	// Check URL and User-Agent for suspicious patterns
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(url, pattern) || strings.Contains(userAgent, pattern) {
			return true
		}
	}

	// Check for suspicious user agents
	suspiciousUAs := []string{
		"sqlmap", "nikto", "burp", "nessus", "openvas",
		"metasploit", "nmap", "dirbuster", "gobuster",
	}

	for _, ua := range suspiciousUAs {
		if strings.Contains(userAgent, ua) {
			return true
		}
	}

	return false
}