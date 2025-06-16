package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNavigationPages(t *testing.T) {
	// Test cases for all pages
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
				"Work Experience",
			},
		},
		{
			name: "Work page shows correct content",
			path: "/work",
			expectedContent: []string{
				"Work Experience",
				"Bank of America",
				"Mythical Games",
				"Senior Backend Engineer with 9",
			},
			notExpected: []string{
				"BLOCKHEAD CONSULTING",
				"Technical Expertise",
			},
		},
		{
			name: "About page shows correct content",
			path: "/about",
			expectedContent: []string{
				"About Lance Rogers",
				"Strategic Systems Architect",
			},
			notExpected: []string{
				"Work Experience",
				"BLOCKHEAD CONSULTING",
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
		{
			name: "Work content HTMX endpoint",
			path: "/content/work",
			expectedContent: []string{
				"Work Experience",
				"Bank of America",
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
			
			// Route to the appropriate handler
			switch tt.path {
			case "/":
				homeHandler(rr, req)
			case "/work":
				workHandler(rr, req)
			case "/about":
				aboutHandler(rr, req)
			case "/content/home":
				homeContentHandler(rr, req)
			case "/content/work":
				workContentHandler(rr, req)
			default:
				t.Fatalf("Unknown path: %s", tt.path)
			}

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
	// Test that refreshing pages doesn't change content
	pages := []struct {
		path    string
		handler func(http.ResponseWriter, *http.Request)
		content string
	}{
		{"/", homeHandler, "BLOCKHEAD CONSULTING"},
		{"/work", workHandler, "Work Experience"},
		{"/about", aboutHandler, "About Lance Rogers"},
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
				page.handler(rr, req)

				if status := rr.Code; status != http.StatusOK {
					t.Errorf("Refresh %d: got status %v want %v", i+1, status, http.StatusOK)
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
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	homeHandler(rr, req)

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