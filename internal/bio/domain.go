package bio

import (
	"context"
	"html/template"
	"time"
)

// Bio represents biographical content
type Bio struct {
	Title    string
	Subtitle string
	Content  template.HTML
	LastMod  time.Time
}

// AboutConfig represents the YAML configuration for the about page
type AboutConfig struct {
	Title    string `yaml:"title"`
	Subtitle string `yaml:"subtitle"`
	Hero     HeroSection `yaml:"hero"`
	WhatIDo  WhatIDoSection `yaml:"what_i_do"`
	HowIWork HowIWorkSection `yaml:"how_i_work"`
	Closing  ClosingSection `yaml:"closing"`
}

type HeroSection struct {
	Introduction string `yaml:"introduction"`
}

type WhatIDoSection struct {
	Title string `yaml:"title"`
	Items []WhatIDoItem `yaml:"items"`
}

type WhatIDoItem struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

type HowIWorkSection struct {
	Title string `yaml:"title"`
	Items []HowIWorkItem `yaml:"items"`
}

type HowIWorkItem struct {
	Number      string `yaml:"number"`
	Principle   string `yaml:"principle"`
	Description string `yaml:"description"`
}

type ClosingSection struct {
	Text string `yaml:"text"`
}

// Service defines the bio service interface
type Service interface {
	GetBrief(ctx context.Context) (*Bio, error)
	GetFull(ctx context.Context) (*Bio, error)
	GetAboutConfig(ctx context.Context) (*AboutConfig, error)
}