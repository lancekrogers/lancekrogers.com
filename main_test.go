package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"blockhead.consulting/internal/app"
	"github.com/gorilla/mux"
)

// Test helper to create a test app instance
func createTestApp(t *testing.T) (*app.App, *mux.Router) {
	// Use embedded file systems from main.go
	testApp := app.New(templateFS, staticFS, blogFS)
	if err := testApp.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}
	if err := testApp.CreateHandler(); err != nil {
		t.Fatalf("Failed to create test handler: %v", err)
	}
	router, err := testApp.CreateRouter()
	if err != nil {
		t.Fatalf("Failed to create test router: %v", err)
	}
	return testApp, router
}

func TestHomeHandler(t *testing.T) {
	_, router := createTestApp(t)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

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
	_, router := createTestApp(t)

	req, err := http.NewRequest("GET", "/content/home", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

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
	_, router := createTestApp(t)

	req, err := http.NewRequest("GET", "/content/blog", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

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
	// This test is skipped because calendar is disabled in current config
	// If calendar gets re-enabled, update this test
	t.Skip("Calendar functionality is currently disabled")
}

func TestSlotsAPI(t *testing.T) {
	// This test is skipped because calendar/booking is disabled in current config
	// If calendar gets re-enabled, update this test
	t.Skip("Calendar/booking functionality is currently disabled")
}

func TestRouting(t *testing.T) {
	_, router := createTestApp(t)

	testCases := []struct {
		method string
		path   string
		expectedStatus int
		description    string
	}{
		{"GET", "/", http.StatusOK, "home page"},
		{"GET", "/content/home", http.StatusOK, "home content"},
		{"GET", "/content/blog", http.StatusOK, "blog content"},
		{"GET", "/blog", http.StatusOK, "blog page"},
		{"GET", "/about", http.StatusOK, "about page"},
		{"GET", "/content/about", http.StatusOK, "about content"},
		{"GET", "/health", http.StatusOK, "health check"},
		{"GET", "/nonexistent", http.StatusNotFound, "nonexistent page"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("route %s %s returned wrong status code: got %v want %v",
					tc.method, tc.path, status, tc.expectedStatus)
			}
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	_, router := createTestApp(t)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Test that security headers are present (middleware is applied through router)
	expectedHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
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
	// This test is skipped because booking functionality is disabled
	// If booking gets re-enabled, update this test to use the contact service validation
	t.Skip("Booking functionality is currently disabled")
}

func TestRateLimiting(t *testing.T) {
	testApp, _ := createTestApp(t)

	// Access the rate limiter through the app's dependencies
	// This tests that rate limiting is properly configured
	testIP := "192.168.1.100"
	
	// Test that we can make reasonable number of requests
	// Note: Rate limiting configuration may vary, so just test basic functionality
	for i := 0; i < 10; i++ {
		req, err := http.NewRequest("GET", "/health", nil)
		if err != nil {
			t.Fatal(err)
		}
		// Set remote addr to simulate different IPs for rate limiting
		req.RemoteAddr = testIP + ":12345"
		
		router, _ := testApp.CreateRouter()
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Request %d should be allowed but got status %d", i+1, status)
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
	
	// Test configuration through app initialization
	testApp, _ := createTestApp(t)
	
	// Test that app initializes correctly with expected configuration
	if testApp == nil {
		t.Fatal("Test app should be initialized")
	}
	
	// Test basic functionality works (which means config loaded correctly)
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	router, _ := testApp.CreateRouter()
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health check failed, configuration may be incorrect. Got status: %d", status)
	}
}

func TestCalendarDisabling(t *testing.T) {
	// Test calendar disabled configuration
	os.Setenv("CALENDAR_ENABLED", "false")
	defer os.Unsetenv("CALENDAR_ENABLED")
	
	// Test that calendar routes are not accessible
	_, router := createTestApp(t)
	
	req, err := http.NewRequest("GET", "/calendar", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	// Should return 404 because calendar is disabled
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Calendar route should not be available when disabled. Got status: %d", status)
	}
}


func TestMobileMenuFunctionality(t *testing.T) {
	t.Run("Mobile menu elements present", func(t *testing.T) {
		_, router := createTestApp(t)

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

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

