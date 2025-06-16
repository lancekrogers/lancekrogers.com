package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHeadersMiddleware(t *testing.T) {
	config := &SecurityHeaders{
		CSPNonce:                "test-nonce-123",
		EnableHSTS:              true,
		HSTSMaxAge:              31536000,
		EnableXSSProtection:     true,
		EnableContentTypeNoSniff: true,
		EnableFrameOptions:      true,
		FrameOptions:            "DENY",
	}

	handler := HeadersMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check CSP header
	csp := w.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "nonce-test-nonce-123") {
		t.Errorf("CSP header should contain nonce, got: %s", csp)
	}
	if !strings.Contains(csp, "default-src 'self'") {
		t.Errorf("CSP header should contain default-src 'self', got: %s", csp)
	}

	// Check HSTS header
	hsts := w.Header().Get("Strict-Transport-Security")
	expectedHSTS := "max-age=31536000; includeSubDomains"
	if hsts != expectedHSTS {
		t.Errorf("Expected HSTS header %s, got %s", expectedHSTS, hsts)
	}

	// Check X-Content-Type-Options
	ctOptions := w.Header().Get("X-Content-Type-Options")
	if ctOptions != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options nosniff, got %s", ctOptions)
	}

	// Check X-Frame-Options
	frameOptions := w.Header().Get("X-Frame-Options")
	if frameOptions != "DENY" {
		t.Errorf("Expected X-Frame-Options DENY, got %s", frameOptions)
	}

	// Check X-XSS-Protection
	xssProtection := w.Header().Get("X-XSS-Protection")
	if xssProtection != "1; mode=block" {
		t.Errorf("Expected X-XSS-Protection '1; mode=block', got %s", xssProtection)
	}

	// Check Referrer-Policy
	referrerPolicy := w.Header().Get("Referrer-Policy")
	if referrerPolicy != "strict-origin-when-cross-origin" {
		t.Errorf("Expected Referrer-Policy 'strict-origin-when-cross-origin', got %s", referrerPolicy)
	}
}

func TestHeadersMiddlewareWithNilConfig(t *testing.T) {
	handler := HeadersMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should still set basic security headers even with nil config
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Should set X-Content-Type-Options even with nil config")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	config := &RateLimiterConfig{
		MaxRequests:   2,
		Window:        time.Minute,
		CleanupPeriod: time.Minute,
		BlockDuration: 5 * time.Minute,
	}

	rateLimiter := NewRateLimiter(config)
	
	handler := RateLimitMiddleware(rateLimiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	// First two requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should succeed, got status %d", i+1, w.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Third request should be rate limited, got status %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Rate limit exceeded") {
		t.Error("Response should contain rate limit message")
	}
}

func TestInputValidationMiddleware(t *testing.T) {
	handler := InputValidationMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "normal path",
			path:           "/normal/path",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "path traversal attempt",
			path:           "/../../etc/passwd",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "script injection attempt",
			path:           "/search?q=<script>alert('xss')</script>",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "javascript protocol",
			path:           "/link?url=javascript:alert(1)",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "event handler attempt",
			path:           "/test?param=onload=malicious()",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestSecurityMiddleware(t *testing.T) {
	nonce, err := GenerateNonce()
	if err != nil {
		t.Fatalf("Failed to generate nonce: %v", err)
	}

	config := &Config{
		CSPNonce: nonce,
		RateLimiter: NewRateLimiter(&RateLimiterConfig{
			MaxRequests:   10,
			Window:        time.Minute,
			CleanupPeriod: time.Minute,
			BlockDuration: 5 * time.Minute,
		}),
	}

	handler := SecurityMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}

	// Check that security headers are set
	if w.Header().Get("Content-Security-Policy") == "" {
		t.Error("CSP header should be set")
	}

	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("X-Frame-Options should be set to DENY")
	}
}

func TestCORSMiddleware(t *testing.T) {
	allowedOrigins := []string{"https://example.com", "https://test.com"}
	
	handler := CORSMiddleware(allowedOrigins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		method         string
		origin         string
		expectedOrigin string
		expectedStatus int
	}{
		{
			name:           "allowed origin",
			method:         "GET",
			origin:         "https://example.com",
			expectedOrigin: "https://example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "disallowed origin",
			method:         "GET",
			origin:         "https://evil.com",
			expectedOrigin: "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "preflight request",
			method:         "OPTIONS",
			origin:         "https://example.com",
			expectedOrigin: "https://example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no origin header",
			method:         "GET",
			origin:         "",
			expectedOrigin: "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if corsOrigin != tt.expectedOrigin {
				t.Errorf("Expected CORS origin %s, got %s", tt.expectedOrigin, corsOrigin)
			}

			// Check that CORS headers are always set
			methods := w.Header().Get("Access-Control-Allow-Methods")
			if methods == "" {
				t.Error("Access-Control-Allow-Methods should be set")
			}

			headers := w.Header().Get("Access-Control-Allow-Headers")
			if headers == "" {
				t.Error("Access-Control-Allow-Headers should be set")
			}
		})
	}
}

func TestCORSMiddlewareWithWildcard(t *testing.T) {
	allowedOrigins := []string{"*"}
	
	handler := CORSMiddleware(allowedOrigins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://any-origin.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if corsOrigin != "https://any-origin.com" {
		t.Errorf("Wildcard should allow any origin, got %s", corsOrigin)
	}
}

func TestDetectSuspiciousActivity(t *testing.T) {
	tests := []struct {
		name      string
		setupReq  func() *http.Request
		expected  bool
	}{
		{
			name: "normal request",
			setupReq: func() *http.Request {
				r := httptest.NewRequest("GET", "/about", nil)
				r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
				return r
			},
			expected: false,
		},
		{
			name: "script injection in URL",
			setupReq: func() *http.Request {
				r := httptest.NewRequest("GET", "/search?q=<script>alert('xss')</script>", nil)
				r.Header.Set("User-Agent", "Mozilla/5.0")
				return r
			},
			expected: true,
		},
		{
			name: "path traversal attempt",
			setupReq: func() *http.Request {
				r := httptest.NewRequest("GET", "/files/../../../etc/passwd", nil)
				r.Header.Set("User-Agent", "Mozilla/5.0")
				return r
			},
			expected: true,
		},
		{
			name: "SQL injection attempt",
			setupReq: func() *http.Request {
				r := httptest.NewRequest("GET", "/user?id=1'or'1'='1", nil)
				r.Header.Set("User-Agent", "Mozilla/5.0")
				return r
			},
			expected: true,
		},
		{
			name: "malicious user agent",
			setupReq: func() *http.Request {
				r := httptest.NewRequest("GET", "/", nil)
				r.Header.Set("User-Agent", "sqlmap/1.4.7")
				return r
			},
			expected: true,
		},
		{
			name: "scanning tool user agent",
			setupReq: func() *http.Request {
				r := httptest.NewRequest("GET", "/", nil)
				r.Header.Set("User-Agent", "nikto/2.1.6")
				return r
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			result := DetectSuspiciousActivity(req)
			if result != tt.expected {
				t.Errorf("DetectSuspiciousActivity() = %v, expected %v", result, tt.expected)
			}
		})
	}
}