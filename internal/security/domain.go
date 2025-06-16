package security

import (
	"regexp"
	"sync"
	"time"
)

// Common validation patterns
var (
	EmailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	NameRegex     = regexp.MustCompile(`^[a-zA-Z\s'-]{1,100}$`)
	CompanyRegex  = regexp.MustCompile(`^[a-zA-Z0-9\s&.,-]{1,200}$`)
	MessageRegex  = regexp.MustCompile(`^[\w\s.,!?'"()\-_@#$%&*+=/:;<>{}[\]\\|~\n\r]+$`)
	SlotIDRegex   = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{2}:[0-9]{2}$`)
)

// Config holds security configuration
type Config struct {
	CSPNonce         string
	RateLimiter      *RateLimiter
	ValidFileTypes   map[string]bool
	MaxUploadSize    int64
	SessionTimeout   time.Duration
	MaxRequestSize   int64
}

// RateLimiter implements IP-based rate limiting
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*ClientInfo
	cleanup  time.Duration
	maxReqs  int
	window   time.Duration
}

// ClientInfo tracks request information for a specific client
type ClientInfo struct {
	requests   []time.Time
	blocked    bool
	blockUntil time.Time
}

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	MaxRequests    int           // Maximum requests per window
	Window         time.Duration // Time window for rate limiting
	CleanupPeriod  time.Duration // How often to clean up stale clients
	BlockDuration  time.Duration // How long to block clients that exceed limits
}

// SecurityHeaders holds configuration for HTTP security headers
type SecurityHeaders struct {
	CSPNonce                string
	EnableHSTS              bool
	HSTSMaxAge              int
	EnableXSSProtection     bool
	EnableContentTypeNoSniff bool
	EnableFrameOptions      bool
	FrameOptions            string // DENY, SAMEORIGIN, or ALLOW-FROM uri
	PermissionsPolicy       string // Custom permissions policy, empty to use default
}

// DefaultSecurityHeaders returns a secure default configuration
func DefaultSecurityHeaders() *SecurityHeaders {
	return &SecurityHeaders{
		EnableHSTS:              true,
		HSTSMaxAge:              31536000, // 1 year
		EnableXSSProtection:     true,
		EnableContentTypeNoSniff: true,
		EnableFrameOptions:      true,
		FrameOptions:            "DENY",
		// Default permissions policy blocks sensitive features most sites don't need
		PermissionsPolicy:       "geolocation=(), microphone=(), camera=(), payment=(), usb=(), fullscreen=(self)",
	}
}

// DefaultRateLimiterConfig returns a reasonable default configuration
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		MaxRequests:   500,                // 500 requests per minute
		Window:        time.Minute,        // 1 minute window
		CleanupPeriod: 5 * time.Minute,    // Clean up every 5 minutes
		BlockDuration: 10 * time.Minute,   // Block for 10 minutes
	}
}

// ConsultingWebsiteHeaders returns security headers appropriate for consulting/business websites
func ConsultingWebsiteHeaders() *SecurityHeaders {
	return &SecurityHeaders{
		EnableHSTS:              true,
		HSTSMaxAge:              31536000, // 1 year
		EnableXSSProtection:     true,
		EnableContentTypeNoSniff: true,
		EnableFrameOptions:      true,
		FrameOptions:            "DENY",
		// Business-friendly policy: allows payments and fullscreen, blocks sensitive location/hardware access
		PermissionsPolicy:       "geolocation=(), usb=(), fullscreen=(self), payment=(self)",
	}
}