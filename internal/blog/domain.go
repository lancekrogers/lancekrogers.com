package blog

import (
	"context"
	"html/template"
	"time"
)

// Post represents a blog post
type Post struct {
	ID          int           `json:"id" db:"id"`
	Slug        string        `json:"slug" db:"slug"`
	Title       string        `json:"title" db:"title"`
	Date        time.Time     `json:"date" db:"published_at"`
	Summary     string        `json:"summary" db:"summary"`
	SummaryHTML template.HTML `json:"-" db:"-"`
	Content     template.HTML `json:"-" db:"content"`
	ReadingTime int           `json:"reading_time" db:"reading_time"`
	Tags        []string      `json:"tags" db:"-"`
	FileName    string        `json:"file_name" db:"file_path"`
	Category    string        `json:"category" db:"category"`
	WordCount   int           `json:"word_count" db:"word_count"`
	SearchRank  float64       `json:"search_rank,omitempty" db:"-"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`

	// Social media metadata
	FeaturedImage string `json:"featured_image,omitempty" db:"featured_image"`
	OGDescription string `json:"og_description,omitempty" db:"og_description"`
	TwitterHandle string `json:"twitter_handle,omitempty" db:"twitter_handle"`

	// File tracking for change detection
	FileModTime int64  `json:"-" db:"file_mod_time"`
	ContentHash string `json:"-" db:"content_hash"`
}

// Frontmatter represents the YAML frontmatter of a blog post
type Frontmatter struct {
	Title         string    `yaml:"title"`
	Date          time.Time `yaml:"date"`
	Summary       string    `yaml:"summary"`
	Tags          []string  `yaml:"tags"`
	ReadingTime   int       `yaml:"readingTime"`
	FeaturedImage string    `yaml:"featured_image"`
	OGDescription string    `yaml:"og_description"`
	TwitterHandle string    `yaml:"twitter_handle"`
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

// PaginationParams represents parameters for paginated queries
type PaginationParams struct {
	Page     int
	PerPage  int
	Category string
	Tag      string
	Search   string
}

// PaginatedPosts represents paginated blog posts result
type PaginatedPosts struct {
	Posts       []Post
	TotalPosts  int
	TotalPages  int
	CurrentPage int
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
}

// PostNavigation represents navigation between blog posts
type PostNavigation struct {
	Previous *Post
	Current  *Post
	Next     *Post
}

// RelatedPost represents a related blog post with relevance score
type RelatedPost struct {
	Post      *Post
	Relevance float64
}

// TagCount represents a tag with its usage count
type TagCount struct {
	Tag   string
	Count int
}

// PostFilter represents filters for blog posts
type PostFilter struct {
	Category string
	Tag      string
	DateFrom *time.Time
	DateTo   *time.Time
}

// PostPath represents a historical path for a blog post
type PostPath struct {
	ID          int       `json:"id" db:"id"`
	ContentHash string    `json:"content_hash" db:"content_hash"`
	Path        string    `json:"path" db:"path"`
	IsCurrent   bool      `json:"is_current" db:"is_current"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PathHistoryService defines the interface for managing post path history
type PathHistoryService interface {
	// AddPath adds a new path for a post identified by content hash
	AddPath(ctx context.Context, contentHash, path string, isCurrent bool) error

	// GetCurrentPath returns the current canonical path for a post
	GetCurrentPath(ctx context.Context, contentHash string) (string, error)

	// GetAllPaths returns all historical paths for a post
	GetAllPaths(ctx context.Context, contentHash string) ([]PostPath, error)

	// FindPostByPath returns the content hash for a post given any of its historical paths
	FindPostByPath(ctx context.Context, path string) (string, error)

	// UpdateCurrentPath marks a new path as current and old paths as historical
	UpdateCurrentPath(ctx context.Context, contentHash, newPath string) error

	// CleanupOrphanedPaths removes paths for posts that no longer exist
	CleanupOrphanedPaths(ctx context.Context) error
}
