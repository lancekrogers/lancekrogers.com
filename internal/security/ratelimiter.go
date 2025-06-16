package security

import (
	"crypto/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config *RateLimiterConfig) *RateLimiter {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	rl := &RateLimiter{
		clients: make(map[string]*ClientInfo),
		cleanup: config.CleanupPeriod,
		maxReqs: config.MaxRequests,
		window:  config.Window,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// IsAllowed checks if a request from the given IP should be allowed
func (rl *RateLimiter) IsAllowed(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientIP]

	if !exists {
		client = &ClientInfo{
			requests: []time.Time{now},
		}
		rl.clients[clientIP] = client
		return true
	}

	// Check if client is currently blocked
	if client.blocked && now.Before(client.blockUntil) {
		return false
	}

	// Clear block if time has passed
	if client.blocked && now.After(client.blockUntil) {
		client.blocked = false
		client.requests = []time.Time{now}
		return true
	}

	// Remove old requests outside the window
	cutoff := now.Add(-rl.window)
	var recentRequests []time.Time
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			recentRequests = append(recentRequests, reqTime)
		}
	}

	// Add current request
	recentRequests = append(recentRequests, now)
	client.requests = recentRequests

	// Check if we've exceeded the limit
	if len(recentRequests) > rl.maxReqs {
		client.blocked = true
		client.blockUntil = now.Add(10 * time.Minute) // Block for 10 minutes
		return false
	}

	return true
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := map[string]interface{}{
		"total_clients": len(rl.clients),
		"max_requests":  rl.maxReqs,
		"window":        rl.window.String(),
	}

	blockedCount := 0
	for _, client := range rl.clients {
		if client.blocked && time.Now().Before(client.blockUntil) {
			blockedCount++
		}
	}
	stats["blocked_clients"] = blockedCount

	return stats
}

// cleanupLoop runs periodically to clean up stale client entries
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window * 2) // Keep some history

		for ip, client := range rl.clients {
			// Remove completely stale clients
			if len(client.requests) == 0 {
				delete(rl.clients, ip)
				continue
			}

			// Remove old requests
			var recentRequests []time.Time
			for _, reqTime := range client.requests {
				if reqTime.After(cutoff) {
					recentRequests = append(recentRequests, reqTime)
				}
			}

			if len(recentRequests) == 0 && (!client.blocked || now.After(client.blockUntil)) {
				delete(rl.clients, ip)
			} else {
				client.requests = recentRequests
			}
		}
		rl.mu.Unlock()
	}
}

// ExtractClientIP extracts the real client IP from the request
func ExtractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if idx := strings.Index(xff, ","); idx != -1 {
			ip := strings.TrimSpace(xff[:idx])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
		if net.ParseIP(xff) != nil {
			return xff
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if net.ParseIP(host) != nil {
			return host
		}
	}

	return r.RemoteAddr
}

// GenerateNonce generates a cryptographically secure random nonce
func GenerateNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	// Convert to hex string
	nonce := make([]byte, 32)
	const hexDigits = "0123456789abcdef"
	for i, b := range bytes {
		nonce[i*2] = hexDigits[b>>4]
		nonce[i*2+1] = hexDigits[b&0xf]
	}
	
	return string(nonce), nil
}