package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestAdminSlotsHandler_MissingCredentials(t *testing.T) {
	// Ensure environment variables are not set
	os.Unsetenv("ADMIN_USERNAME")
	os.Unsetenv("ADMIN_PASSWORD")

	req := httptest.NewRequest("GET", "/admin/slots", nil)
	w := httptest.NewRecorder()

	adminSlotsHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Admin interface is not configured") {
		t.Errorf("Expected error message about missing configuration")
	}
}

func TestAdminSlotsHandler_WithCredentials(t *testing.T) {
	// Set test credentials
	os.Setenv("ADMIN_USERNAME", "testadmin")
	os.Setenv("ADMIN_PASSWORD", "testpass123")
	defer func() {
		os.Unsetenv("ADMIN_USERNAME")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	// Test without auth
	req := httptest.NewRequest("GET", "/admin/slots", nil)
	w := httptest.NewRecorder()

	adminSlotsHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for no auth, got %d", http.StatusUnauthorized, w.Code)
	}

	// Test with wrong credentials
	req = httptest.NewRequest("GET", "/admin/slots", nil)
	req.SetBasicAuth("wronguser", "wrongpass")
	w = httptest.NewRecorder()

	adminSlotsHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for wrong auth, got %d", http.StatusUnauthorized, w.Code)
	}

	// Test with correct credentials
	req = httptest.NewRequest("GET", "/admin/slots", nil)
	req.SetBasicAuth("testadmin", "testpass123")
	w = httptest.NewRecorder()

	adminSlotsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for correct auth, got %d", http.StatusOK, w.Code)
	}
}