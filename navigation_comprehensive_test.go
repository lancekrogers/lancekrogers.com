package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"blockhead.consulting/internal/security"
	"github.com/gorilla/mux"
)

func TestComprehensiveNavigation(t *testing.T) {
	// Initialize everything
	initializeTestEnvironment(t)
	
	// Create router
	r := mux.NewRouter()
	setupRoutes(r)
	
	// Test cases for all navigation scenarios
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedContent []string
		notExpected    []string
	}{
		{
			name:           "Home page loads correctly",
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"BLOCKHEAD CONSULTING",
				"Technical Expertise",
				"Core Languages",
				"Blockchain",
				"AI Engineering",
			},
			notExpected: []string{
				"Work Experience", // This should not appear on home page
			},
		},
		{
			name:           "Work page loads correctly",
			method:         "GET",
			path:           "/work",
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"Work Experience",
				"Bank of America",
				"Mythical Games",
				"FinTech &amp; Enterprise",
			},
			notExpected: []string{
				"BLOCKHEAD CONSULTING", // Hero title shouldn't appear on work page
				"Technical Expertise",   // This is only on home page
			},
		},
		{
			name:           "About page loads correctly",
			method:         "GET",
			path:           "/about",
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"About Lance Rogers",
				"Strategic Systems Architect",
			},
			notExpected: []string{
				"Work Experience",
				"Technical Expertise",
			},
		},
		{
			name:           "Blog page loads correctly",
			method:         "GET",
			path:           "/blog",
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"Blog",
			},
		},
		{
			name:           "Home content endpoint",
			method:         "GET",
			path:           "/content/home",
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"Technical Expertise",
				"Core Languages",
			},
		},
		{
			name:           "Work content endpoint",
			method:         "GET",
			path:           "/content/work",
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"Work Experience",
				"Bank of America",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			body := rr.Body.String()

			// Check expected content
			for _, expected := range tt.expectedContent {
				if !strings.Contains(body, expected) {
					t.Errorf("handler returned body without expected content '%s'", expected)
				}
			}

			// Check content that should NOT be present
			for _, notExpected := range tt.notExpected {
				if strings.Contains(body, notExpected) {
					t.Errorf("handler returned body with unexpected content '%s'", notExpected)
				}
			}
		})
	}
}

func TestPageRefreshNavigation(t *testing.T) {
	// Initialize everything
	initializeTestEnvironment(t)
	
	// Create router
	r := mux.NewRouter()
	setupRoutes(r)
	
	// Test multiple refreshes don't cause navigation issues
	pages := []struct {
		path            string
		expectedContent string
	}{
		{"/", "BLOCKHEAD CONSULTING"},
		{"/work", "Work Experience"},
		{"/about", "About Lance Rogers"},
		{"/blog", "Blog"},
	}

	for _, page := range pages {
		t.Run("Multiple refreshes of "+page.path, func(t *testing.T) {
			// Refresh same page 3 times
			for i := 0; i < 3; i++ {
				req, err := http.NewRequest("GET", page.path, nil)
				if err != nil {
					t.Fatal(err)
				}

				rr := httptest.NewRecorder()
				r.ServeHTTP(rr, req)

				if status := rr.Code; status != http.StatusOK {
					t.Errorf("Refresh %d: handler returned wrong status code: got %v want %v", i+1, status, http.StatusOK)
				}

				body := rr.Body.String()
				if !strings.Contains(body, page.expectedContent) {
					t.Errorf("Refresh %d: page content changed - missing '%s'", i+1, page.expectedContent)
				}
			}
		})
	}
}

func TestHTMXNavigation(t *testing.T) {
	// Initialize everything
	initializeTestEnvironment(t)
	
	// Create router
	r := mux.NewRouter()
	setupRoutes(r)
	
	// Test HTMX requests
	tests := []struct {
		name            string
		path            string
		htmxHeaders     map[string]string
		expectedContent string
	}{
		{
			name: "HTMX home content request",
			path: "/content/home",
			htmxHeaders: map[string]string{
				"HX-Request": "true",
				"HX-Target":  "#main-content",
			},
			expectedContent: "Technical Expertise",
		},
		{
			name: "HTMX work content request",
			path: "/content/work",
			htmxHeaders: map[string]string{
				"HX-Request": "true",
				"HX-Target":  "#main-content",
			},
			expectedContent: "Work Experience",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Add HTMX headers
			for k, v := range tt.htmxHeaders {
				req.Header.Set(k, v)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			}

			body := rr.Body.String()
			if !strings.Contains(body, tt.expectedContent) {
				t.Errorf("handler returned body without expected content '%s'", tt.expectedContent)
			}
		})
	}
}

// Helper functions

func initializeTestEnvironment(t *testing.T) {
	// Initialize configuration
	initializeConfig()
	
	// Initialize blog service
	if err := initializeBlogService(); err != nil {
		t.Fatalf("Failed to initialize blog service: %v", err)
	}
	
	// Initialize security
	initializeSecurity()
}

func setupRoutes(r *mux.Router) {
	// Routes
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/work", workHandler).Methods("GET")
	r.HandleFunc("/about", aboutHandler).Methods("GET")
	r.HandleFunc("/blog", blogHandler).Methods("GET")
	
	// HTMX content-only routes
	r.HandleFunc("/content/home", homeContentHandler).Methods("GET")
	r.HandleFunc("/content/work", workContentHandler).Methods("GET")
	r.HandleFunc("/content/about", aboutContentHandler).Methods("GET")
	r.HandleFunc("/content/blog", blogContentHandler).Methods("GET")
	
	// Other routes
	r.HandleFunc("/contact", contactHandler).Methods("POST")
	r.HandleFunc("/health", healthHandler).Methods("GET")
	
	// Blog routes
	if siteConfig.BlogEnabled {
		r.HandleFunc("/blog/{slug}", blogPostHandler).Methods("GET")
	}
	
	// Calendar routes
	if siteConfig.CalendarEnabled {
		r.HandleFunc("/calendar", calendarHandler).Methods("GET")
		r.HandleFunc("/content/calendar", calendarContentHandler).Methods("GET")
		r.HandleFunc("/api/slots", slotsHandler).Methods("GET")
		r.HandleFunc("/api/book", bookingHandler).Methods("POST")
	}
	
	// Apply middleware
	r.Use(security.SecurityMiddleware(securityConfig))
	r.Use(loggingMiddleware)
}