package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"blockhead.consulting/internal/security"
	"github.com/gorilla/mux"
)

func TestHomeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homeHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check if response contains expected content
	body := rr.Body.String()
	if !strings.Contains(body, "BLOCKHEAD CONSULTING") {
		if len(body) > 500 {
			body = body[:500]
		}
		t.Errorf("handler returned unexpected body: missing 'BLOCKHEAD CONSULTING'. Got: %s", body)
	}
}

func TestHomeContentHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/content/home", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homeContentHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()
	
	// Check for key home page elements
	if !strings.Contains(body, "hero") {
		t.Errorf("home content missing hero section")
	}
	if !strings.Contains(body, "BLOCKHEAD CONSULTING") {
		t.Errorf("home content missing main heading")
	}
	if !strings.Contains(body, "services") {
		t.Errorf("home content missing services section")
	}
	if !strings.Contains(body, `id="services"`) {
		t.Errorf("home content missing services anchor for navigation")
	}
}

func TestBlogContentHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/content/blog", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(blogContentHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()
	
	// Check for blog page elements
	if !strings.Contains(body, "Technical Insights") {
		t.Errorf("blog content missing page title")
	}
	if !strings.Contains(body, "blog-section") {
		t.Errorf("blog content missing blog section")
	}
}

func TestCalendarContentHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/content/calendar", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(calendarContentHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()
	
	// Check for calendar page elements
	if !strings.Contains(body, "Book a Consultation") {
		t.Errorf("calendar content missing page title")
	}
	if !strings.Contains(body, "calendar-section") {
		t.Errorf("calendar content missing calendar section")
	}
	if !strings.Contains(body, "time-slots") {
		t.Errorf("calendar content missing time slots")
	}
}

func TestSlotsAPI(t *testing.T) {
	// Initialize time slots for testing
	initializeTimeSlots()

	req, err := http.NewRequest("GET", "/api/slots", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(slotsHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check content type
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, "application/json")
	}

	// Check if response contains JSON array
	body := strings.TrimSpace(rr.Body.String())
	if !strings.HasPrefix(body, "[") || !strings.HasSuffix(body, "]") {
		t.Errorf("handler returned non-JSON array response: %s", body)
	}
}

func TestRouting(t *testing.T) {
	r := mux.NewRouter()
	
	// Add the same routes as main
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/content/home", homeContentHandler).Methods("GET")
	r.HandleFunc("/content/blog", blogContentHandler).Methods("GET")
	r.HandleFunc("/content/calendar", calendarContentHandler).Methods("GET")
	r.HandleFunc("/api/slots", slotsHandler).Methods("GET")

	testCases := []struct {
		method string
		path   string
		expectedStatus int
	}{
		{"GET", "/", http.StatusOK},
		{"GET", "/content/home", http.StatusOK},
		{"GET", "/content/blog", http.StatusOK},
		{"GET", "/content/calendar", http.StatusOK},
		{"GET", "/api/slots", http.StatusOK},
		{"GET", "/nonexistent", http.StatusNotFound},
	}

	for _, tc := range testCases {
		req, err := http.NewRequest(tc.method, tc.path, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != tc.expectedStatus {
			t.Errorf("route %s %s returned wrong status code: got %v want %v",
				tc.method, tc.path, status, tc.expectedStatus)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	// Initialize security for testing
	initializeSecurity()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	
	// Create security headers config with nonce for testing
	testNonce, _ := security.GenerateNonce()
	securityHeaders := security.ConsultingWebsiteHeaders()
	securityHeaders.CSPNonce = testNonce
	
	handler := security.HeadersMiddleware(securityHeaders)(http.HandlerFunc(homeHandler))

	handler.ServeHTTP(rr, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":           "nosniff",
		"X-Frame-Options":                  "DENY",
		"X-XSS-Protection":                 "1; mode=block",
		"Strict-Transport-Security":        "max-age=31536000; includeSubDomains",
		"Referrer-Policy":                  "strict-origin-when-cross-origin",
		"Permissions-Policy":               "geolocation=(), usb=(), fullscreen=(self), payment=(self)",
	}

	for header, expectedValue := range expectedHeaders {
		if actualValue := rr.Header().Get(header); actualValue != expectedValue {
			t.Errorf("Security header %s: got %v want %v", header, actualValue, expectedValue)
		}
	}

	// Check CSP header exists
	if csp := rr.Header().Get("Content-Security-Policy"); csp == "" {
		t.Error("Content-Security-Policy header is missing")
	}
}

func TestInputValidation(t *testing.T) {
	testCases := []struct {
		name    string
		request BookingRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: BookingRequest{
				SlotID:      "2025-05-26-10:00",
				Name:        "John Doe",
				Email:       "john@example.com",
				ServiceType: "crypto-infrastructure",
			},
			valid: true,
		},
		{
			name: "invalid email",
			request: BookingRequest{
				SlotID:      "2025-05-26-10:00",
				Name:        "John Doe",
				Email:       "invalid-email",
				ServiceType: "crypto-infrastructure",
			},
			valid: false,
		},
		{
			name: "XSS attempt in name",
			request: BookingRequest{
				SlotID:      "2025-05-26-10:00",
				Name:        "<script>alert('xss')</script>",
				Email:       "john@example.com",
				ServiceType: "crypto-infrastructure",
			},
			valid: false,
		},
		{
			name: "SQL injection attempt",
			request: BookingRequest{
				SlotID:      "2025-05-26-10:00",
				Name:        "John'; DROP TABLE users; --",
				Email:       "john@example.com",
				ServiceType: "crypto-infrastructure",
			},
			valid: false,
		},
		{
			name: "invalid service type",
			request: BookingRequest{
				SlotID:      "2025-05-26-10:00",
				Name:        "John Doe",
				Email:       "john@example.com",
				ServiceType: "malicious-service",
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBookingRequest(&tc.request)
			if tc.valid && err != nil {
				t.Errorf("Expected valid request but got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Errorf("Expected invalid request but validation passed")
			}
		})
	}
}

func TestRateLimiting(t *testing.T) {
	// Initialize security for testing
	initializeSecurity()

	testIP := "192.168.1.100"
	rateLimiter := securityConfig.RateLimiter

	// Test normal usage - should be allowed (new default is 500 requests per minute)
	for i := 0; i < 400; i++ {
		if !rateLimiter.IsAllowed(testIP) {
			t.Errorf("Request %d should be allowed but was blocked", i+1)
		}
	}

	// Test rate limit - should start blocking after 500 requests
	for i := 400; i < 520; i++ {
		allowed := rateLimiter.IsAllowed(testIP)
		if i < 500 && !allowed {
			t.Errorf("Request %d should be allowed but was blocked", i+1)
		}
		if i >= 500 && allowed {
			t.Errorf("Request %d should be blocked but was allowed", i+1)
		}
	}
}

func TestConfiguration(t *testing.T) {
	// Clear any existing environment variables for clean test
	os.Unsetenv("CALENDAR_ENABLED")
	os.Unsetenv("ENVIRONMENT")
	defer func() {
		// Restore any env vars that might be needed for other tests
		os.Unsetenv("CALENDAR_ENABLED")
		os.Unsetenv("ENVIRONMENT")
	}()
	
	// Test default configuration
	initializeConfig()
	
	if siteConfig == nil {
		t.Fatal("Site config should be initialized")
	}
	
	// Calendar is currently disabled in site.yml
	if siteConfig.CalendarEnabled {
		t.Error("Calendar should be disabled as per site.yml configuration")
	}
	
	if siteConfig.Environment != "development" {
		t.Errorf("Expected development environment, got %s", siteConfig.Environment)
	}
	
	if siteConfig.SiteName != "Blockhead Consulting" {
		t.Errorf("Expected 'Blockhead Consulting', got %s", siteConfig.SiteName)
	}
}

func TestCalendarDisabling(t *testing.T) {
	// Test calendar disabled configuration
	os.Setenv("CALENDAR_ENABLED", "false")
	defer os.Unsetenv("CALENDAR_ENABLED")
	
	initializeConfig()
	
	if siteConfig.CalendarEnabled {
		t.Error("Calendar should be disabled when CALENDAR_ENABLED=false")
	}
}

func TestNavigationWorkflows(t *testing.T) {
	// Initialize config for testing
	initializeConfig()
	
	t.Run("Services button navigation from home page", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(homeHandler)
		handler.ServeHTTP(rr, req)

		body := rr.Body.String()
		
		// Check that services section anchor exists in main content
		if !strings.Contains(body, `id="services"`) {
			t.Error("Services section anchor missing from home page")
		}
		
		// Check that services navigation link has correct data attribute in nav
		if !strings.Contains(body, `data-scroll-to="services"`) {
			t.Error("Services navigation link missing scroll-to data attribute")
		}
		
		// Check that navigation has proper structure
		if !strings.Contains(body, `href="/#services"`) {
			t.Error("Services navigation link missing proper href")
		}
	})
	
	t.Run("Services button navigation from blog page", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/blog", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(blogHandler)
		handler.ServeHTTP(rr, req)

		body := rr.Body.String()
		
		// Check that services link exists and points to home page with anchor  
		if !strings.Contains(body, `href="/#services"`) {
			t.Error("Services link missing correct href from blog page")
		}
		
		// Check that services link has HTMX attributes for SPA navigation
		if !strings.Contains(body, `hx-get="/content/home"`) {
			t.Error("Services link missing HTMX navigation from blog page")
		}
		
		if !strings.Contains(body, `data-scroll-to="services"`) {
			t.Error("Services link missing scroll-to data attribute from blog page")
		}
	})
	
	t.Run("Mobile navigation structure", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(homeHandler)
		handler.ServeHTTP(rr, req)

		body := rr.Body.String()
		
		// Check that mobile navigation elements exist
		if !strings.Contains(body, `class="mobile-nav"`) {
			t.Error("Mobile navigation wrapper missing")
		}
		
		if !strings.Contains(body, `id="hamburger-toggle"`) {
			t.Error("Hamburger toggle button missing")
		}
		
		if !strings.Contains(body, `id="mobile-menu"`) {
			t.Error("Mobile menu container missing")
		}
		
		// Check that desktop and mobile navigation are separate
		if !strings.Contains(body, `desktop-nav`) {
			t.Error("Desktop navigation wrapper missing")
		}
	})
	
	t.Run("Calendar redirect when disabled", func(t *testing.T) {
		// Test with calendar disabled
		os.Setenv("CALENDAR_ENABLED", "false")
		defer os.Unsetenv("CALENDAR_ENABLED")
		
		initializeConfig()
		
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(homeHandler)
		handler.ServeHTTP(rr, req)

		body := rr.Body.String()
		
		// Check that consultation button redirects to contact section
		if !strings.Contains(body, `href="#contact"`) {
			t.Error("Consultation button should redirect to contact when calendar disabled")
		}
		
		// Check that contact section has proper id
		if !strings.Contains(body, `id="contact"`) {
			t.Error("Contact section missing id anchor")
		}
		
		// Check that calendar links are removed from navigation
		if strings.Contains(body, "/calendar") {
			t.Error("Calendar links should be removed when calendar disabled")
		}
		
		// Check that footer doesn't have calendar links
		if strings.Contains(body, `href="/calendar"`) {
			t.Error("Footer should not have calendar links when calendar disabled")
		}
	})
	
	t.Run("Calendar enabled functionality", func(t *testing.T) {
		// Test with calendar enabled (default)
		os.Setenv("CALENDAR_ENABLED", "true")
		defer os.Unsetenv("CALENDAR_ENABLED")
		
		initializeConfig()
		
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(homeHandler)
		handler.ServeHTTP(rr, req)

		body := rr.Body.String()
		
		// Check that consultation button goes to calendar
		if !strings.Contains(body, `href="/calendar"`) {
			t.Error("Consultation button should link to calendar when enabled")
		}
		
		// Check that footer has Book Time link
		if !strings.Contains(body, `href="/calendar">Book Time</a>`) {
			t.Error("Footer should have Book Time link when calendar enabled")
		}
	})
}

func TestMobileMenuFunctionality(t *testing.T) {
	t.Run("Mobile menu elements present", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(homeHandler)
		handler.ServeHTTP(rr, req)

		body := rr.Body.String()
		
		// Check hamburger menu structure
		if !strings.Contains(body, `class="hamburger-line"`) {
			t.Error("Hamburger menu lines missing")
		}
		
		// Should have 3 hamburger lines
		lineCount := strings.Count(body, `class="hamburger-line"`)
		if lineCount != 3 {
			t.Errorf("Expected 3 hamburger lines, got %d", lineCount)
		}
		
		// Check mobile menu has correct links
		if !strings.Contains(body, `<div class="mobile-menu" id="mobile-menu">`) {
			t.Error("Mobile menu container missing proper structure")
		}
	})
}

func TestServicesScrollingWorkflow(t *testing.T) {
	t.Run("Services section anchor validation", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/content/home", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(homeContentHandler)
		handler.ServeHTTP(rr, req)

		body := rr.Body.String()
		
		// Validate services section has proper anchor ID
		if !strings.Contains(body, `<section id="services" class="services">`) {
			t.Error("Services section missing proper ID anchor for scrolling")
		}
		
		// Check that section has substantial content to scroll to
		if !strings.Contains(body, "Crypto Infrastructure") || !strings.Contains(body, "AI/LLM Consulting") {
			t.Error("Services section missing expected content")
		}
	})
	
	t.Run("Cross-page services navigation structure", func(t *testing.T) {
		// Test from blog page
		req, err := http.NewRequest("GET", "/blog", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(blogHandler)
		handler.ServeHTTP(rr, req)

		body := rr.Body.String()
		
		// Verify services link has all required attributes for proper navigation
		if !strings.Contains(body, `href="/#services"`) || 
		   !strings.Contains(body, `hx-get="/content/home"`) || 
		   !strings.Contains(body, `data-scroll-to="services"`) {
			t.Error("Services link from blog missing required navigation attributes")
		}
	})
}