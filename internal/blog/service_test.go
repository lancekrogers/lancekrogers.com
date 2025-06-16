package blog

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"
	"testing"
	"time"

	"blockhead.consulting/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/*
var testFS embed.FS

type mockEventBus struct {
	publishedEvents []events.Event
}

func (m *mockEventBus) Subscribe(eventType events.EventType, handler events.EventHandler) string {
	return "mock-subscription-id"
}

func (m *mockEventBus) Unsubscribe(subscriptionID string) {
	// No-op
}

func (m *mockEventBus) Publish(ctx context.Context, event events.Event) error {
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *mockEventBus) Start(ctx context.Context) error {
	return nil
}

func (m *mockEventBus) Stop() error {
	return nil
}

func createTestService(t *testing.T) (*service, *mockEventBus) {
	logger := log.New(os.Stdout, "[blog-test] ", log.LstdFlags)
	mockBus := &mockEventBus{}
	
	// Get the testdata subdirectory
	testDataFS, err := fs.Sub(testFS, "testdata")
	require.NoError(t, err)
	
	// Use "." as the blog directory since testDataFS is already the blog directory
	svc := NewServiceWithOptions(testDataFS, ".", logger, mockBus)
	
	return svc, mockBus
}

func TestNewService(t *testing.T) {
	logger := log.New(os.Stdout, "[test] ", log.LstdFlags)
	mockBus := &mockEventBus{}
	
	svc := NewService(testFS, logger, mockBus)
	assert.NotNil(t, svc)
	assert.Equal(t, "blog", svc.Name())
}

func TestServiceLifecycle(t *testing.T) {
	svc, mockBus := createTestService(t)
	ctx := context.Background()
	
	// Test Start
	err := svc.Start(ctx)
	assert.NoError(t, err)
	assert.Len(t, svc.posts, 2) // We have 2 test posts
	assert.Len(t, mockBus.publishedEvents, 1) // Should publish BlogInitialized event
	
	// Test Health
	err = svc.Health(ctx)
	assert.NoError(t, err)
	
	// Test Stop
	err = svc.Stop(ctx)
	assert.NoError(t, err)
}

func TestGetAll(t *testing.T) {
	svc, _ := createTestService(t)
	ctx := context.Background()
	
	err := svc.Start(ctx)
	require.NoError(t, err)
	
	posts := svc.GetAll(ctx)
	assert.Len(t, posts, 2)
	
	// Posts should be sorted by date (newest first)
	assert.Equal(t, "second-post", posts[0].Slug)
	assert.Equal(t, "first-post", posts[1].Slug)
}

func TestGetBySlug(t *testing.T) {
	svc, mockBus := createTestService(t)
	ctx := context.Background()
	
	err := svc.Start(ctx)
	require.NoError(t, err)
	
	// Clear initialization events
	mockBus.publishedEvents = []events.Event{}
	
	// Test existing post
	post, err := svc.GetBySlug(ctx, "first-post")
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "First Test Post", post.Title)
	// Author field removed from Post struct
	assert.Contains(t, post.Content, "<h1 id=\"first-test-post\">First Test Post</h1>")
	
	// Event publishing will be added later
	
	// Test non-existing post
	post, err = svc.GetBySlug(ctx, "non-existent")
	assert.Error(t, err)
	assert.Nil(t, post)
}

func TestSearch(t *testing.T) {
	svc, _ := createTestService(t)
	ctx := context.Background()
	
	err := svc.Start(ctx)
	require.NoError(t, err)
	
	// Search in title
	results := svc.Search(ctx, "First")
	assert.Len(t, results, 1)
	assert.Equal(t, "first-post", results[0].Slug)
	
	// Search in content
	results = svc.Search(ctx, "blockchain")
	assert.Len(t, results, 1)
	assert.Equal(t, "first-post", results[0].Slug)
	
	// Search in description
	results = svc.Search(ctx, "AI systems")
	assert.Len(t, results, 1)
	assert.Equal(t, "second-post", results[0].Slug)
	
	// Search with no results
	results = svc.Search(ctx, "nonexistent")
	assert.Len(t, results, 0)
	
	// Case insensitive search
	results = svc.Search(ctx, "BLOCKCHAIN")
	assert.Len(t, results, 1)
}

func TestGetByTag(t *testing.T) {
	svc, _ := createTestService(t)
	ctx := context.Background()
	
	err := svc.Start(ctx)
	require.NoError(t, err)
	
	// Get posts by existing tag
	posts := svc.GetByTag(ctx, "blockchain")
	assert.Len(t, posts, 1)
	assert.Equal(t, "first-post", posts[0].Slug) // first-post has blockchain tag
	
	// Get posts by another tag
	posts = svc.GetByTag(ctx, "ai")
	assert.Len(t, posts, 1)
	assert.Equal(t, "second-post", posts[0].Slug) // second-post has ai tag
	
	// Get posts by non-existing tag
	posts = svc.GetByTag(ctx, "nonexistent")
	assert.Len(t, posts, 0)
}

func TestGetTags(t *testing.T) {
	svc, _ := createTestService(t)
	ctx := context.Background()
	
	err := svc.Start(ctx)
	require.NoError(t, err)
	
	tags := svc.GetTags(ctx)
	assert.Len(t, tags, 4) // blockchain, consulting, ai, llm
	
	// Check specific tags exist
	assert.Contains(t, tags, "blockchain")
	assert.Contains(t, tags, "consulting")
	assert.Contains(t, tags, "ai")
	assert.Contains(t, tags, "llm")
}

func TestLoadPosts(t *testing.T) {
	svc, _ := createTestService(t)
	
	// Test initial load
	ctx := context.Background()
	err := svc.LoadPosts(ctx)
	assert.NoError(t, err)
	assert.Len(t, svc.posts, 2)
	
	// Verify post details
	firstPost := svc.postMap["first-post"]
	assert.NotNil(t, firstPost)
	assert.Equal(t, "First Test Post", firstPost.Title)
	assert.Equal(t, []string{"blockchain", "consulting"}, firstPost.Tags)
	assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), firstPost.Date)
	
	secondPost := svc.postMap["second-post"]
	assert.NotNil(t, secondPost)
	assert.Equal(t, "Second Test Post", secondPost.Title)
	assert.Equal(t, []string{"ai", "llm"}, secondPost.Tags)
	assert.Equal(t, time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), secondPost.Date)
}

func TestEmptyBlogFS(t *testing.T) {
	logger := log.New(os.Stdout, "[test] ", log.LstdFlags)
	mockBus := &mockEventBus{}
	
	// Create an empty FS
	emptyFS := embed.FS{}
	
	svc := &service{
		posts:    []Post{},
		postMap:  make(map[string]*Post),
		tagIndex: make(map[string][]int),
		blogFS:   emptyFS,
		blogDir:  ".",
		logger:   logger,
		eventBus: mockBus,
	}
	
	ctx := context.Background()
	err := svc.Start(ctx)
	assert.NoError(t, err)
	assert.Len(t, svc.posts, 0)
}

func TestInvalidMarkdownFile(t *testing.T) {
	// This test would require a specific test file with invalid frontmatter
	// For now, we'll skip this as it would require modifying the testdata
	t.Skip("Requires specific invalid markdown test file")
}