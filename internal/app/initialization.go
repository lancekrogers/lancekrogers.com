package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"blockhead.consulting/internal/bio"
	"blockhead.consulting/internal/blog"
	"blockhead.consulting/internal/booking"
	"blockhead.consulting/internal/config"
	"blockhead.consulting/internal/contact"
	"blockhead.consulting/internal/email"
	"blockhead.consulting/internal/events"
	"blockhead.consulting/internal/github"
	"blockhead.consulting/internal/mermaid"
	"blockhead.consulting/internal/security"
	"blockhead.consulting/internal/storage/git"
	"blockhead.consulting/internal/templates"
)

// Initialize performs all application initialization
func (d *Dependencies) Initialize() error {
	// Create directories if they don't exist
	if err := d.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Initialize template manager
	if err := d.initializeTemplateManager(); err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	// Initialize security configuration first (needed for CSP nonce)
	if err := d.initializeSecurity(); err != nil {
		return fmt.Errorf("failed to initialize security: %w", err)
	}

	// Initialize configuration (uses security config)
	if err := d.initializeConfig(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Initialize mermaid renderer
	d.initializeMermaid()

	// Load blog posts (after config is initialized)
	if err := d.initializeBlogService(); err != nil {
		return fmt.Errorf("failed to initialize blog service: %w", err)
	}

	// Initialize booking service if calendar is enabled
	if d.AppConfig != nil && d.AppConfig.Features.CalendarEnabled {
		if err := d.initializeBookingService(); err != nil {
			log.Printf("Failed to initialize booking service: %v", err)
			// Continue without booking service
		}
	}

	// Initialize GitHub service
	if err := d.initializeGitHubService(); err != nil {
		log.Printf("Failed to initialize GitHub service: %v", err)
		// Continue without GitHub service - it will use default values
	}

	return nil
}

func (d *Dependencies) createDirectories() error {
	dirs := []string{"templates", "static", "data", "content/blog"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Warning: Could not create directory %s: %v", dir, err)
		}
	}
	return nil
}

func (d *Dependencies) initializeTemplateManager() error {
	logger := log.New(os.Stdout, "[templates] ", log.LstdFlags)
	var err error
	d.TemplateManager, err = templates.NewManager(d.TemplateFS, logger)
	if err != nil {
		return fmt.Errorf("failed to create template manager: %w", err)
	}
	return nil
}

func (d *Dependencies) initializeSecurity() error {
	// Generate CSP nonce
	nonce, err := security.GenerateNonce()
	if err != nil {
		return fmt.Errorf("failed to generate CSP nonce: %w", err)
	}

	// Initialize rate limiter with default config
	rateLimiterConfig := security.DefaultRateLimiterConfig()
	rateLimiter := security.NewRateLimiter(rateLimiterConfig)

	d.SecurityConfig = &security.Config{
		CSPNonce:    nonce,
		RateLimiter: rateLimiter,
		ValidFileTypes: map[string]bool{
			".css":   true,
			".js":    true,
			".png":   true,
			".jpg":   true,
			".jpeg":  true,
			".gif":   true,
			".svg":   true,
			".ico":   true,
			".woff":  true,
			".woff2": true,
		},
		MaxUploadSize:  10 << 20,        // 10MB
		SessionTimeout: 30 * time.Minute,
		MaxRequestSize: 1 << 20, // 1MB
	}
	return nil
}

func (d *Dependencies) initializeConfig() error {
	logger := log.New(os.Stdout, "[config] ", log.LstdFlags)
	initializer := config.NewInitializer(logger)

	var err error
	d.SiteConfig, d.AppConfig, d.WorkConfig, err = initializer.InitializeConfig(d.SecurityConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Initialize config service for handlers
	d.ConfigService = config.NewService(logger)
	return nil
}

func (d *Dependencies) initializeMermaid() {
	mermaidLogger := log.New(os.Stdout, "[mermaid] ", log.LstdFlags)
	d.MermaidRenderer = mermaid.NewRenderer("data/mermaid-temp", mermaidLogger)
}

func (d *Dependencies) initializeBlogService() error {
	// Skip if blog is disabled
	if !d.SiteConfig.BlogEnabled {
		log.Printf("Blog disabled - skipping blog service initialization")
		return nil
	}

	// Create logger
	logger := log.New(os.Stdout, "[blog] ", log.LstdFlags)

	// Create event bus
	eventBus := events.NewInMemoryEventBus(5, logger)

	// Check if SQLite blog service is enabled
	useSQLite := config.GetEnvBool("BLOG_SQLITE_ENABLED", false)
	dbPath := config.GetEnv("BLOG_DB_PATH", "./data/blog.db")

	if useSQLite {
		log.Printf("BLOG: Initializing SQLite blog service...")

		// Ensure database directory exists
		os.MkdirAll(filepath.Dir(dbPath), 0755)

		// Create SQLite blog service
		sqliteService, err := blog.NewSQLiteService(dbPath, d.BlogFS, "content/blog", logger, eventBus)
		if err != nil {
			log.Printf("BLOG: Failed to initialize SQLite service, falling back to file-based: %v", err)
			// Fallback to file-based service
			d.BlogService = blog.NewService(d.BlogFS, logger, eventBus)
		} else {
			d.BlogService = sqliteService
			log.Printf("BLOG: SQLite blog service initialized successfully")
		}
	} else {
		// Create standard file-based blog service
		d.BlogService = blog.NewService(d.BlogFS, logger, eventBus)
		log.Printf("BLOG: File-based blog service initialized")
	}

	// Initialize email service
	emailConfig := &email.EmailConfig{
		SMTPHost:    config.GetEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:    config.GetEnvInt("SMTP_PORT", 587),
		Username:    config.GetEnv("SMTP_USERNAME", ""),
		Password:    config.GetEnv("SMTP_PASSWORD", ""),
		FromAddress: config.GetEnv("SMTP_FROM_ADDRESS", ""),
		FromName:    config.GetEnv("SMTP_FROM_NAME", "Blockhead Consulting"),
		TLSEnabled:  config.GetEnvBool("SMTP_TLS_ENABLED", true),
	}
	d.EmailService = email.NewService(emailConfig)

	// Log email configuration status
	if emailConfig.SMTPHost != "" && emailConfig.Username != "" && emailConfig.Password != "" {
		log.Printf("EMAIL: Service initialized with SMTP host: %s", emailConfig.SMTPHost)
	} else {
		log.Printf("EMAIL: Service initialized but missing configuration")
		if emailConfig.SMTPHost == "" {
			log.Printf("EMAIL: Missing SMTP_HOST")
		}
		if emailConfig.Username == "" {
			log.Printf("EMAIL: Missing SMTP_USERNAME")
		}
		if emailConfig.Password == "" {
			log.Printf("EMAIL: Missing SMTP_PASSWORD")
		}
	}

	// Initialize Git storage service
	gitConfig := git.StorageConfig{
		RepoPath:      config.GetEnv("GIT_REPO_PATH", "./data/messages"),
		EncryptionKey: config.GetEnv("GIT_ENCRYPTION_KEY", ""),
		RemoteURL:     config.GetEnv("GIT_REMOTE_URL", ""),
		PushOnWrite:   config.GetEnvBool("GIT_PUSH_ON_WRITE", false),
		Branch:        config.GetEnv("GIT_BRANCH", "main"),
		CommitAuthor:  config.GetEnv("GIT_COMMIT_AUTHOR", "Blockhead Consulting Bot"),
		CommitEmail:   config.GetEnv("GIT_COMMIT_EMAIL", "bot@blockhead.consulting"),
	}

	var err error
	d.GitStorageService, err = git.NewService(gitConfig, logger, eventBus)
	if err != nil {
		log.Printf("Warning: Git storage service initialization failed: %v", err)
		// Continue without git storage for development
	}

	// Initialize contact service
	contactLogger := log.New(os.Stdout, "[contact] ", log.LstdFlags)
	adminEmail := config.GetEnv("ADMIN_EMAIL", "")
	if adminEmail != "" {
		log.Printf("CONTACT: Admin email configured: %s", adminEmail)
	} else {
		log.Printf("CONTACT: WARNING - No ADMIN_EMAIL configured, contact form emails will not be sent!")
	}
	d.ContactService = contact.NewService(d.GitStorageService, eventBus, d.EmailService, adminEmail, contactLogger)

	// Initialize bio service
	bioLogger := log.New(os.Stdout, "[bio] ", log.LstdFlags)
	d.BioService = bio.NewService(bioLogger)

	// Start services
	ctx := context.Background()
	if err := eventBus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	// Start the blog service (cast to registry.Service interface)
	if starter, ok := d.BlogService.(interface{ Start(context.Context) error }); ok {
		if err := starter.Start(ctx); err != nil {
			return fmt.Errorf("failed to start blog service: %w", err)
		}
	}

	log.Printf("Blog service initialized successfully")
	return nil
}

func (d *Dependencies) initializeBookingService() error {
	bookingsFile := "data/bookings.json"
	bookingLogger := log.New(os.Stdout, "[booking] ", log.LstdFlags)

	var err error
	d.BookingService, err = booking.NewFileBasedService(bookingsFile, bookingLogger)
	if err != nil {
		return fmt.Errorf("failed to create booking service: %w", err)
	}

	log.Printf("Booking service initialized successfully")
	return nil
}

func (d *Dependencies) initializeGitHubService() error {
	// Create GitHub configuration
	githubConfig := &github.GitHubConfig{
		Username:    config.GetEnv("GITHUB_USERNAME", ""),
		Token:       config.GetEnv("GITHUB_TOKEN", ""),
		CacheExpiry: 1 * time.Hour,
		Enabled:     config.GetEnvBool("GITHUB_ENABLED", false),
	}

	// If both username and token are provided, enable GitHub integration
	if githubConfig.Username != "" && githubConfig.Token != "" {
		githubConfig.Enabled = true
		log.Printf("✅ GITHUB: Service enabled for user '%s' - will fetch live GitHub stats", githubConfig.Username)
	} else {
		log.Printf("⚠️  GITHUB: Service disabled - showing fallback stats")
		log.Printf("   To enable live GitHub data:")
		log.Printf("   1. Set GITHUB_USERNAME=lancekrogers")
		log.Printf("   2. Set GITHUB_TOKEN=<your_github_token>")
		log.Printf("   3. Generate token at: https://github.com/settings/tokens")
		log.Printf("   4. Token needs 'read:user' and 'repo' scopes")
		githubConfig.Enabled = false
	}

	// Initialize GitHub service
	d.GitHubService = github.NewService(githubConfig)

	return nil
}