package pages

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

// service implements the pages service using file storage
type service struct {
	contentDir string
	logger     *log.Logger
}

// NewService creates a new pages service
func NewService(contentDir string, logger *log.Logger) Service {
	if contentDir == "" {
		contentDir = "content/pages"
	}
	return &service{
		contentDir: contentDir,
		logger:     logger,
	}
}

// GetPage returns a page by slug
func (s *service) GetPage(ctx context.Context, slug string) (*Page, error) {
	filename := slug + ".md"
	return s.loadPage(ctx, filename, slug)
}

// ListPages returns all pages
func (s *service) ListPages(ctx context.Context) ([]*Page, error) {
	files, err := filepath.Glob(filepath.Join(s.contentDir, "*.md"))
	if err != nil {
		return nil, fmt.Errorf("failed to list page files: %w", err)
	}

	var pages []*Page
	for _, file := range files {
		base := filepath.Base(file)
		slug := strings.TrimSuffix(base, ".md")
		page, err := s.loadPage(ctx, base, slug)
		if err != nil {
			s.logger.Printf("PAGES: Failed to load page %s: %v", base, err)
			continue
		}
		pages = append(pages, page)
	}

	return pages, nil
}

// loadPage loads and processes a page markdown file
func (s *service) loadPage(ctx context.Context, filename, slug string) (*Page, error) {
	s.logger.Printf("PAGES: Loading page from %s", filename)

	// Read the markdown file
	filePath := filepath.Join(s.contentDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		s.logger.Printf("PAGES: Error reading %s: %v", filePath, err)
		return nil, fmt.Errorf("failed to read page file %s: %w", filePath, err)
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

	// Extract frontmatter with flexible structure
	var frontMatter map[string]interface{}
	if d := frontmatter.Get(parseCtx); d != nil {
		if err := d.Decode(&frontMatter); err != nil {
			s.logger.Printf("PAGES: Warning - failed to decode frontmatter in %s: %v", filename, err)
			frontMatter = make(map[string]interface{})
		}
	} else {
		frontMatter = make(map[string]interface{})
	}

	// Convert markdown to HTML
	var buf strings.Builder
	if err := md.Renderer().Render(&buf, content, doc); err != nil {
		return nil, fmt.Errorf("failed to render markdown: %w", err)
	}

	// Extract standard fields with fallbacks
	title := ""
	if t, ok := frontMatter["title"].(string); ok {
		title = t
	}
	
	subtitle := ""
	if s, ok := frontMatter["subtitle"].(string); ok {
		subtitle = s
	}

	page := &Page{
		Slug:     slug,
		Title:    title,
		Subtitle: subtitle,
		Content:  template.HTML(buf.String()),
		Meta:     frontMatter,
		LastMod:  time.Now(), // TODO: Get actual file modification time
	}

	s.logger.Printf("PAGES: Successfully loaded page: %s", page.Title)
	return page, nil
}