package blog

import (
	"html/template"
	"time"
)

// Post represents a blog post
type Post struct {
	Slug        string        `json:"slug"`
	Title       string        `json:"title"`
	Date        time.Time     `json:"date"`
	Summary     string        `json:"summary"`
	Content     template.HTML `json:"-"`
	ReadingTime int           `json:"reading_time"`
	Tags        []string      `json:"tags"`
	FileName    string        `json:"file_name"`
}

// Frontmatter represents the YAML frontmatter of a blog post
type Frontmatter struct {
	Title       string    `yaml:"title"`
	Date        time.Time `yaml:"date"`
	Summary     string    `yaml:"summary"`
	Tags        []string  `yaml:"tags"`
	ReadingTime int       `yaml:"readingTime"`
}

// BlogConfig represents the blog configuration
type BlogConfig struct {
	Blog struct {
		Title      string       `yaml:"title"`
		Subtitle   string       `yaml:"subtitle"`
		TagFilters []TagFilter  `yaml:"tag_filters"`
		Search     SearchConfig `yaml:"search"`
	} `yaml:"blog"`
}

// TagFilter represents a tag filter button configuration
type TagFilter struct {
	Display string   `yaml:"display"`
	Tag     string   `yaml:"tag"`
	Active  bool     `yaml:"active,omitempty"`
	Aliases []string `yaml:"aliases,omitempty"`
}

// SearchConfig represents search configuration
type SearchConfig struct {
	Placeholder   string `yaml:"placeholder"`
	CaseSensitive bool   `yaml:"case_sensitive"`
}