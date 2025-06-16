package bio

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

// service implements the bio service using file storage
type service struct {
	contentDir string
	logger     *log.Logger
}

// NewService creates a new bio service
func NewService(logger *log.Logger) Service {
	return &service{
		contentDir: "content",
		logger:     logger,
	}
}

// GetBrief returns the brief bio content for homepage
func (s *service) GetBrief(ctx context.Context) (*Bio, error) {
	return s.loadBio(ctx, "bio-brief.md")
}

// GetFull returns the full bio content for about page
func (s *service) GetFull(ctx context.Context) (*Bio, error) {
	return s.loadBio(ctx, "about.md")
}

// loadBio loads and processes a bio markdown file
func (s *service) loadBio(requestCtx context.Context, filename string) (*Bio, error) {
	s.logger.Printf("BIO: Loading bio from %s", filename)

	// Read the markdown file
	filePath := filepath.Join(s.contentDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		s.logger.Printf("BIO: Error reading %s: %v", filePath, err)
		return nil, fmt.Errorf("failed to read bio file %s: %w", filePath, err)
	}

	// Parse frontmatter and markdown
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&frontmatter.Extender{},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	// Create parser context and parse the document
	parseCtx := parser.NewContext()
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader, parser.WithContext(parseCtx))

	// Extract frontmatter
	var frontMatter struct {
		Title    string `yaml:"title"`
		Subtitle string `yaml:"subtitle"`
	}

	if d := frontmatter.Get(parseCtx); d != nil {
		if err := d.Decode(&frontMatter); err != nil {
			s.logger.Printf("BIO: Warning - failed to decode frontmatter in %s: %v", filename, err)
		}
	}

	// Convert markdown to HTML
	var buf strings.Builder
	if err := md.Renderer().Render(&buf, content, doc); err != nil {
		return nil, fmt.Errorf("failed to render markdown: %w", err)
	}

	bio := &Bio{
		Title:    frontMatter.Title,
		Subtitle: frontMatter.Subtitle,
		Content:  template.HTML(buf.String()),
		LastMod:  time.Now(), // TODO: Get actual file modification time from git
	}

	s.logger.Printf("BIO: Successfully loaded bio: %s", bio.Title)
	return bio, nil
}