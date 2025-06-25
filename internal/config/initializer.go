package config

import (
	"log"
	"os"
	"strconv"
	
	"blockhead.consulting/internal/security"
	"github.com/joho/godotenv"
)

// Initializer handles configuration initialization
type Initializer struct {
	configService Service
	logger        *log.Logger
}

// NewInitializer creates a new configuration initializer
func NewInitializer(logger *log.Logger) *Initializer {
	return &Initializer{
		configService: NewService(logger),
		logger:        logger,
	}
}

// InitializeConfig loads configuration from YAML or environment variables
func (i *Initializer) InitializeConfig(securityConfig *security.Config) (*LegacySiteConfig, *SiteConfig, *WorkConfig, error) {
	// Load .env file if it exists (ignore errors - file may not exist)
	godotenv.Load()
	
	// Load configuration from YAML file
	appConfig, err := i.configService.LoadConfig("content/site.yml")
	if err != nil {
		i.logger.Printf("Warning: Failed to load site.yml, using environment variables: %v", err)
		// Fallback to environment-based configuration
		siteConfig, appConfig := i.initializeLegacyConfig(securityConfig)
		return siteConfig, appConfig, nil, nil
	}
	
	// Create legacy SiteConfig for backward compatibility
	environment := getEnv("ENVIRONMENT", "development")
	siteConfig := &LegacySiteConfig{
		CalendarEnabled: getEnvBool("CALENDAR_ENABLED", appConfig.Features.CalendarEnabled),
		BlogEnabled:     getEnvBool("BLOG_ENABLED", appConfig.Features.BlogEnabled),
		SiteName:        appConfig.Site.Name,
		Environment:     environment,
		HeroStyle:       getEnv("HERO_STYLE", "professional"),
		ConsoleLogging:  getEnvBool("CONSOLE_LOGGING", environment == "development"),
		CSPNonce:        securityConfig.CSPNonce,
	}
	
	// Validate hero style
	if siteConfig.HeroStyle != "professional" && siteConfig.HeroStyle != "cyberpunk" {
		i.logger.Printf("Warning: Invalid HERO_STYLE '%s', defaulting to 'professional'", siteConfig.HeroStyle)
		siteConfig.HeroStyle = "professional"
	}
	
	i.logger.Printf("CONFIG: Loaded from site.yml - %s", appConfig.Site.Name)
	i.logger.Printf("CONFIG: Calendar enabled: %v", siteConfig.CalendarEnabled)
	i.logger.Printf("CONFIG: Blog enabled: %v", siteConfig.BlogEnabled)
	i.logger.Printf("CONFIG: Environment: %s", siteConfig.Environment)
	i.logger.Printf("CONFIG: Hero style: %s", siteConfig.HeroStyle)
	
	// Load work configuration
	workConfig, err := i.configService.LoadWorkConfig("content/work.yml")
	if err != nil {
		i.logger.Printf("Warning: Failed to load work.yml: %v", err)
		workConfig = nil
	} else {
		i.logger.Printf("CONFIG: Loaded work configuration with %d sections", 3)
	}
	
	return siteConfig, appConfig, workConfig, nil
}

// initializeLegacyConfig creates configuration from environment variables
func (i *Initializer) initializeLegacyConfig(securityConfig *security.Config) (*LegacySiteConfig, *SiteConfig) {
	calendarEnabled := true // Default to enabled
	
	// Check environment variable
	if envCalendar := os.Getenv("CALENDAR_ENABLED"); envCalendar != "" {
		if parsed, err := strconv.ParseBool(envCalendar); err == nil {
			calendarEnabled = parsed
		}
	}
	
	// Blog enabled (default true, can be disabled for development)
	blogEnabled := true
	if blogEnv := os.Getenv("BLOG_ENABLED"); blogEnv != "" {
		if parsed, err := strconv.ParseBool(blogEnv); err == nil {
			blogEnabled = parsed
		}
	}
	
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}
	
	siteName := os.Getenv("SITE_NAME")
	if siteName == "" {
		siteName = "Blockhead Consulting"
	}
	
	heroStyle := os.Getenv("HERO_STYLE")
	if heroStyle == "" {
		heroStyle = "professional" // Default to professional mode
	}
	
	siteConfig := &LegacySiteConfig{
		CalendarEnabled: calendarEnabled,
		BlogEnabled:     blogEnabled,
		SiteName:        siteName,
		Environment:     environment,
		HeroStyle:       heroStyle,
		ConsoleLogging:  getEnvBool("CONSOLE_LOGGING", environment == "development"),
		CSPNonce:        securityConfig.CSPNonce,
	}
	
	// Create minimal appConfig for compatibility
	appConfig := &SiteConfig{
		Site: SiteInfo{
			Name:      siteName,
			HeroStyle: heroStyle,
		},
		Features: FeaturesInfo{
			CalendarEnabled: calendarEnabled,
			BlogEnabled:     blogEnabled,
		},
	}
	
	i.logger.Printf("CONFIG: Using legacy environment variables")
	i.logger.Printf("CONFIG: Calendar enabled: %v", siteConfig.CalendarEnabled)
	i.logger.Printf("CONFIG: Blog enabled: %v", siteConfig.BlogEnabled)
	i.logger.Printf("CONFIG: Environment: %s", siteConfig.Environment)
	i.logger.Printf("CONFIG: Hero style: %s", siteConfig.HeroStyle)
	i.logger.Printf("CONFIG: Console logging: %v", siteConfig.ConsoleLogging)
	
	return siteConfig, appConfig
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetEnv exports the helper function for use in other packages
func GetEnv(key, defaultValue string) string {
	return getEnv(key, defaultValue)
}

// GetEnvInt exports the helper function for use in other packages
func GetEnvInt(key string, defaultValue int) int {
	return getEnvInt(key, defaultValue)
}

// GetEnvBool exports the helper function for use in other packages
func GetEnvBool(key string, defaultValue bool) bool {
	return getEnvBool(key, defaultValue)
}