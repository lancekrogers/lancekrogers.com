package blog

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"math"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"blockhead.consulting/internal/errors"
	"blockhead.consulting/internal/events"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

// SQLiteService implements blog service using SQLite database
type SQLiteService struct {
	db         *sql.DB
	blogFS     fs.FS
	blogDir    string
	logger     *log.Logger
	eventBus   events.EventBus
	blogConfig *BlogConfig
	ftsEnabled bool
}

// NewSQLiteService creates a new SQLite-based blog service
func NewSQLiteService(dbPath string, blogFS fs.FS, blogDir string, logger *log.Logger, eventBus events.EventBus) (*SQLiteService, error) {
	if logger == nil {
		logger = log.Default()
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", dbPath+"?_busy_timeout=10000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	s := &SQLiteService{
		db:       db,
		blogFS:   blogFS,
		blogDir:  blogDir,
		logger:   logger,
		eventBus: eventBus,
	}

	// Initialize database schema
	if err := s.initializeSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Load blog configuration
	s.loadBlogConfig()

	return s, nil
}

// initializeSchema creates the database schema with optional FTS5 support
func (s *SQLiteService) initializeSchema() error {
	// Main posts table
	schema := `
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT UNIQUE NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			summary TEXT,
			published_at DATETIME NOT NULL,
			tags TEXT, -- JSON array: ["tag1", "tag2"]
			category TEXT,
			file_path TEXT NOT NULL,
			word_count INTEGER NOT NULL DEFAULT 0,
			reading_time INTEGER NOT NULL DEFAULT 1,
			featured_image TEXT,
			og_description TEXT,
			twitter_handle TEXT,
			file_mod_time INTEGER NOT NULL DEFAULT 0, -- Unix timestamp of file modification
			content_hash TEXT NOT NULL DEFAULT '', -- SHA256 hash of file content
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Performance indexes
		CREATE INDEX IF NOT EXISTS idx_posts_published_at ON posts(published_at DESC);
		CREATE INDEX IF NOT EXISTS idx_posts_category ON posts(category);
		CREATE INDEX IF NOT EXISTS idx_posts_slug ON posts(slug);
		CREATE INDEX IF NOT EXISTS idx_posts_file_path ON posts(file_path);
		CREATE INDEX IF NOT EXISTS idx_posts_content_hash ON posts(content_hash);
		
		-- Post paths table for tracking historical URLs
		CREATE TABLE IF NOT EXISTS post_paths (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content_hash TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			is_current BOOLEAN NOT NULL DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (content_hash) REFERENCES posts(content_hash) ON DELETE CASCADE
		);
		
		-- Performance indexes for post_paths
		CREATE INDEX IF NOT EXISTS idx_post_paths_content_hash ON post_paths(content_hash);
		CREATE INDEX IF NOT EXISTS idx_post_paths_path ON post_paths(path);
		CREATE INDEX IF NOT EXISTS idx_post_paths_current ON post_paths(is_current) WHERE is_current = TRUE;
	`

	_, err := s.db.Exec(schema)
	if err != nil {
		return err
	}

	// Run schema migrations for existing databases
	if err := s.runMigrations(); err != nil {
		return err
	}

	// Try to create FTS5 tables - if it fails, we'll fall back to basic search
	ftsSchema := `
		-- Full-text search index using FTS5
		CREATE VIRTUAL TABLE IF NOT EXISTS posts_fts USING fts5(
			title, content, summary, tags,
			content=posts,
			content_rowid=id
		);

		-- Triggers to keep FTS in sync
		CREATE TRIGGER IF NOT EXISTS posts_ai AFTER INSERT ON posts BEGIN
			INSERT INTO posts_fts(rowid, title, content, summary, tags) 
			VALUES (new.id, new.title, new.content, new.summary, new.tags);
		END;

		CREATE TRIGGER IF NOT EXISTS posts_ad AFTER DELETE ON posts BEGIN
			INSERT INTO posts_fts(posts_fts, rowid, title, content, summary, tags) 
			VALUES('delete', old.id, old.title, old.content, old.summary, old.tags);
		END;

		CREATE TRIGGER IF NOT EXISTS posts_au AFTER UPDATE ON posts BEGIN
			INSERT INTO posts_fts(posts_fts, rowid, title, content, summary, tags) 
			VALUES('delete', old.id, old.title, old.content, old.summary, old.tags);
			INSERT INTO posts_fts(rowid, title, content, summary, tags) 
			VALUES (new.id, new.title, new.content, new.summary, new.tags);
		END;
	`

	_, err = s.db.Exec(ftsSchema)
	if err != nil {
		s.logger.Printf("BLOG: FTS5 not available, falling back to basic search: %v", err)
		s.ftsEnabled = false
	} else {
		s.ftsEnabled = true
		s.logger.Printf("BLOG: FTS5 search enabled")
	}

	return nil
}

// runMigrations applies schema changes for existing databases
func (s *SQLiteService) runMigrations() error {
	// Column migrations for posts table
	columnMigrations := []string{
		`ALTER TABLE posts ADD COLUMN file_mod_time INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE posts ADD COLUMN content_hash TEXT NOT NULL DEFAULT ''`,
	}

	for _, migration := range columnMigrations {
		_, err := s.db.Exec(migration)
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("column migration failed: %v", err)
		}
	}

	// Table migrations - create post_paths table if it doesn't exist
	tableMigrations := []string{
		`CREATE TABLE IF NOT EXISTS post_paths (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content_hash TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			is_current BOOLEAN NOT NULL DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (content_hash) REFERENCES posts(content_hash) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_post_paths_content_hash ON post_paths(content_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_post_paths_path ON post_paths(path)`,
		`CREATE INDEX IF NOT EXISTS idx_post_paths_current ON post_paths(is_current) WHERE is_current = TRUE`,
		`CREATE INDEX IF NOT EXISTS idx_posts_content_hash ON posts(content_hash)`,
	}

	for _, migration := range tableMigrations {
		_, err := s.db.Exec(migration)
		if err != nil {
			return fmt.Errorf("table migration failed: %v", err)
		}
	}

	return nil
}

// Implement the Service interface methods

func (s *SQLiteService) Name() string {
	return "blog-sqlite"
}

func (s *SQLiteService) Start(ctx context.Context) error {
	s.logger.Printf("BLOG: Starting SQLite blog service...")

	// Import posts on startup
	if err := s.ImportMarkdownFiles(ctx); err != nil {
		return fmt.Errorf("failed to import markdown files: %w", err)
	}

	// Get post count
	count, err := s.getPostCount()
	if err != nil {
		return fmt.Errorf("failed to get post count: %w", err)
	}

	s.logger.Printf("BLOG: SQLite blog service started with %d posts", count)
	return nil
}

func (s *SQLiteService) Stop(ctx context.Context) error {
	s.logger.Printf("BLOG: Stopping SQLite blog service...")
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLiteService) Health(ctx context.Context) error {
	count, err := s.getPostCount()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("no blog posts loaded")
	}
	return nil
}

func (s *SQLiteService) GetAll(ctx context.Context) []Post {
	params := PaginationParams{
		Page:    1,
		PerPage: 1000, // Get all posts
	}

	result, err := s.GetPaginatedPosts(ctx, params)
	if err != nil {
		s.logger.Printf("Error getting all posts: %v", err)
		return []Post{}
	}

	return result.Posts
}

func (s *SQLiteService) GetBySlug(ctx context.Context, slug string) (*Post, error) {
	query := `
		SELECT id, slug, title, content, summary, published_at, tags, category, 
		       file_path, word_count, reading_time, created_at, updated_at
		FROM posts 
		WHERE slug = ?
	`

	row := s.db.QueryRowContext(ctx, query, slug)

	var post Post
	var tagsJSON string
	err := row.Scan(
		&post.ID, &post.Slug, &post.Title, &post.Content, &post.Summary,
		&post.Date, &tagsJSON, &post.Category, &post.FileName,
		&post.WordCount, &post.ReadingTime, &post.CreatedAt, &post.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("blog post")
		}
		return nil, err
	}

	// Parse tags JSON
	if tagsJSON != "" {
		json.Unmarshal([]byte(tagsJSON), &post.Tags)
	}

	// Convert summary markdown to HTML
	s.populateSummaryHTML(&post)

	return &post, nil
}

func (s *SQLiteService) Search(ctx context.Context, query string) []Post {
	if query == "" {
		return s.GetAll(ctx)
	}

	result, err := s.SearchPosts(ctx, query, 1, 100)
	if err != nil {
		s.logger.Printf("Error searching posts: %v", err)
		return []Post{}
	}

	return result.Posts
}

func (s *SQLiteService) GetByTag(ctx context.Context, tag string) []Post {
	params := PaginationParams{
		Page:    1,
		PerPage: 1000,
		Tag:     tag,
	}

	result, err := s.GetPaginatedPosts(ctx, params)
	if err != nil {
		s.logger.Printf("Error getting posts by tag: %v", err)
		return []Post{}
	}

	return result.Posts
}

func (s *SQLiteService) GetTags(ctx context.Context) []string {
	tagCounts, err := s.GetTagCounts(ctx)
	if err != nil {
		s.logger.Printf("Error getting tags: %v", err)
		return []string{}
	}

	tags := make([]string, len(tagCounts))
	for i, tc := range tagCounts {
		tags[i] = tc.Tag
	}

	return tags
}

func (s *SQLiteService) LoadPosts(ctx context.Context) error {
	return s.ImportMarkdownFiles(ctx)
}

func (s *SQLiteService) GetBlogConfig() *BlogConfig {
	return s.blogConfig
}

// New SQLite-specific methods

// GetPaginatedPosts returns paginated blog posts with filtering
func (s *SQLiteService) GetPaginatedPosts(ctx context.Context, params PaginationParams) (*PaginatedPosts, error) {
	// Default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 12
	}

	// Build WHERE clause
	var whereClause strings.Builder
	var args []interface{}
	whereClause.WriteString("WHERE 1=1")

	if params.Category != "" {
		whereClause.WriteString(" AND category = ?")
		args = append(args, params.Category)
	}

	if params.Tag != "" {
		whereClause.WriteString(" AND tags LIKE ?")
		args = append(args, "%"+params.Tag+"%")
	}

	if params.Search != "" {
		if s.ftsEnabled {
			// Use FTS5 for search
			whereClause.WriteString(" AND id IN (SELECT rowid FROM posts_fts WHERE posts_fts MATCH ?)")
			args = append(args, s.sanitizeSearchQuery(params.Search))
		} else {
			// Fall back to basic LIKE search
			whereClause.WriteString(" AND (title LIKE ? OR content LIKE ? OR summary LIKE ?)")
			searchTerm := "%" + params.Search + "%"
			args = append(args, searchTerm, searchTerm, searchTerm)
		}
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM posts " + whereClause.String()
	var totalPosts int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalPosts)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(totalPosts) / float64(params.PerPage)))
	offset := (params.Page - 1) * params.PerPage

	// Get posts for current page
	postsQuery := `
		SELECT id, slug, title, content, summary, published_at, tags, category, 
		       file_path, word_count, reading_time, featured_image, og_description, twitter_handle,
		       created_at, updated_at
		FROM posts ` + whereClause.String() + `
		ORDER BY published_at DESC
		LIMIT ? OFFSET ?
	`

	queryArgs := append(args, params.PerPage, offset)
	rows, err := s.db.QueryContext(ctx, postsQuery, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var tagsJSON string
		err := rows.Scan(
			&post.ID, &post.Slug, &post.Title, &post.Content, &post.Summary,
			&post.Date, &tagsJSON, &post.Category, &post.FileName,
			&post.WordCount, &post.ReadingTime, &post.FeaturedImage, &post.OGDescription, &post.TwitterHandle,
			&post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse tags JSON
		if tagsJSON != "" {
			json.Unmarshal([]byte(tagsJSON), &post.Tags)
		}

		// Convert summary markdown to HTML
		s.populateSummaryHTML(&post)

		posts = append(posts, post)
	}

	return &PaginatedPosts{
		Posts:       posts,
		TotalPosts:  totalPosts,
		TotalPages:  totalPages,
		CurrentPage: params.Page,
		HasPrev:     params.Page > 1,
		HasNext:     params.Page < totalPages,
		PrevPage:    params.Page - 1,
		NextPage:    params.Page + 1,
	}, nil
}

// SearchPosts performs full-text search using FTS5 or falls back to basic search
func (s *SQLiteService) SearchPosts(ctx context.Context, query string, page, perPage int) (*PaginatedPosts, error) {
	if query == "" {
		return &PaginatedPosts{}, nil
	}

	offset := (page - 1) * perPage

	if s.ftsEnabled {
		return s.searchWithFTS5(ctx, query, page, perPage, offset)
	} else {
		return s.searchWithLike(ctx, query, page, perPage, offset)
	}
}

func (s *SQLiteService) searchWithFTS5(ctx context.Context, query string, page, perPage, offset int) (*PaginatedPosts, error) {
	// Sanitize search query for FTS5
	searchQuery := s.sanitizeSearchQuery(query)

	// Count total matches
	countSQL := `
		SELECT COUNT(*) 
		FROM posts_fts 
		WHERE posts_fts MATCH ?
	`
	var totalPosts int
	err := s.db.QueryRowContext(ctx, countSQL, searchQuery).Scan(&totalPosts)
	if err != nil {
		return nil, err
	}

	// Get ranked search results
	searchSQL := `
		SELECT p.id, p.slug, p.title, p.content, p.summary, p.published_at, 
		       p.tags, p.category, p.file_path, p.word_count, p.reading_time,
		       p.featured_image, p.og_description, p.twitter_handle,
		       p.created_at, p.updated_at, bm25(posts_fts) as rank
		FROM posts_fts
		JOIN posts p ON posts_fts.rowid = p.id
		WHERE posts_fts MATCH ?
		ORDER BY rank, p.published_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, searchSQL, searchQuery, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var tagsJSON string
		var rank float64

		err := rows.Scan(
			&post.ID, &post.Slug, &post.Title, &post.Content, &post.Summary,
			&post.Date, &tagsJSON, &post.Category, &post.FileName,
			&post.WordCount, &post.ReadingTime, &post.FeaturedImage, &post.OGDescription, &post.TwitterHandle,
			&post.CreatedAt, &post.UpdatedAt, &rank,
		)
		if err != nil {
			return nil, err
		}

		// Parse tags JSON
		if tagsJSON != "" {
			json.Unmarshal([]byte(tagsJSON), &post.Tags)
		}

		// Convert summary markdown to HTML
		s.populateSummaryHTML(&post)

		post.SearchRank = rank
		posts = append(posts, post)
	}

	totalPages := int(math.Ceil(float64(totalPosts) / float64(perPage)))

	return &PaginatedPosts{
		Posts:       posts,
		TotalPosts:  totalPosts,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
		PrevPage:    page - 1,
		NextPage:    page + 1,
	}, nil
}

func (s *SQLiteService) searchWithLike(ctx context.Context, query string, page, perPage, offset int) (*PaginatedPosts, error) {
	searchTerm := "%" + query + "%"

	// Count total matches
	countSQL := `
		SELECT COUNT(*) 
		FROM posts 
		WHERE title LIKE ? OR content LIKE ? OR summary LIKE ?
	`
	var totalPosts int
	err := s.db.QueryRowContext(ctx, countSQL, searchTerm, searchTerm, searchTerm).Scan(&totalPosts)
	if err != nil {
		return nil, err
	}

	// Get search results
	searchSQL := `
		SELECT id, slug, title, content, summary, published_at, 
		       tags, category, file_path, word_count, reading_time,
		       featured_image, og_description, twitter_handle,
		       created_at, updated_at
		FROM posts
		WHERE title LIKE ? OR content LIKE ? OR summary LIKE ?
		ORDER BY published_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, searchSQL, searchTerm, searchTerm, searchTerm, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var tagsJSON string

		err := rows.Scan(
			&post.ID, &post.Slug, &post.Title, &post.Content, &post.Summary,
			&post.Date, &tagsJSON, &post.Category, &post.FileName,
			&post.WordCount, &post.ReadingTime, &post.FeaturedImage, &post.OGDescription, &post.TwitterHandle,
			&post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse tags JSON
		if tagsJSON != "" {
			json.Unmarshal([]byte(tagsJSON), &post.Tags)
		}

		// Convert summary markdown to HTML
		s.populateSummaryHTML(&post)

		posts = append(posts, post)
	}

	totalPages := int(math.Ceil(float64(totalPosts) / float64(perPage)))

	return &PaginatedPosts{
		Posts:       posts,
		TotalPosts:  totalPosts,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
		PrevPage:    page - 1,
		NextPage:    page + 1,
	}, nil
}

// GetPostNavigation returns navigation links for a post
func (s *SQLiteService) GetPostNavigation(ctx context.Context, slug string) (*PostNavigation, error) {
	current, err := s.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Get previous post (newer)
	var previous *Post
	prevQuery := `
		SELECT id, slug, title, summary, published_at
		FROM posts 
		WHERE published_at > ? 
		ORDER BY published_at ASC 
		LIMIT 1
	`
	row := s.db.QueryRowContext(ctx, prevQuery, current.Date)
	prev := &Post{}
	err = row.Scan(&prev.ID, &prev.Slug, &prev.Title, &prev.Summary, &prev.Date)
	if err == nil {
		s.populateSummaryHTML(prev)
		previous = prev
	}

	// Get next post (older)
	var next *Post
	nextQuery := `
		SELECT id, slug, title, summary, published_at
		FROM posts 
		WHERE published_at < ? 
		ORDER BY published_at DESC 
		LIMIT 1
	`
	row = s.db.QueryRowContext(ctx, nextQuery, current.Date)
	nxt := &Post{}
	err = row.Scan(&nxt.ID, &nxt.Slug, &nxt.Title, &nxt.Summary, &nxt.Date)
	if err == nil {
		s.populateSummaryHTML(nxt)
		next = nxt
	}

	return &PostNavigation{
		Previous: previous,
		Current:  current,
		Next:     next,
	}, nil
}

// GetRelatedPosts finds related posts by tags and category
func (s *SQLiteService) GetRelatedPosts(ctx context.Context, slug string, limit int) ([]*RelatedPost, error) {
	current, err := s.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Get all other posts for comparison
	query := `
		SELECT id, slug, title, summary, tags, category, published_at
		FROM posts 
		WHERE slug != ?
		ORDER BY published_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []Post
	for rows.Next() {
		var post Post
		var tagsJSON string
		err := rows.Scan(&post.ID, &post.Slug, &post.Title, &post.Summary, &tagsJSON, &post.Category, &post.Date)
		if err != nil {
			continue
		}

		if tagsJSON != "" {
			json.Unmarshal([]byte(tagsJSON), &post.Tags)
		}

		// Convert summary markdown to HTML
		s.populateSummaryHTML(&post)

		candidates = append(candidates, post)
	}

	// Calculate relevance scores
	var related []*RelatedPost
	for _, candidate := range candidates {
		relevance := s.calculateRelevance(current, &candidate)
		if relevance > 0 {
			related = append(related, &RelatedPost{
				Post:      &candidate,
				Relevance: relevance,
			})
		}
	}

	// Sort by relevance
	sort.Slice(related, func(i, j int) bool {
		return related[i].Relevance > related[j].Relevance
	})

	// Limit results
	if len(related) > limit {
		related = related[:limit]
	}

	return related, nil
}

// GetTagCounts returns all tags with their usage counts
func (s *SQLiteService) GetTagCounts(ctx context.Context) ([]TagCount, error) {
	query := "SELECT tags FROM posts WHERE tags IS NOT NULL AND tags != ''"

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagMap := make(map[string]int)

	for rows.Next() {
		var tagsJSON string
		if err := rows.Scan(&tagsJSON); err != nil {
			continue
		}

		var tags []string
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
			continue
		}

		for _, tag := range tags {
			tagMap[strings.ToLower(tag)]++
		}
	}

	var tagCounts []TagCount
	for tag, count := range tagMap {
		tagCounts = append(tagCounts, TagCount{
			Tag:   tag,
			Count: count,
		})
	}

	// Sort by count (descending) then by name
	sort.Slice(tagCounts, func(i, j int) bool {
		if tagCounts[i].Count == tagCounts[j].Count {
			return tagCounts[i].Tag < tagCounts[j].Tag
		}
		return tagCounts[i].Count > tagCounts[j].Count
	})

	return tagCounts, nil
}

// Helper methods

func (s *SQLiteService) getPostCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&count)
	return count, err
}

func (s *SQLiteService) sanitizeSearchQuery(query string) string {
	// Escape special FTS5 characters
	query = strings.ReplaceAll(query, `"`, `""`)

	// Support phrase search with quotes
	if strings.Contains(query, " ") && !strings.HasPrefix(query, `"`) {
		return `"` + query + `"`
	}

	return query
}

func (s *SQLiteService) calculateRelevance(current, candidate *Post) float64 {
	relevance := 0.0

	// Same category bonus
	if current.Category != "" && current.Category == candidate.Category {
		relevance += 2.0
	}

	// Shared tags bonus
	for _, currentTag := range current.Tags {
		for _, candidateTag := range candidate.Tags {
			if strings.EqualFold(currentTag, candidateTag) {
				relevance += 1.0
			}
		}
	}

	// Recency factor (newer posts get slight bonus)
	daysDiff := time.Since(candidate.Date).Hours() / 24
	if daysDiff < 30 {
		relevance += 0.5 * (30 - daysDiff) / 30
	}

	return relevance
}

func (s *SQLiteService) loadBlogConfig() {
	// Skip loading if blogFS is nil (for tests)
	if s.blogFS == nil {
		s.blogConfig = s.getDefaultBlogConfig()
		return
	}
	
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

func (s *SQLiteService) getDefaultBlogConfig() *BlogConfig {
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

// populateSummaryHTML converts the Summary markdown to HTML for a post
func (s *SQLiteService) populateSummaryHTML(post *Post) {
	if post.Summary != "" {
		summaryHTML := s.markdownToHTML([]byte(post.Summary))
		post.SummaryHTML = template.HTML(summaryHTML)
	}
}

// ImportMarkdownFiles imports markdown files to SQLite database with smart change detection
func (s *SQLiteService) ImportMarkdownFiles(ctx context.Context) error {
	s.logger.Printf("BLOG: Starting intelligent markdown import...")

	// Get existing file tracking data from database
	existingFiles, err := s.getExistingFilesData(ctx)
	if err != nil {
		return fmt.Errorf("failed to get existing files data: %w", err)
	}

	var newPosts []Post
	var updatedPosts []Post
	var removedFiles []string
	processedFiles := make(map[string]bool)

	// Walk through filesystem and check for changes
	err = fs.WalkDir(s.blogFS, s.blogDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip drafts
		if strings.Contains(path, "/drafts/") {
			return nil
		}

		processedFiles[path] = true

		// Read file content for hash calculation
		content, err := fs.ReadFile(s.blogFS, path)
		if err != nil {
			s.logger.Printf("Error reading %s: %v", path, err)
			return nil
		}

		// Get file info for change detection
		fileInfo, err := fs.Stat(s.blogFS, path)
		if err != nil {
			s.logger.Printf("Error getting file info for %s: %v", path, err)
			return nil
		}

		// Calculate content hash
		hash := sha256.Sum256(content)
		contentHash := hex.EncodeToString(hash[:])
		modTime := fileInfo.ModTime().Unix()

		// Check if file has changed
		existingFile, exists := existingFiles[path]
		hasChanged := !exists ||
			existingFile.ModTime != modTime ||
			existingFile.ContentHash != contentHash

		if !hasChanged {
			// File hasn't changed, skip processing
			return nil
		}

		// Parse the markdown file
		post, err := s.parseMarkdownFileWithMetadata(path, content, modTime, contentHash)
		if err != nil {
			s.logger.Printf("Error parsing %s: %v", path, err)
			return nil
		}

		// Extract metadata from path
		post.Category = s.extractCategoryFromPath(path)
		post.FileName = path

		// Calculate reading metrics
		post.ReadingTime, post.WordCount = s.calculateReadingTime(string(post.Content))

		if exists {
			updatedPosts = append(updatedPosts, *post)
		} else {
			newPosts = append(newPosts, *post)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Find removed files (exist in DB but not in filesystem)
	for filePath := range existingFiles {
		if !processedFiles[filePath] {
			removedFiles = append(removedFiles, filePath)
		}
	}

	// Process changes
	totalChanges := len(newPosts) + len(updatedPosts) + len(removedFiles)
	if totalChanges == 0 {
		s.logger.Printf("BLOG: No changes detected in markdown files")
		return nil
	}

	s.logger.Printf("BLOG: Processing changes - %d new, %d updated, %d removed",
		len(newPosts), len(updatedPosts), len(removedFiles))

	// Insert new posts
	if len(newPosts) > 0 {
		if err := s.batchInsertPosts(ctx, newPosts); err != nil {
			return fmt.Errorf("failed to insert new posts: %w", err)
		}
	}

	// Update existing posts
	if len(updatedPosts) > 0 {
		if err := s.batchUpdatePosts(ctx, updatedPosts); err != nil {
			return fmt.Errorf("failed to update posts: %w", err)
		}
	}

	// Remove deleted posts
	if len(removedFiles) > 0 {
		if err := s.removePostsByFilePath(ctx, removedFiles); err != nil {
			return fmt.Errorf("failed to remove deleted posts: %w", err)
		}
	}

	// Update path history only for new posts (updated posts keep same slug/path)
	if len(newPosts) > 0 {
		if err := s.updatePathHistory(ctx, newPosts); err != nil {
			s.logger.Printf("BLOG: Warning - failed to update path history: %v", err)
			// Don't fail the entire import for path history issues
		}
	}

	// Cleanup orphaned paths after processing
	if err := s.CleanupOrphanedPaths(ctx); err != nil {
		s.logger.Printf("BLOG: Warning - failed to cleanup orphaned paths: %v", err)
	}

	s.logger.Printf("BLOG: Import completed - %d changes processed", totalChanges)

	// Publish event
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, events.NewEventWithContext(ctx,
			events.EventBlogPublished,
			map[string]interface{}{
				"new_count":     len(newPosts),
				"updated_count": len(updatedPosts),
				"removed_count": len(removedFiles),
			},
		))
	}

	return nil
}

func (s *SQLiteService) parseMarkdownFile(filename string) (*Post, error) {
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

	// Generate slug from filename
	baseName := filepath.Base(filename)
	slug := strings.TrimSuffix(baseName, ".md")

	// Calculate reading time if not provided
	readingTime := frontmatter.ReadingTime
	if readingTime == 0 {
		readingTime, _ = s.calculateReadingTime(string(markdownContent))
	}

	return &Post{
		Slug:          slug,
		Title:         frontmatter.Title,
		Date:          frontmatter.Date,
		Summary:       frontmatter.Summary,
		Content:       template.HTML(htmlContent),
		ReadingTime:   readingTime,
		Tags:          frontmatter.Tags,
		FileName:      filename,
		FeaturedImage: frontmatter.FeaturedImage,
		OGDescription: frontmatter.OGDescription,
		TwitterHandle: frontmatter.TwitterHandle,
	}, nil
}

func (s *SQLiteService) extractCategoryFromPath(path string) string {
	dir := filepath.Dir(path)
	parts := strings.Split(dir, "/")

	// Skip date directories and blog root
	dateRegex := regexp.MustCompile(`^\d{4}$|^\d{2}$`)
	for _, part := range parts {
		if part != "blog" && part != "content" && !dateRegex.MatchString(part) {
			return part
		}
	}
	return ""
}

func (s *SQLiteService) batchInsertPosts(ctx context.Context, posts []Post) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO posts 
		(slug, title, content, summary, published_at, tags, category, file_path, word_count, reading_time, featured_image, og_description, twitter_handle, file_mod_time, content_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, post := range posts {
		tagsJSON, _ := json.Marshal(post.Tags)
		_, err = stmt.ExecContext(ctx,
			post.Slug, post.Title, string(post.Content), post.Summary,
			post.Date, string(tagsJSON), post.Category, post.FileName,
			post.WordCount, post.ReadingTime, post.FeaturedImage, post.OGDescription, post.TwitterHandle,
			post.FileModTime, post.ContentHash,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Reuse existing parsing methods from the original service
func (s *SQLiteService) parseFrontmatter(content []byte) (*Frontmatter, []byte, error) {
	// Check if content starts with frontmatter delimiter
	if !strings.HasPrefix(string(content), "---\n") {
		return nil, nil, errors.New(errors.ErrCodeInvalidFormat, "missing frontmatter delimiter")
	}

	// Find the end of frontmatter
	endDelimiter := "\n---\n"
	endIndex := strings.Index(string(content[4:]), endDelimiter)
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

func (s *SQLiteService) markdownToHTML(mdContent []byte) string {
	// Configure markdown parser
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	// Configure HTML renderer with syntax highlighting
	htmlFlags := mdhtml.CommonFlags | mdhtml.HrefTargetBlank
	opts := mdhtml.RendererOptions{
		Flags:          htmlFlags,
		RenderNodeHook: s.chromaRenderHook,
	}
	renderer := mdhtml.NewRenderer(opts)

	// Convert markdown to HTML
	return string(markdown.ToHTML(mdContent, p, renderer))
}

// chromaRenderHook provides syntax highlighting for code blocks
func (s *SQLiteService) chromaRenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
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

func (s *SQLiteService) calculateReadingTime(text string) (int, int) {
	// Average reading speed: 225 words per minute
	words := len(strings.Fields(text))
	minutes := int(math.Ceil(float64(words) / 225.0))
	if minutes < 1 {
		minutes = 1
	}
	return minutes, words
}

// FileMetadata holds file tracking information
type FileMetadata struct {
	FilePath    string
	ModTime     int64
	ContentHash string
}

// getExistingFilesData retrieves file metadata from database for change detection
func (s *SQLiteService) getExistingFilesData(ctx context.Context) (map[string]FileMetadata, error) {
	query := `SELECT file_path, file_mod_time, content_hash FROM posts`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filesData := make(map[string]FileMetadata)

	for rows.Next() {
		var filePath string
		var modTime int64
		var contentHash string

		if err := rows.Scan(&filePath, &modTime, &contentHash); err != nil {
			return nil, err
		}

		filesData[filePath] = FileMetadata{
			FilePath:    filePath,
			ModTime:     modTime,
			ContentHash: contentHash,
		}
	}

	return filesData, rows.Err()
}

// parseMarkdownFileWithMetadata parses markdown with file metadata
func (s *SQLiteService) parseMarkdownFileWithMetadata(filename string, content []byte, modTime int64, contentHash string) (*Post, error) {
	// Parse frontmatter and content
	frontmatter, markdownContent, err := s.parseFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// Convert markdown to HTML
	htmlContent := s.markdownToHTML(markdownContent)

	// Generate slug from filename
	baseName := filepath.Base(filename)
	slug := strings.TrimSuffix(baseName, ".md")

	// Calculate reading time if not provided
	readingTime := frontmatter.ReadingTime
	if readingTime == 0 {
		readingTime, _ = s.calculateReadingTime(string(markdownContent))
	}

	// Convert summary markdown to HTML
	summaryHTML := ""
	if frontmatter.Summary != "" {
		summaryHTML = s.markdownToHTML([]byte(frontmatter.Summary))
	}

	return &Post{
		Slug:          slug,
		Title:         frontmatter.Title,
		Date:          frontmatter.Date,
		Summary:       frontmatter.Summary,
		SummaryHTML:   template.HTML(summaryHTML),
		Content:       template.HTML(htmlContent),
		ReadingTime:   readingTime,
		Tags:          frontmatter.Tags,
		FileName:      filename,
		FeaturedImage: frontmatter.FeaturedImage,
		OGDescription: frontmatter.OGDescription,
		TwitterHandle: frontmatter.TwitterHandle,
		FileModTime:   modTime,
		ContentHash:   contentHash,
	}, nil
}

// batchUpdatePosts updates existing posts preserving created_at timestamps
func (s *SQLiteService) batchUpdatePosts(ctx context.Context, posts []Post) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		UPDATE posts SET 
			title = ?, content = ?, summary = ?, published_at = ?, tags = ?, 
			category = ?, word_count = ?, reading_time = ?, featured_image = ?, 
			og_description = ?, twitter_handle = ?, file_mod_time = ?, content_hash = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE slug = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, post := range posts {
		tagsJSON, _ := json.Marshal(post.Tags)
		_, err = stmt.ExecContext(ctx,
			post.Title, string(post.Content), post.Summary, post.Date, string(tagsJSON),
			post.Category, post.WordCount, post.ReadingTime, post.FeaturedImage,
			post.OGDescription, post.TwitterHandle, post.FileModTime, post.ContentHash,
			post.Slug,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// removePostsByFilePath removes posts that no longer exist in filesystem
func (s *SQLiteService) removePostsByFilePath(ctx context.Context, filePaths []string) error {
	if len(filePaths) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create placeholders for IN clause
	placeholders := strings.Repeat("?,", len(filePaths))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf("DELETE FROM posts WHERE file_path IN (%s)", placeholders)

	args := make([]interface{}, len(filePaths))
	for i, path := range filePaths {
		args[i] = path
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Path History Service Implementation

// AddPath adds a new path for a post identified by content hash
func (s *SQLiteService) AddPath(ctx context.Context, contentHash, path string, isCurrent bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabase, "failed to begin transaction")
	}
	defer tx.Rollback()

	// If this is being set as current, mark all other paths for this content as non-current
	if isCurrent {
		_, err = tx.ExecContext(ctx, `
			UPDATE post_paths SET is_current = FALSE 
			WHERE content_hash = ?`, contentHash)
		if err != nil {
			return errors.Wrap(err, errors.ErrCodeDatabase, "failed to update existing paths to non-current")
		}
	}

	// Insert or update the path
	_, err = tx.ExecContext(ctx, `
		INSERT INTO post_paths (content_hash, path, is_current) 
		VALUES (?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET 
			is_current = excluded.is_current,
			content_hash = excluded.content_hash`,
		contentHash, path, isCurrent)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabase, "failed to insert path")
	}

	return tx.Commit()
}

// GetCurrentPath returns the current canonical path for a post
func (s *SQLiteService) GetCurrentPath(ctx context.Context, contentHash string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	var path string
	err := s.db.QueryRowContext(ctx, `
		SELECT path FROM post_paths 
		WHERE content_hash = ? AND is_current = TRUE
		LIMIT 1`, contentHash).Scan(&path)

	if err == sql.ErrNoRows {
		return "", errors.New(errors.ErrCodeNotFound, "no current path found for content hash")
	}
	if err != nil {
		return "", errors.Wrap(err, errors.ErrCodeDatabase, "failed to query current path")
	}

	return path, nil
}

// GetAllPaths returns all historical paths for a post
func (s *SQLiteService) GetAllPaths(ctx context.Context, contentHash string) ([]PostPath, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, content_hash, path, is_current, created_at 
		FROM post_paths 
		WHERE content_hash = ? 
		ORDER BY created_at DESC`, contentHash)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabase, "failed to query paths")
	}
	defer rows.Close()

	var paths []PostPath
	for rows.Next() {
		var path PostPath
		err := rows.Scan(&path.ID, &path.ContentHash, &path.Path,
			&path.IsCurrent, &path.CreatedAt)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrCodeDatabase, "failed to scan path row")
		}
		paths = append(paths, path)
	}

	return paths, rows.Err()
}

// FindPostByPath returns the content hash for a post given any of its historical paths
func (s *SQLiteService) FindPostByPath(ctx context.Context, path string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	var contentHash string
	err := s.db.QueryRowContext(ctx, `
		SELECT content_hash FROM post_paths 
		WHERE path = ?
		LIMIT 1`, path).Scan(&contentHash)

	if err == sql.ErrNoRows {
		return "", errors.New(errors.ErrCodeNotFound, "no post found for path")
	}
	if err != nil {
		return "", errors.Wrap(err, errors.ErrCodeDatabase, "failed to query post by path")
	}

	return contentHash, nil
}

// UpdateCurrentPath marks a new path as current and old paths as historical
func (s *SQLiteService) UpdateCurrentPath(ctx context.Context, contentHash, newPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabase, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Mark all existing paths as non-current
	_, err = tx.ExecContext(ctx, `
		UPDATE post_paths SET is_current = FALSE 
		WHERE content_hash = ?`, contentHash)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabase, "failed to update existing paths")
	}

	// Add or update the new current path
	_, err = tx.ExecContext(ctx, `
		INSERT INTO post_paths (content_hash, path, is_current) 
		VALUES (?, ?, TRUE)
		ON CONFLICT(path) DO UPDATE SET 
			is_current = TRUE,
			content_hash = excluded.content_hash`,
		contentHash, newPath)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabase, "failed to set new current path")
	}

	return tx.Commit()
}

// CleanupOrphanedPaths removes paths for posts that no longer exist
func (s *SQLiteService) CleanupOrphanedPaths(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_, err := s.db.ExecContext(ctx, `
		DELETE FROM post_paths 
		WHERE content_hash NOT IN (
			SELECT DISTINCT content_hash FROM posts
		)`)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabase, "failed to cleanup orphaned paths")
	}

	return nil
}

// updatePathHistory updates the path history for a list of posts
func (s *SQLiteService) updatePathHistory(ctx context.Context, posts []Post) error {
	for _, post := range posts {
		// Generate blog path from post slug
		blogPath := fmt.Sprintf("blog/%s", post.Slug)

		// Update the current path for this post
		if err := s.UpdateCurrentPath(ctx, post.ContentHash, blogPath); err != nil {
			return fmt.Errorf("failed to update path history for post %s: %w", post.Slug, err)
		}
	}
	return nil
}

// generateBlogPath creates a blog URL path from a post slug
func (s *SQLiteService) generateBlogPath(slug string) string {
	return fmt.Sprintf("blog/%s", slug)
}
