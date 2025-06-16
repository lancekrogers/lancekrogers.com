package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Service interface for configuration management
type Service interface {
	LoadConfig(configPath string) (*SiteConfig, error)
	LoadWorkConfig(configPath string) (*WorkConfig, error)
}

// service implements the configuration service
type service struct {
	logger *log.Logger
}

// NewService creates a new configuration service
func NewService(logger *log.Logger) Service {
	return &service{
		logger: logger,
	}
}

// LoadConfig loads the site configuration from YAML file
func (s *service) LoadConfig(configPath string) (*SiteConfig, error) {
	if configPath == "" {
		configPath = "content/site.yml"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read the YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config SiteConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Set defaults
	s.setDefaults(&config)

	s.logger.Printf("Loaded configuration from %s", configPath)
	return &config, nil
}

// setDefaults sets default values for missing config
func (s *service) setDefaults(config *SiteConfig) {
	if config.Site.HeroStyle == "" {
		config.Site.HeroStyle = "professional"
	}
	
	if config.About.ProfileImage == "" {
		config.About.ProfileImage = "/static/images/profile.jpg"
	}
	
	if config.Branding.LogoMain == "" {
		config.Branding.LogoMain = "/static/logos/logo.svg"
	}
	
	if config.Branding.LogoHero == "" {
		config.Branding.LogoHero = "/static/logos/logo-hero.svg"
	}
	
	if config.Branding.PrimaryColor == "" {
		config.Branding.PrimaryColor = "#00ff88"
	}
	
	if config.Branding.SecondaryColor == "" {
		config.Branding.SecondaryColor = "#00d4ff"
	}
}

// GetBootSequence returns the boot sequence for the given hero style
func (c *SiteConfig) GetBootSequence() []string {
	switch c.Site.HeroStyle {
	case "cyberpunk":
		return c.BootSequences.Cyberpunk.Desktop
	default:
		return c.BootSequences.Professional.Desktop
	}
}

// IsCalendarEnabled returns whether calendar functionality is enabled
func (c *SiteConfig) IsCalendarEnabled() bool {
	return c.Features.CalendarEnabled
}

// IsBlogEnabled returns whether blog functionality is enabled
func (c *SiteConfig) IsBlogEnabled() bool {
	return c.Features.BlogEnabled
}

// LoadWorkConfig loads the work configuration from YAML file
func (s *service) LoadWorkConfig(configPath string) (*WorkConfig, error) {
	if configPath == "" {
		configPath = "content/work.yml"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("work config file not found: %s", configPath)
	}

	// Read the YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read work config file: %w", err)
	}

	// Parse YAML
	var config WorkConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse work config YAML: %w", err)
	}

	s.logger.Printf("Loaded work configuration from %s", configPath)
	return &config, nil
}