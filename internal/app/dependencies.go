package app

import (
	"embed"

	"blockhead.consulting/internal/bio"
	"blockhead.consulting/internal/blog"
	"blockhead.consulting/internal/booking"
	"blockhead.consulting/internal/config"
	"blockhead.consulting/internal/contact"
	"blockhead.consulting/internal/email"
	"blockhead.consulting/internal/github"
	"blockhead.consulting/internal/handlers"
	"blockhead.consulting/internal/mermaid"
	"blockhead.consulting/internal/security"
	"blockhead.consulting/internal/storage/git"
	"blockhead.consulting/internal/templates"
)

// Dependencies holds all application dependencies and services
type Dependencies struct {
	// Embedded file systems
	TemplateFS embed.FS
	StaticFS   embed.FS
	BlogFS     embed.FS

	// Services
	TemplateManager   *templates.Manager
	BlogService       blog.Service
	BioService        bio.Service
	ContactService    contact.Service
	EmailService      email.Service
	GitStorageService git.Service
	BookingService    booking.Service
	GitHubService     github.Service
	MermaidRenderer   *mermaid.Renderer
	Handler           *handlers.Handler

	// Configuration
	SecurityConfig *security.Config
	SiteConfig     *config.LegacySiteConfig
	ConfigService  config.Service
	AppConfig      *config.SiteConfig
	WorkConfig     *config.WorkConfig
}

// NewDependencies creates a new Dependencies instance
func NewDependencies(templateFS, staticFS, blogFS embed.FS) *Dependencies {
	return &Dependencies{
		TemplateFS: templateFS,
		StaticFS:   staticFS,
		BlogFS:     blogFS,
	}
}