package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNavigationPages(t *testing.T) {
	_, router := createTestApp(t)
	
	// Test cases for existing pages only (work and services content is now in about)
	tests := []struct {
		name            string
		path            string
		expectedContent []string
		notExpected     []string
	}{
		{
			name: "Home page shows correct content",
			path: "/",
			expectedContent: []string{
				"BLOCKHEAD CONSULTING",
				"Technical Expertise",
				"Core Languages",
				"Blockchain",
				"AI Engineering",
			},
			notExpected: []string{
				"Work Experience", // Work content is in about page now
			},
		},
		{
			name: "About page shows correct content",
			path: "/about",
			expectedContent: []string{
				"About Lance Rogers",
				"Fractional CTO",
			},
			notExpected: []string{
				"BLOCKHEAD CONSULTING", // Hero content only on home
			},
		},
		{
			name: "Home content HTMX endpoint",
			path: "/content/home",
			expectedContent: []string{
				"Technical Expertise",
				"Core Languages",
			},
			notExpected: []string{
				"<!doctype html>", // Should not include full HTML
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			}

			body := rr.Body.String()

			// Check expected content
			for _, expected := range tt.expectedContent {
				if !strings.Contains(body, expected) {
					t.Errorf("Page %s missing expected content '%s'", tt.path, expected)
				}
			}

			// Check content that should NOT be present
			for _, notExpected := range tt.notExpected {
				if strings.Contains(body, notExpected) {
					t.Errorf("Page %s contains unexpected content '%s'", tt.path, notExpected)
				}
			}
		})
	}
}

func TestPageRefreshes(t *testing.T) {
	_, router := createTestApp(t)
	
	// Test that refreshing pages doesn't change content (only existing pages)
	pages := []struct {
		path           string
		content        string
		expectedStatus int
	}{
		{"/", "BLOCKHEAD CONSULTING", http.StatusOK},
		{"/about", "About Lance Rogers", http.StatusOK},
		{"/blog", "Blog", http.StatusOK},
	}

	for _, page := range pages {
		t.Run("Refresh "+page.path, func(t *testing.T) {
			// Test 3 consecutive refreshes
			for i := 0; i < 3; i++ {
				req, err := http.NewRequest("GET", page.path, nil)
				if err != nil {
					t.Fatal(err)
				}

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				if status := rr.Code; status != page.expectedStatus {
					t.Errorf("Refresh %d: got status %v want %v", i+1, status, page.expectedStatus)
				}

				body := rr.Body.String()
				if !strings.Contains(body, page.content) {
					t.Errorf("Refresh %d: missing content '%s'", i+1, page.content)
				}
			}
		})
	}
}

func TestTechnicalExpertiseUpdated(t *testing.T) {
	_, router := createTestApp(t)
	
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	body := rr.Body.String()

	// Check that expertise section has been updated
	expectedExpertise := []string{
		"Core Languages",
		"Go • Python • Solidity",
		"Blockchain", 
		"Ethereum • Polygon • DeFi • Smart Contracts",
		"AI Engineering",
		"OpenAI • Claude • Agent Frameworks • RAG",
	}

	for _, expected := range expectedExpertise {
		if !strings.Contains(body, expected) {
			t.Errorf("Missing updated expertise content: %s", expected)
		}
	}

	// Check that Infrastructure section is not in expertise (it's OK in service names)
	// Look for the specific pattern in expertise section
	if strings.Contains(body, `<h4>Infrastructure</h4>`) {
		t.Error("Infrastructure section should be removed from expertise")
	}
	if strings.Contains(body, "K8s • Docker") {
		t.Error("Infrastructure items should be removed from expertise")
	}
}