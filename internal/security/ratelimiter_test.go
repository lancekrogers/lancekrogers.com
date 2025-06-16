package security

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	config := &RateLimiterConfig{
		MaxRequests:   10,
		Window:        time.Minute,
		CleanupPeriod: time.Minute,
		BlockDuration: 5 * time.Minute,
	}

	rl := NewRateLimiter(config)
	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}

	if rl.maxReqs != 10 {
		t.Errorf("Expected maxReqs 10, got %d", rl.maxReqs)
	}

	if rl.window != time.Minute {
		t.Errorf("Expected window 1 minute, got %v", rl.window)
	}
}

func TestNewRateLimiterWithNilConfig(t *testing.T) {
	rl := NewRateLimiter(nil)
	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}

	// Should use default config
	defaultConfig := DefaultRateLimiterConfig()
	if rl.maxReqs != defaultConfig.MaxRequests {
		t.Errorf("Expected default maxReqs %d, got %d", defaultConfig.MaxRequests, rl.maxReqs)
	}
}

func TestRateLimiterIsAllowed(t *testing.T) {
	config := &RateLimiterConfig{
		MaxRequests:   3,
		Window:        time.Minute,
		CleanupPeriod: time.Minute,
		BlockDuration: 5 * time.Minute,
	}

	rl := NewRateLimiter(config)
	clientIP := "192.168.1.100"

	// First few requests should be allowed
	for i := 0; i < 3; i++ {
		if !rl.IsAllowed(clientIP) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Next request should be blocked (exceeds limit)
	if rl.IsAllowed(clientIP) {
		t.Error("Request should be blocked after exceeding limit")
	}

	// Should still be blocked
	if rl.IsAllowed(clientIP) {
		t.Error("Request should still be blocked")
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	config := &RateLimiterConfig{
		MaxRequests:   2,
		Window:        time.Minute,
		CleanupPeriod: time.Minute,
		BlockDuration: 5 * time.Minute,
	}

	rl := NewRateLimiter(config)

	// Different IPs should have independent limits
	ip1 := "192.168.1.100"
	ip2 := "192.168.1.101"

	// Exhaust limit for IP1
	for i := 0; i < 2; i++ {
		if !rl.IsAllowed(ip1) {
			t.Errorf("Request %d for IP1 should be allowed", i+1)
		}
	}

	// IP1 should now be limited
	if rl.IsAllowed(ip1) {
		t.Error("IP1 should be blocked")
	}

	// IP2 should still be allowed
	if !rl.IsAllowed(ip2) {
		t.Error("IP2 should be allowed")
	}
}

func TestRateLimiterGetStats(t *testing.T) {
	config := &RateLimiterConfig{
		MaxRequests:   5,
		Window:        time.Minute,
		CleanupPeriod: time.Minute,
		BlockDuration: 5 * time.Minute,
	}

	rl := NewRateLimiter(config)

	// Add some clients
	rl.IsAllowed("192.168.1.100")
	rl.IsAllowed("192.168.1.101")

	stats := rl.GetStats()

	if stats["total_clients"].(int) != 2 {
		t.Errorf("Expected 2 total clients, got %v", stats["total_clients"])
	}

	if stats["max_requests"].(int) != 5 {
		t.Errorf("Expected max_requests 5, got %v", stats["max_requests"])
	}

	if stats["blocked_clients"].(int) != 0 {
		t.Errorf("Expected 0 blocked clients, got %v", stats["blocked_clients"])
	}
}

func TestExtractClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remoteAddr string
		expected string
	}{
		{
			name:     "X-Forwarded-For single IP",
			headers:  map[string]string{"X-Forwarded-For": "203.0.113.1"},
			remoteAddr: "192.168.1.1:12345",
			expected: "203.0.113.1",
		},
		{
			name:     "X-Forwarded-For multiple IPs",
			headers:  map[string]string{"X-Forwarded-For": "203.0.113.1, 192.168.1.1"},
			remoteAddr: "10.0.0.1:12345",
			expected: "203.0.113.1",
		},
		{
			name:     "X-Real-IP header",
			headers:  map[string]string{"X-Real-IP": "203.0.113.2"},
			remoteAddr: "192.168.1.1:12345",
			expected: "203.0.113.2",
		},
		{
			name:     "RemoteAddr fallback",
			headers:  map[string]string{},
			remoteAddr: "203.0.113.3:12345",
			expected: "203.0.113.3",
		},
		{
			name:     "Invalid X-Forwarded-For falls back to RemoteAddr",
			headers:  map[string]string{"X-Forwarded-For": "invalid-ip"},
			remoteAddr: "203.0.113.4:12345",
			expected: "203.0.113.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := ExtractClientIP(req)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce1, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce failed: %v", err)
	}

	if len(nonce1) != 32 {
		t.Errorf("Expected nonce length 32, got %d", len(nonce1))
	}

	// Generate another nonce to ensure they're different
	nonce2, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce failed: %v", err)
	}

	if nonce1 == nonce2 {
		t.Error("Generated nonces should be different")
	}

	// Check that nonce contains only hex characters
	for _, char := range nonce1 {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Errorf("Nonce contains non-hex character: %c", char)
		}
	}
}