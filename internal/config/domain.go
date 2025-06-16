package config

// SiteConfig represents the complete site configuration
type SiteConfig struct {
	Site           SiteInfo            `yaml:"site"`
	About          AboutInfo           `yaml:"about"`
	Contact        ContactInfo         `yaml:"contact"`
	Branding       BrandingInfo        `yaml:"branding"`
	Services       ServicesInfo        `yaml:"services"`
	Packages       PackagesInfo        `yaml:"packages"`
	Features       FeaturesInfo        `yaml:"features"`
	Stats          []StatInfo          `yaml:"stats"`
	Expertise      ExpertiseInfo       `yaml:"expertise"`
	BootSequences  BootSequencesInfo   `yaml:"boot_sequences"`
	WorkExperience WorkExperienceInfo  `yaml:"work_experience"`
}

type SiteInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Tagline     string `yaml:"tagline"`
	Subtitle    string `yaml:"subtitle"`
	HeroStyle   string `yaml:"hero_style"`
}

type AboutInfo struct {
	Title        string `yaml:"title"`
	Subtitle     string `yaml:"subtitle"`
	ProfileImage string `yaml:"profile_image"`
}

type ContactInfo struct {
	Email string `yaml:"email"`
	Phone string `yaml:"phone"`
}

type BrandingInfo struct {
	LogoMain       string `yaml:"logo_main"`
	LogoHero       string `yaml:"logo_hero"`
	PrimaryColor   string `yaml:"primary_color"`
	SecondaryColor string `yaml:"secondary_color"`
}

type ServiceInfo struct {
	Title string   `yaml:"title"`
	Icon  string   `yaml:"icon"`
	Rate  string   `yaml:"rate"`
	Tags  []string `yaml:"tags"`
	Items []string `yaml:"items"`
}

type ServicesInfo struct {
	DisplayRates bool        `yaml:"display_rates"`
	Crypto       ServiceInfo `yaml:"crypto"`
	AI           ServiceInfo `yaml:"ai"`
}

type PackageInfo struct {
	Title         string   `yaml:"title"`
	Price         string   `yaml:"price"`
	Duration      string   `yaml:"duration"`
	Description   string   `yaml:"description"`
	Featured      bool     `yaml:"featured,omitempty"`
	DetailedInfo  *PackageDetails `yaml:"detailed_info,omitempty"`
}

type PackageDetails struct {
	WhatYouGet []string `yaml:"what_you_get"`
	Process    []string `yaml:"process"`
	Outcomes   []string `yaml:"outcomes"`
}

type PackagesInfo struct {
	DisplayPrices   bool        `yaml:"display_prices"`
	Evaluation      PackageInfo `yaml:"evaluation"`
	Assessment      PackageInfo `yaml:"assessment"`
	Pilot           PackageInfo `yaml:"pilot"`
	Accelerator     PackageInfo `yaml:"accelerator"`
	AIAcceleration  PackageInfo `yaml:"ai_acceleration"`
}

type FeaturesInfo struct {
	CalendarEnabled  bool `yaml:"calendar_enabled"`
	BlogEnabled      bool `yaml:"blog_enabled"`
	AnalyticsEnabled bool `yaml:"analytics_enabled"`
}

type StatInfo struct {
	Value string `yaml:"value"`
	Label string `yaml:"label"`
}

type BootInfo struct {
	Professional []string `yaml:"professional"`
	Cyberpunk    []string `yaml:"cyberpunk"`
}

type ExpertiseItem struct {
	Title        string             `yaml:"title"`
	Items        string             `yaml:"items"`
	DetailedInfo *ExpertiseDetails  `yaml:"detailed_info,omitempty"`
}

type ExpertiseDetails struct {
	Overview      string   `yaml:"overview"`
	Experience    []string `yaml:"experience"`
	Technologies  []string `yaml:"technologies"`
	Achievements  []string `yaml:"achievements"`
}

type ExpertiseInfo struct {
	Languages  ExpertiseItem `yaml:"languages"`
	Blockchain ExpertiseItem `yaml:"blockchain"`
	AIML       ExpertiseItem `yaml:"ai_ml"`
}

type BootSequenceStyle struct {
	Desktop []string `yaml:"desktop"`
	Mobile  []string `yaml:"mobile"`
}

type BootSequencesInfo struct {
	Professional BootSequenceStyle `yaml:"professional"`
	Cyberpunk    BootSequenceStyle `yaml:"cyberpunk"`
}

// Legacy structs for backwards compatibility
type Company struct {
	Name         string   `yaml:"name"`
	Role         string   `yaml:"role"`
	Duration     string   `yaml:"duration"`
	Description  string   `yaml:"description"`
	Achievements []string `yaml:"achievements"`
	Technologies []string `yaml:"technologies"`
	Link         string   `yaml:"link,omitempty"`
}

type WorkCategory struct {
	Title     string    `yaml:"title"`
	Icon      string    `yaml:"icon"`
	Companies []Company `yaml:"companies"`
}

type WorkExperienceInfo struct {
	Intro    string       `yaml:"intro"`
	FinTech  WorkCategory `yaml:"fintech"`
	Crypto   WorkCategory `yaml:"crypto"`
	AI       WorkCategory `yaml:"ai"`
}

// New work page structures
type WorkConfig struct {
	Intro      string            `yaml:"intro"`
	FinTech    WorkSection       `yaml:"fintech"`
	Blockchain WorkSection       `yaml:"blockchain"`
	AI         WorkSection       `yaml:"ai"`
}

type WorkSection struct {
	Title       string        `yaml:"title"`
	Icon        string        `yaml:"icon"`
	Description string        `yaml:"description"`
	Companies   []WorkCompany `yaml:"companies,omitempty"`
	Projects    []WorkProject `yaml:"projects,omitempty"`
}

type WorkCompany struct {
	Name         string              `yaml:"name"`
	Role         string              `yaml:"role"`
	Duration     string              `yaml:"duration"`
	Summary      string              `yaml:"summary"`
	Featured     bool                `yaml:"featured,omitempty"`
	DetailedInfo *WorkCompanyDetails `yaml:"detailed_info,omitempty"`
}

type WorkProject struct {
	Name         string              `yaml:"name"`
	Role         string              `yaml:"role"`
	Duration     string              `yaml:"duration"`
	Summary      string              `yaml:"summary"`
	Featured     bool                `yaml:"featured,omitempty"`
	DetailedInfo *WorkProjectDetails `yaml:"detailed_info,omitempty"`
	Link         string              `yaml:"link,omitempty"`
}

type WorkCompanyDetails struct {
	Description  string           `yaml:"description"`
	KeyProjects  []ProjectDetail  `yaml:"key_projects"`
	Achievements []string         `yaml:"achievements"`
}

type WorkProjectDetails struct {
	Description   string          `yaml:"description"`
	KeyFeatures   []ProjectDetail `yaml:"key_features,omitempty"`
	KeyContributions []ProjectDetail `yaml:"key_contributions,omitempty"`
	Achievements  []string        `yaml:"achievements"`
	Link          string          `yaml:"link,omitempty"`
}

type ProjectDetail struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Impact       string   `yaml:"impact"`
	Technologies []string `yaml:"technologies"`
}