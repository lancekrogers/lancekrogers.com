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

// Service defines the bio service interface
type Service interface {
	GetBrief(ctx context.Context) (*Bio, error)
	GetFull(ctx context.Context) (*Bio, error)
}