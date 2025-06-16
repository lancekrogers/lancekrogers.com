package pages

import (
	"context"
	"html/template"
	"time"
)

// Page represents a generic page with markdown content
type Page struct {
	Slug     string
	Title    string
	Subtitle string
	Content  template.HTML
	Meta     map[string]interface{} // For custom frontmatter fields
	LastMod  time.Time
}

// Service defines the pages service interface
type Service interface {
	GetPage(ctx context.Context, slug string) (*Page, error)
	ListPages(ctx context.Context) ([]*Page, error)
}