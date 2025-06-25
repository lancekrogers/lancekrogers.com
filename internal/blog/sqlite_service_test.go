package blog

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"blockhead.consulting/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFileSystem provides test markdown files
type TestFileSystem struct {
	files map[string][]byte
}

func (tfs *TestFileSystem) Open(name string) (fs.File, error) {
	// Check if it's a file
	if content, exists := tfs.files[name]; exists {
		return &TestFile{name: name, content: content}, nil
	}

	// Check if it's a directory by seeing if any files start with this path
	dirName := strings.TrimSuffix(name, "/")
	for path := range tfs.files {
		if strings.HasPrefix(path, dirName+"/") || path == dirName {
			return &TestDir{name: name, fs: tfs}, nil
		}
	}

	return nil, fs.ErrNotExist
}

func (tfs *TestFileSystem) ReadFile(name string) ([]byte, error) {
	if content, exists := tfs.files[name]; exists {
		return content, nil
	}
	return nil, fs.ErrNotExist
}

func (tfs *TestFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry
	dirs := make(map[string]bool)

	// Normalize the directory name
	dirName := strings.TrimSuffix(name, "/")

	for path := range tfs.files {
		if strings.HasPrefix(path, dirName) {
			// Handle direct files in this directory
			if filepath.Dir(path) == dirName {
				entries = append(entries, &TestDirEntry{
					name:  filepath.Base(path),
					isDir: false,
				})
			}

			// Handle subdirectories
			rel, err := filepath.Rel(dirName, path)
			if err == nil && rel != "." && !strings.Contains(rel, "..") {
				parts := strings.Split(rel, string(filepath.Separator))
				if len(parts) > 1 && !dirs[parts[0]] {
					dirs[parts[0]] = true
					entries = append(entries, &TestDirEntry{
						name:  parts[0],
						isDir: true,
					})
				}
			}
		}
	}

	if len(entries) == 0 {
		return nil, fs.ErrNotExist
	}

	return entries, nil
}

func (tfs *TestFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	// Normalize root path
	root = strings.TrimSuffix(root, "/")
	if root == "" {
		root = "."
	}

	// Call walkdir function for the root directory first
	err := fn(root, &TestDirEntry{
		name:  filepath.Base(root),
		isDir: true,
	}, nil)
	if err != nil {
		return err
	}

	// Process all files that match the root path
	for path, content := range tfs.files {
		// Check if this file is under the root path
		var matchesRoot bool
		if root == "." {
			matchesRoot = true
		} else {
			matchesRoot = strings.HasPrefix(path, root+"/") || path == root
		}

		if matchesRoot && strings.HasSuffix(path, ".md") && len(content) > 0 {
			err := fn(path, &TestDirEntry{
				name:  filepath.Base(path),
				isDir: false,
			}, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type TestDirEntry struct {
	name  string
	isDir bool
}

func (tde *TestDirEntry) Name() string               { return tde.name }
func (tde *TestDirEntry) IsDir() bool                { return tde.isDir }
func (tde *TestDirEntry) Type() fs.FileMode          { return 0 }
func (tde *TestDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// TestFile implements fs.File for test files
type TestFile struct {
	name    string
	content []byte
	pos     int64
}

func (tf *TestFile) Stat() (fs.FileInfo, error) {
	return &TestFileInfo{name: tf.name, size: int64(len(tf.content)), isDir: false}, nil
}

func (tf *TestFile) Read(b []byte) (int, error) {
	if tf.pos >= int64(len(tf.content)) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(b, tf.content[tf.pos:])
	tf.pos += int64(n)
	return n, nil
}

func (tf *TestFile) Close() error {
	return nil
}

// TestDir implements fs.File for test directories
type TestDir struct {
	name string
	fs   *TestFileSystem
}

func (td *TestDir) Stat() (fs.FileInfo, error) {
	return &TestFileInfo{name: td.name, isDir: true}, nil
}

func (td *TestDir) Read(b []byte) (int, error) {
	return 0, fmt.Errorf("is a directory")
}

func (td *TestDir) Close() error {
	return nil
}

// TestFileInfo implements fs.FileInfo
type TestFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (tfi *TestFileInfo) Name() string { return tfi.name }
func (tfi *TestFileInfo) Size() int64  { return tfi.size }
func (tfi *TestFileInfo) Mode() fs.FileMode {
	if tfi.isDir {
		return fs.ModeDir
	}
	return 0
}
func (tfi *TestFileInfo) ModTime() time.Time { return time.Now() }
func (tfi *TestFileInfo) IsDir() bool        { return tfi.isDir }
func (tfi *TestFileInfo) Sys() interface{}   { return nil }

func setupTestSQLiteService(t *testing.T) (*SQLiteService, func()) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create test filesystem with sample posts
	testFS := &TestFileSystem{
		files: map[string][]byte{
			"content/blog/test-post-1.md": []byte(`---
title: "Test Post 1"
date: 2025-01-01T00:00:00Z
summary: "This is a test post about Go"
tags: ["golang", "testing"]
---

# Test Post 1

This is the content of test post 1. It discusses Go programming and testing strategies.

## Key Points

- Go is fast
- Testing is important
- SQLite is awesome
`),
			"content/blog/test-post-2.md": []byte(`---
title: "Test Post 2"  
date: 2025-01-02T00:00:00Z
summary: "This is a test post about blockchain"
tags: ["blockchain", "crypto"]
---

# Test Post 2

This is the content of test post 2. It covers blockchain technology and cryptocurrency.

## Topics Covered

- Bitcoin basics
- Smart contracts
- DeFi protocols
`),
			"content/blog/blockchain/advanced-post.md": []byte(`---
title: "Advanced Blockchain Post"
date: 2025-01-03T00:00:00Z
summary: "Advanced blockchain concepts"
tags: ["blockchain", "advanced", "defi"]
---

# Advanced Blockchain Concepts

This post covers advanced blockchain topics for experienced developers.

## Advanced Topics

- Zero-knowledge proofs
- Layer 2 scaling
- Cross-chain bridges
`),
		},
	}

	// Create logger that doesn't output during tests
	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)

	// Create event bus
	eventBus := events.NewInMemoryEventBus(5, logger)

	// Create service
	service, err := NewSQLiteService(dbPath, testFS, "content/blog", logger, eventBus)
	require.NoError(t, err)

	// Start service to import test data
	ctx := context.Background()
	err = service.Start(ctx)
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		service.Stop(ctx)
		os.RemoveAll(tmpDir)
	}

	return service, cleanup
}

func TestSQLiteService_ImportMarkdownFiles(t *testing.T) {
	service, cleanup := setupTestSQLiteService(t)
	defer cleanup()

	ctx := context.Background()

	// Re-import to test the process
	err := service.ImportMarkdownFiles(ctx)
	require.NoError(t, err)

	// Verify posts were imported
	posts := service.GetAll(ctx)
	assert.Len(t, posts, 3)

	// Check that posts are sorted by date (newest first)
	assert.True(t, posts[0].Date.After(posts[1].Date))
	assert.True(t, posts[1].Date.After(posts[2].Date))
}

func TestSQLiteService_GetBySlug(t *testing.T) {
	service, cleanup := setupTestSQLiteService(t)
	defer cleanup()

	ctx := context.Background()

	// Test getting existing post
	post, err := service.GetBySlug(ctx, "test-post-1")
	require.NoError(t, err)
	assert.Equal(t, "Test Post 1", post.Title)
	assert.Equal(t, "test-post-1", post.Slug)
	assert.Contains(t, post.Tags, "golang")
	assert.Contains(t, post.Tags, "testing")

	// Test getting non-existent post
	_, err = service.GetBySlug(ctx, "non-existent")
	assert.Error(t, err)
}

func TestSQLiteService_GetPaginatedPosts(t *testing.T) {
	service, cleanup := setupTestSQLiteService(t)
	defer cleanup()

	ctx := context.Background()

	// Test basic pagination
	params := PaginationParams{
		Page:    1,
		PerPage: 2,
	}

	result, err := service.GetPaginatedPosts(ctx, params)
	require.NoError(t, err)

	assert.Len(t, result.Posts, 2)
	assert.Equal(t, 3, result.TotalPosts)
	assert.Equal(t, 2, result.TotalPages)
	assert.Equal(t, 1, result.CurrentPage)
	assert.False(t, result.HasPrev)
	assert.True(t, result.HasNext)

	// Test second page
	params.Page = 2
	result, err = service.GetPaginatedPosts(ctx, params)
	require.NoError(t, err)

	assert.Len(t, result.Posts, 1)
	assert.True(t, result.HasPrev)
	assert.False(t, result.HasNext)
}

func TestSQLiteService_SearchPosts(t *testing.T) {
	service, cleanup := setupTestSQLiteService(t)
	defer cleanup()

	ctx := context.Background()

	// Test search for "Go"
	result, err := service.SearchPosts(ctx, "Go", 1, 10)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(result.Posts), 1)

	// Test search for "blockchain"
	result, err = service.SearchPosts(ctx, "blockchain", 1, 10)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(result.Posts), 2)

	// Test search with no results
	result, err = service.SearchPosts(ctx, "nonexistent", 1, 10)
	require.NoError(t, err)

	assert.Len(t, result.Posts, 0)
}

func TestSQLiteService_GetByTag(t *testing.T) {
	service, cleanup := setupTestSQLiteService(t)
	defer cleanup()

	ctx := context.Background()

	// Test getting posts by tag
	posts := service.GetByTag(ctx, "golang")
	if !assert.Len(t, posts, 1) {
		t.Logf("Expected 1 post with 'golang' tag, got %d posts", len(posts))
		return
	}
	assert.Equal(t, "Test Post 1", posts[0].Title)

	posts = service.GetByTag(ctx, "blockchain")
	if !assert.Len(t, posts, 2) {
		t.Logf("Expected 2 posts with 'blockchain' tag, got %d posts", len(posts))
		return
	}

	posts = service.GetByTag(ctx, "nonexistent")
	assert.Len(t, posts, 0)
}

func TestSQLiteService_GetTagCounts(t *testing.T) {
	service, cleanup := setupTestSQLiteService(t)
	defer cleanup()

	ctx := context.Background()

	tagCounts, err := service.GetTagCounts(ctx)
	require.NoError(t, err)

	// Should have multiple tags
	assert.Greater(t, len(tagCounts), 0)

	// Find blockchain tag (should appear in 2 posts)
	var blockchainCount int
	for _, tc := range tagCounts {
		if tc.Tag == "blockchain" {
			blockchainCount = tc.Count
			break
		}
	}
	assert.Equal(t, 2, blockchainCount)
}

func TestSQLiteService_GetPostNavigation(t *testing.T) {
	service, cleanup := setupTestSQLiteService(t)
	defer cleanup()

	ctx := context.Background()

	// Test navigation for middle post
	nav, err := service.GetPostNavigation(ctx, "test-post-2")
	require.NoError(t, err)

	assert.NotNil(t, nav.Current)
	assert.Equal(t, "test-post-2", nav.Current.Slug)

	// Should have both previous and next
	assert.NotNil(t, nav.Next)     // Older post
	assert.NotNil(t, nav.Previous) // Newer post
}

func TestSQLiteService_GetRelatedPosts(t *testing.T) {
	service, cleanup := setupTestSQLiteService(t)
	defer cleanup()

	ctx := context.Background()

	// Test related posts for blockchain post
	related, err := service.GetRelatedPosts(ctx, "test-post-2", 3)
	require.NoError(t, err)

	// Should find at least one related post (the other blockchain post)
	assert.Greater(t, len(related), 0)

	// Related posts should have relevance scores
	for _, rp := range related {
		assert.Greater(t, rp.Relevance, 0.0)
		assert.NotNil(t, rp.Post)
	}
}

func TestSQLiteService_CategoryExtraction(t *testing.T) {
	service, cleanup := setupTestSQLiteService(t)
	defer cleanup()

	// Test category extraction from path
	category := service.extractCategoryFromPath("content/blog/blockchain/advanced-post.md")
	assert.Equal(t, "blockchain", category)

	category = service.extractCategoryFromPath("content/blog/test-post.md")
	assert.Equal(t, "", category) // No subdirectory
}

func TestSQLiteService_ReadingTimeCalculation(t *testing.T) {
	service, cleanup := setupTestSQLiteService(t)
	defer cleanup()

	// Test reading time calculation
	readingTime, wordCount := service.calculateReadingTime("This is a test with exactly ten words in total.")
	assert.Equal(t, 1, readingTime) // Should be minimum 1 minute
	assert.Equal(t, 10, wordCount)

	// Test with longer text
	longText := ""
	for i := 0; i < 300; i++ {
		longText += "word "
	}

	readingTime, wordCount = service.calculateReadingTime(longText)
	assert.Equal(t, 2, readingTime) // 300 words / 225 WPM = ~1.33, rounded up to 2
	assert.Equal(t, 300, wordCount)
}

func TestSQLiteService_SearchQuerySanitization(t *testing.T) {
	service, cleanup := setupTestSQLiteService(t)
	defer cleanup()

	// Test search query sanitization
	sanitized := service.sanitizeSearchQuery(`test "query"`)
	assert.Equal(t, `"test ""query"""`, sanitized)

	sanitized = service.sanitizeSearchQuery("single word")
	assert.Equal(t, `"single word"`, sanitized)

	sanitized = service.sanitizeSearchQuery("singleword")
	assert.Equal(t, "singleword", sanitized)
}

func TestSQLiteService_PaginationWithFilters(t *testing.T) {
	service, cleanup := setupTestSQLiteService(t)
	defer cleanup()

	ctx := context.Background()

	// Test pagination with tag filter
	params := PaginationParams{
		Page:    1,
		PerPage: 10,
		Tag:     "blockchain",
	}

	result, err := service.GetPaginatedPosts(ctx, params)
	require.NoError(t, err)

	assert.Len(t, result.Posts, 2) // Should find 2 blockchain posts
	assert.Equal(t, 2, result.TotalPosts)

	// Test pagination with search
	params = PaginationParams{
		Page:    1,
		PerPage: 10,
		Search:  "Go",
	}

	result, err = service.GetPaginatedPosts(ctx, params)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(result.Posts), 1)
}

// Benchmark tests
func BenchmarkSQLiteService_GetPaginatedPosts(b *testing.B) {
	// Create a service with test data
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")

	testFS := &TestFileSystem{
		files: map[string][]byte{
			"content/blog/post.md": []byte(`---
title: "Benchmark Post"
date: 2025-01-01T00:00:00Z
summary: "Benchmark test post"
tags: ["benchmark", "test"]
---

# Benchmark Post

This is benchmark content.
`),
		},
	}

	logger := log.New(os.Stderr, "[bench] ", log.LstdFlags)
	eventBus := events.NewInMemoryEventBus(5, logger)

	service, err := NewSQLiteService(dbPath, testFS, "content/blog", logger, eventBus)
	if err != nil {
		b.Fatal(err)
	}
	defer service.Stop(context.Background())

	ctx := context.Background()
	service.Start(ctx)

	params := PaginationParams{
		Page:    1,
		PerPage: 12,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetPaginatedPosts(ctx, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSQLiteService_SearchPosts(b *testing.B) {
	// Similar setup as above benchmark
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")

	testFS := &TestFileSystem{
		files: map[string][]byte{
			"content/blog/post.md": []byte(`---
title: "Searchable Post"
date: 2025-01-01T00:00:00Z
summary: "Post for search benchmarks"
tags: ["search", "benchmark"]
---

# Searchable Content

This post contains searchable content for benchmarking FTS5 performance.
`),
		},
	}

	logger := log.New(os.Stderr, "[bench] ", log.LstdFlags)
	eventBus := events.NewInMemoryEventBus(5, logger)

	service, err := NewSQLiteService(dbPath, testFS, "content/blog", logger, eventBus)
	if err != nil {
		b.Fatal(err)
	}
	defer service.Stop(context.Background())

	ctx := context.Background()
	service.Start(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.SearchPosts(ctx, "searchable", 1, 12)
		if err != nil {
			b.Fatal(err)
		}
	}
}
