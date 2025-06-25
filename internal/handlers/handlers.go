package handlers

import (
	"html/template"
	"log"

	"blockhead.consulting/internal/bio"
	"blockhead.consulting/internal/blog"
	"blockhead.consulting/internal/config"
	"blockhead.consulting/internal/contact"
	"blockhead.consulting/internal/email"
	"blockhead.consulting/internal/github"
	"blockhead.consulting/internal/security"
	"blockhead.consulting/internal/storage/git"
	"blockhead.consulting/internal/booking"
	"blockhead.consulting/internal/mermaid"
)

// Handler holds shared dependencies for all HTTP handlers
type Handler struct {
	Templates         *template.Template
	BlogService       blog.Service
	BioService        bio.Service
	ContactService    contact.Service
	EmailService      email.Service
	GitStorageService git.Service
	BookingService    booking.Service
	GitHubService     github.Service
	MermaidRenderer   *mermaid.Renderer
	SecurityConfig    *security.Config
	SiteConfig        *SiteConfig
	ConfigService     config.Service
	AppConfig         *config.SiteConfig
	WorkConfig        *config.WorkConfig
	Logger            *log.Logger
}

// SiteConfig holds site-specific configuration
type SiteConfig struct {
	CalendarEnabled bool
	BlogEnabled     bool
	SiteName        string
	Environment     string
	HeroStyle       string // "professional" or "cyberpunk"
	ConsoleLogging  bool   // Enable/disable JavaScript console logging
	CSPNonce        string // Content Security Policy nonce for inline scripts
}

// New creates a new Handler with all dependencies
func New(
	templates *template.Template,
	blogService blog.Service,
	bioService bio.Service,
	contactService contact.Service,
	emailService email.Service,
	gitStorageService git.Service,
	securityConfig *security.Config,
	siteConfig *SiteConfig,
	configService config.Service,
	appConfig *config.SiteConfig,
	workConfig *config.WorkConfig,
	bookingService booking.Service,
	githubService github.Service,
	mermaidRenderer *mermaid.Renderer,
	logger *log.Logger,
) *Handler {
	return &Handler{
		Templates:         templates,
		BlogService:       blogService,
		BioService:        bioService,
		ContactService:    contactService,
		EmailService:      emailService,
		GitStorageService: gitStorageService,
		BookingService:    bookingService,
		GitHubService:     githubService,
		MermaidRenderer:   mermaidRenderer,
		SecurityConfig:    securityConfig,
		SiteConfig:        siteConfig,
		ConfigService:     configService,
		AppConfig:         appConfig,
		WorkConfig:        workConfig,
		Logger:            logger,
	}
}