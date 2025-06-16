package blog

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"math"
	"path/filepath"
	"sort"
	"strings"

	"blockhead.consulting/internal/errors"
	"blockhead.consulting/internal/events"
	"blockhead.consulting/internal/registry"
	
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v3"
)

// Service provides blog functionality
type Service interface {
	// GetAll returns all published blog posts
	GetAll(ctx context.Context) []Post
	
	// GetBySlug returns a blog post by its slug
	GetBySlug(ctx context.Context, slug string) (*Post, error)
	
	// Search searches blog posts by query
	Search(ctx context.Context, query string) []Post
	
	// GetByTag returns posts with a specific tag
	GetByTag(ctx context.Context, tag string) []Post
	
	// GetTags returns all unique tags
	GetTags(ctx context.Context) []string
	
	// LoadPosts loads posts from the filesystem
	LoadPosts(ctx context.Context) error
	
	// GetBlogConfig returns the blog configuration
	GetBlogConfig() *BlogConfig
}

// service implements the blog service
type service struct {
	posts      []Post
	postMap    map[string]*Post
	tagIndex   map[string][]int // tag -> post indices
	blogFS     fs.FS
	blogDir    string // directory containing blog posts
	logger     *log.Logger
	eventBus   events.EventBus
	blogConfig *BlogConfig
}

// NewService creates a new blog service
func NewService(blogFS fs.FS, logger *log.Logger, eventBus events.EventBus) *service {
	return NewServiceWithOptions(blogFS, "content/blog", logger, eventBus)
}

// NewServiceWithOptions creates a new blog service with custom options
func NewServiceWithOptions(blogFS fs.FS, blogDir string, logger *log.Logger, eventBus events.EventBus) *service {
	if logger == nil {
		logger = log.Default()
	}
	
	s := &service{
		postMap:  make(map[string]*Post),
		tagIndex: make(map[string][]int),
		blogFS:   blogFS,
		blogDir:  blogDir,
		logger:   logger,
		eventBus: eventBus,
	}
	
	// Load blog configuration
	s.loadBlogConfig()
	
	return s
}

// Implement registry.Service interface
func (s *service) Name() string {
	return "blog"
}

func (s *service) Start(ctx context.Context) error {
	s.logger.Printf("BLOG: Starting blog service...")
	
	// Load posts on startup
	if err := s.LoadPosts(ctx); err != nil {
		return fmt.Errorf("failed to load blog posts: %w", err)
	}
	
	s.logger.Printf("BLOG: Blog service started with %d posts", len(s.posts))
	
	return nil
}

func (s *service) Stop(ctx context.Context) error {
	s.logger.Printf("BLOG: Stopping blog service...")
	return nil
}

func (s *service) Health(ctx context.Context) error {
	if len(s.posts) == 0 {
		return fmt.Errorf("no blog posts loaded")
	}
	return nil
}

// GetAll returns all blog posts
func (s *service) GetAll(ctx context.Context) []Post {
	// Return a copy to prevent modification
	result := make([]Post, len(s.posts))
	copy(result, s.posts)
	return result
}

// GetBySlug returns a blog post by slug
func (s *service) GetBySlug(ctx context.Context, slug string) (*Post, error) {
	post, exists := s.postMap[slug]
	if !exists {
		return nil, errors.NotFound("blog post")
	}
	
	// Return a copy
	result := *post
	return &result, nil
}

// Search searches blog posts
func (s *service) Search(ctx context.Context, query string) []Post {
	if query == "" {
		return s.GetAll(ctx)
	}
	
	query = strings.ToLower(query)
	var results []Post
	
	for _, post := range s.posts {
		// Search in title, summary, and tags
		if strings.Contains(strings.ToLower(post.Title), query) ||
			strings.Contains(strings.ToLower(post.Summary), query) ||
			s.containsTag(post.Tags, query) {
			results = append(results, post)
		}
	}
	
	return results
}

// GetByTag returns posts with a specific tag
func (s *service) GetByTag(ctx context.Context, tag string) []Post {
	indices, exists := s.tagIndex[strings.ToLower(tag)]
	if !exists {
		return []Post{}
	}
	
	results := make([]Post, len(indices))
	for i, idx := range indices {
		results[i] = s.posts[idx]
	}
	
	return results
}

// GetTags returns all unique tags
func (s *service) GetTags(ctx context.Context) []string {
	var tags []string
	for tag := range s.tagIndex {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// LoadPosts loads blog posts from the filesystem
func (s *service) LoadPosts(ctx context.Context) error {
	s.posts = []Post{}
	s.postMap = make(map[string]*Post)
	s.tagIndex = make(map[string][]int)
	
	// Read blog directory
	files, err := fs.ReadDir(s.blogFS, s.blogDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeIO, "failed to read blog directory")
	}
	
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		
		filePath := file.Name()
		if s.blogDir != "." {
			filePath = s.blogDir + "/" + file.Name()
		}
		post, err := s.loadMarkdownPost(filePath)
		if err != nil {
			s.logger.Printf("BLOG: Warning - failed to load %s: %v", file.Name(), err)
			continue
		}
		
		// Add to collections
		s.posts = append(s.posts, *post)
		s.postMap[post.Slug] = post
		s.logger.Printf("BLOG: Loaded post with slug: '%s'", post.Slug)
	}
	
	// Sort posts by date (newest first)
	sort.Slice(s.posts, func(i, j int) bool {
		return s.posts[i].Date.After(s.posts[j].Date)
	})
	
	// Build tag index AFTER sorting
	for idx, post := range s.posts {
		for _, tag := range post.Tags {
			tagLower := strings.ToLower(tag)
			s.tagIndex[tagLower] = append(s.tagIndex[tagLower], idx)
		}
	}
	
	s.logger.Printf("BLOG: Loaded %d blog posts", len(s.posts))
	
	// Publish event
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, events.NewEventWithContext(ctx,
			events.EventBlogPublished,
			map[string]interface{}{
				"count": len(s.posts),
			},
		))
	}
	
	return nil
}

// loadMarkdownPost loads a single markdown post
func (s *service) loadMarkdownPost(filename string) (*Post, error) {
	// Read file
	content, err := fs.ReadFile(s.blogFS, filename)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeIO, "failed to read file")
	}
	
	// Parse frontmatter and content
	frontmatter, markdownContent, err := s.parseFrontmatter(content)
	if err != nil {
		return nil, err
	}
	
	// Convert markdown to HTML
	htmlContent := s.markdownToHTML(markdownContent)
	
	// Generate slug from filename (strip directory path and extension)
	baseName := filepath.Base(filename)
	slug := strings.TrimSuffix(baseName, ".md")
	
	// Calculate reading time if not provided
	readingTime := frontmatter.ReadingTime
	if readingTime == 0 {
		readingTime = s.calculateReadingTime(string(markdownContent))
	}
	
	return &Post{
		Slug:        slug,
		Title:       frontmatter.Title,
		Date:        frontmatter.Date,
		Summary:     frontmatter.Summary,
		Content:     template.HTML(htmlContent),
		ReadingTime: readingTime,
		Tags:        frontmatter.Tags,
		FileName:    filename,
	}, nil
}

// parseFrontmatter parses YAML frontmatter from markdown content
func (s *service) parseFrontmatter(content []byte) (*Frontmatter, []byte, error) {
	// Check if content starts with frontmatter delimiter
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return nil, nil, errors.New(errors.ErrCodeInvalidFormat, "missing frontmatter delimiter")
	}
	
	// Find the end of frontmatter
	endDelimiter := []byte("\n---\n")
	endIndex := bytes.Index(content[4:], endDelimiter)
	if endIndex == -1 {
		return nil, nil, errors.New(errors.ErrCodeInvalidFormat, "missing frontmatter end delimiter")
	}
	
	// Extract frontmatter and content
	frontmatterBytes := content[4 : endIndex+4]
	markdownContent := content[endIndex+8:] // Skip past "\n---\n"
	
	// Parse YAML frontmatter
	var frontmatter Frontmatter
	if err := yaml.Unmarshal(frontmatterBytes, &frontmatter); err != nil {
		return nil, nil, errors.Wrap(err, errors.ErrCodeInvalidFormat, "failed to parse YAML frontmatter")
	}
	
	return &frontmatter, markdownContent, nil
}

// markdownToHTML converts markdown to HTML with syntax highlighting
func (s *service) markdownToHTML(mdContent []byte) string {
	// Configure markdown parser
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	
	// Configure HTML renderer with syntax highlighting
	htmlFlags := mdhtml.CommonFlags | mdhtml.HrefTargetBlank
	opts := mdhtml.RendererOptions{
		Flags: htmlFlags,
		RenderNodeHook: s.chromaRenderHook,
	}
	renderer := mdhtml.NewRenderer(opts)
	
	// Convert markdown to HTML
	return string(markdown.ToHTML(mdContent, p, renderer))
}

// chromaRenderHook provides syntax highlighting for code blocks
func (s *service) chromaRenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	if code, ok := node.(*ast.CodeBlock); ok && entering {
		// Get the language from the code block info
		language := ""
		if code.Info != nil {
			language = string(code.Info)
		}
		
		// Handle Mermaid diagrams (client-side rendering for now)
		if language == "mermaid" {
			w.Write([]byte(`<div class="mermaid-container mermaid-csr"><div class="mermaid">`))
			w.Write(code.Literal)
			w.Write([]byte("</div></div>"))
			return ast.GoToNext, true
		}
		
		// Get lexer for the language
		lexer := lexers.Get(language)
		if lexer == nil {
			lexer = lexers.Fallback
		}
		
		// Configure formatter with cyberpunk theme
		formatter := html.New(html.WithClasses(true), html.TabWidth(2))
		style := styles.Get("monokai")
		if style == nil {
			style = styles.Fallback
		}
		
		// Create iterator from code content
		iterator, err := lexer.Tokenise(nil, string(code.Literal))
		if err != nil {
			// Fallback to plain text
			w.Write([]byte("<pre><code>"))
			w.Write(code.Literal)
			w.Write([]byte("</code></pre>"))
			return ast.GoToNext, true
		}
		
		// Format the code
		err = formatter.Format(w, style, iterator)
		if err != nil {
			// Fallback to plain text
			w.Write([]byte("<pre><code>"))
			w.Write(code.Literal)
			w.Write([]byte("</code></pre>"))
		}
		
		return ast.GoToNext, true
	}
	
	return ast.GoToNext, false
}

// calculateReadingTime calculates reading time based on word count
func (s *service) calculateReadingTime(text string) int {
	// Average reading speed: 200 words per minute
	words := len(strings.Fields(text))
	minutes := int(math.Ceil(float64(words) / 200.0))
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}

// containsTag checks if tags contain a query (case-insensitive)
func (s *service) containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// loadBlogConfig loads the blog configuration from blog.yml
func (s *service) loadBlogConfig() {
	configPath := "content/blog.yml"
	configData, err := fs.ReadFile(s.blogFS, configPath)
	if err != nil {
		s.logger.Printf("BLOG: Warning - could not load blog config from %s: %v, using defaults", configPath, err)
		s.blogConfig = s.getDefaultBlogConfig()
		return
	}

	var config BlogConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		s.logger.Printf("BLOG: Warning - could not parse blog config: %v, using defaults", err)
		s.blogConfig = s.getDefaultBlogConfig()
		return
	}

	s.blogConfig = &config
	s.logger.Printf("BLOG: Loaded blog configuration with %d tag filters", len(config.Blog.TagFilters))
}

// getDefaultBlogConfig returns default blog configuration
func (s *service) getDefaultBlogConfig() *BlogConfig {
	return &BlogConfig{
		Blog: struct {
			Title      string       `yaml:"title"`
			Subtitle   string       `yaml:"subtitle"`
			TagFilters []TagFilter  `yaml:"tag_filters"`
			Search     SearchConfig `yaml:"search"`
		}{
			Title:    "Technical Insights",
			Subtitle: "Deep dives into blockchain, AI, and production engineering.",
			TagFilters: []TagFilter{
				{Display: "All", Tag: "all", Active: true},
				{Display: "Blockchain", Tag: "blockchain"},
				{Display: "AI/ML", Tag: "ai"},
				{Display: "Go", Tag: "golang"},
			},
			Search: SearchConfig{
				Placeholder:   "Search posts...",
				CaseSensitive: false,
			},
		},
	}
}

// GetBlogConfig returns the blog configuration
func (s *service) GetBlogConfig() *BlogConfig {
	return s.blogConfig
}

// Ensure service implements required interfaces
var (
	_ Service          = (*service)(nil)
	_ registry.Service = (*service)(nil)
)