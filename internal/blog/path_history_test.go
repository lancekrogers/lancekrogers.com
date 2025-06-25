package blog

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestPathHistoryService(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a test blog service with path history
	service, err := NewSQLiteService(":memory:", nil, "", log.Default(), nil)
	if err != nil {
		t.Fatalf("Failed to create SQLite service: %v", err)
	}

	// Disable foreign key constraints for testing
	_, err = service.db.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		t.Fatalf("Failed to disable foreign keys: %v", err)
	}

	ctx := context.Background()

	// Test content hashes for test posts
	testHash1 := "abc123def456"
	testHash2 := "def456ghi789"

	// Create test posts in the database first to satisfy foreign key constraints
	testPosts := []Post{
		{
			Slug:        "test-post",
			Title:       "Test Post",
			Content:     "Content for test post",
			Summary:     "Summary",
			Date:        time.Now(),
			Tags:        []string{"test"},
			Category:    "test",
			FileName:    "test.md",
			WordCount:   5,
			ReadingTime: 1,
			ContentHash: testHash1,
			FileModTime: time.Now().Unix(),
		},
		{
			Slug:        "another-post",
			Title:       "Another Post",
			Content:     "Content for another post",
			Summary:     "Another Summary",
			Date:        time.Now(),
			Tags:        []string{"test"},
			Category:    "test",
			FileName:    "another.md",
			WordCount:   6,
			ReadingTime: 1,
			ContentHash: testHash2,
			FileModTime: time.Now().Unix(),
		},
	}

	// Insert the test posts
	err = service.batchInsertPosts(ctx, testPosts)
	if err != nil {
		t.Fatalf("Failed to insert test posts: %v", err)
	}

	t.Run("AddPath", func(t *testing.T) {
		// Add a current path for first post
		err := service.AddPath(ctx, testHash1, "blog/test-post", true)
		if err != nil {
			t.Errorf("Failed to add path: %v", err)
		}

		// Add a historical path for the same post
		err = service.AddPath(ctx, testHash1, "blog/old-test-post", false)
		if err != nil {
			t.Errorf("Failed to add historical path: %v", err)
		}

		// Add current path for second post
		err = service.AddPath(ctx, testHash2, "blog/another-post", true)
		if err != nil {
			t.Errorf("Failed to add path for second post: %v", err)
		}
	})

	t.Run("GetCurrentPath", func(t *testing.T) {
		// Get current path for first post
		currentPath, err := service.GetCurrentPath(ctx, testHash1)
		if err != nil {
			t.Errorf("Failed to get current path: %v", err)
		}
		if currentPath != "blog/test-post" {
			t.Errorf("Expected current path 'blog/test-post', got '%s'", currentPath)
		}

		// Get current path for second post
		currentPath2, err := service.GetCurrentPath(ctx, testHash2)
		if err != nil {
			t.Errorf("Failed to get current path for second post: %v", err)
		}
		if currentPath2 != "blog/another-post" {
			t.Errorf("Expected current path 'blog/another-post', got '%s'", currentPath2)
		}
	})

	t.Run("GetAllPaths", func(t *testing.T) {
		// Get all paths for first post
		paths, err := service.GetAllPaths(ctx, testHash1)
		if err != nil {
			t.Errorf("Failed to get all paths: %v", err)
		}
		if len(paths) != 2 {
			t.Errorf("Expected 2 paths, got %d", len(paths))
		}

		// Check that one is current and one is not
		var hasCurrent, hasHistorical bool
		for _, path := range paths {
			if path.Path == "blog/test-post" && path.IsCurrent {
				hasCurrent = true
			}
			if path.Path == "blog/old-test-post" && !path.IsCurrent {
				hasHistorical = true
			}
		}
		if !hasCurrent {
			t.Error("Should have a current path for 'blog/test-post'")
		}
		if !hasHistorical {
			t.Error("Should have a historical path for 'blog/old-test-post'")
		}
	})

	t.Run("FindPostByPath", func(t *testing.T) {
		// Find post by current path
		hash, err := service.FindPostByPath(ctx, "blog/test-post")
		if err != nil {
			t.Errorf("Failed to find post by current path: %v", err)
		}
		if hash != testHash1 {
			t.Errorf("Expected hash '%s', got '%s'", testHash1, hash)
		}

		// Find post by historical path
		hash2, err := service.FindPostByPath(ctx, "blog/old-test-post")
		if err != nil {
			t.Errorf("Failed to find post by historical path: %v", err)
		}
		if hash2 != testHash1 {
			t.Errorf("Expected hash '%s', got '%s'", testHash1, hash2)
		}

		// Find second post
		hash3, err := service.FindPostByPath(ctx, "blog/another-post")
		if err != nil {
			t.Errorf("Failed to find second post: %v", err)
		}
		if hash3 != testHash2 {
			t.Errorf("Expected hash '%s', got '%s'", testHash2, hash3)
		}
	})

	t.Run("UpdateCurrentPath", func(t *testing.T) {
		// Update current path for first post
		err := service.UpdateCurrentPath(ctx, testHash1, "blog/renamed-test-post")
		if err != nil {
			t.Errorf("Failed to update current path: %v", err)
		}

		// Verify old current path is now historical
		paths, err := service.GetAllPaths(ctx, testHash1)
		if err != nil {
			t.Errorf("Failed to get paths after update: %v", err)
		}
		if len(paths) != 3 {
			t.Errorf("Expected 3 paths after update, got %d", len(paths))
		}

		// Check current path is updated
		currentPath, err := service.GetCurrentPath(ctx, testHash1)
		if err != nil {
			t.Errorf("Failed to get updated current path: %v", err)
		}
		if currentPath != "blog/renamed-test-post" {
			t.Errorf("Expected updated current path 'blog/renamed-test-post', got '%s'", currentPath)
		}

		// Verify old paths are historical
		for _, path := range paths {
			if path.Path == "blog/test-post" && path.IsCurrent {
				t.Error("Old current path should now be historical")
			}
			if path.Path == "blog/old-test-post" && path.IsCurrent {
				t.Error("Historical path should remain historical")
			}
			if path.Path == "blog/renamed-test-post" && !path.IsCurrent {
				t.Error("New path should be current")
			}
		}
	})

	t.Run("FindPostByPath_NotFound", func(t *testing.T) {
		// Try to find non-existent path
		_, err := service.FindPostByPath(ctx, "blog/non-existent")
		if err == nil {
			t.Error("Expected error for non-existent path")
		}
	})

	t.Run("GetCurrentPath_NotFound", func(t *testing.T) {
		// Try to get current path for non-existent hash
		_, err := service.GetCurrentPath(ctx, "nonexistenthash")
		if err == nil {
			t.Error("Expected error for non-existent content hash")
		}
	})

	t.Run("CleanupOrphanedPaths", func(t *testing.T) {
		// First create an orphaned post and then delete it from posts table to simulate orphaned paths
		orphanedHash := "orphaned123"
		orphanedPost := Post{
			Slug:        "orphaned-post",
			Title:       "Orphaned Post",
			Content:     "Content for orphaned post",
			Summary:     "Orphaned Summary",
			Date:        time.Now(),
			Tags:        []string{"test"},
			Category:    "test",
			FileName:    "orphaned.md",
			WordCount:   7,
			ReadingTime: 1,
			ContentHash: orphanedHash,
			FileModTime: time.Now().Unix(),
		}

		// Insert the orphaned post
		err := service.batchInsertPosts(ctx, []Post{orphanedPost})
		if err != nil {
			t.Errorf("Failed to insert orphaned post: %v", err)
		}

		// Add a path for the orphaned post
		err = service.AddPath(ctx, orphanedHash, "blog/orphaned-post", true)
		if err != nil {
			t.Errorf("Failed to add orphaned path: %v", err)
		}

		// Verify it can be found before cleanup
		_, err = service.FindPostByPath(ctx, "blog/orphaned-post")
		if err != nil {
			t.Errorf("Orphaned path should exist before cleanup: %v", err)
		}

		// Delete the post from posts table to make the path orphaned
		_, err = service.db.ExecContext(ctx, "DELETE FROM posts WHERE content_hash = ?", orphanedHash)
		if err != nil {
			t.Errorf("Failed to delete orphaned post: %v", err)
		}

		// Run cleanup
		err = service.CleanupOrphanedPaths(ctx)
		if err != nil {
			t.Errorf("Failed to cleanup orphaned paths: %v", err)
		}

		// Verify orphaned path is removed
		_, err = service.FindPostByPath(ctx, "blog/orphaned-post")
		if err == nil {
			t.Error("Orphaned path should be removed after cleanup")
		}

		// Verify existing paths are still there
		_, err = service.FindPostByPath(ctx, "blog/renamed-test-post")
		if err != nil {
			t.Errorf("Existing path should not be removed: %v", err)
		}
	})
}

func TestPathHistoryIntegration(t *testing.T) {
	// Create an in-memory SQLite database for testing
	service, err := NewSQLiteService(":memory:", nil, "", log.Default(), nil)
	if err != nil {
		t.Fatalf("Failed to create SQLite service: %v", err)
	}

	// Disable foreign key constraints for testing
	_, err = service.db.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		t.Fatalf("Failed to disable foreign keys: %v", err)
	}

	ctx := context.Background()

	// Create test posts directly in database
	testPosts := []Post{
		{
			Slug:        "test-post-1",
			Title:       "Test Post 1",
			Content:     "Content for test post 1",
			Summary:     "Summary 1",
			Date:        time.Now(),
			Tags:        []string{"test", "blog"},
			Category:    "ai",
			FileName:    "content/blog/ai/test-post-1.md",
			WordCount:   10,
			ReadingTime: 1,
			ContentHash: "hash1",
			FileModTime: time.Now().Unix(),
		},
		{
			Slug:        "test-post-2",
			Title:       "Test Post 2",
			Content:     "Content for test post 2",
			Summary:     "Summary 2",
			Date:        time.Now(),
			Tags:        []string{"test", "crypto"},
			Category:    "crypto",
			FileName:    "content/blog/crypto/test-post-2.md",
			WordCount:   15,
			ReadingTime: 1,
			ContentHash: "hash2",
			FileModTime: time.Now().Unix(),
		},
	}

	t.Run("IntegrationTest_PostsAndPaths", func(t *testing.T) {
		// Insert test posts
		err := service.batchInsertPosts(ctx, testPosts)
		if err != nil {
			t.Fatalf("Failed to insert test posts: %v", err)
		}

		// Update path history for posts
		err = service.updatePathHistory(ctx, testPosts)
		if err != nil {
			t.Fatalf("Failed to update path history: %v", err)
		}

		// Verify paths were created correctly
		for _, post := range testPosts {
			expectedPath := service.generateBlogPath(post.Slug)

			// Check current path
			currentPath, err := service.GetCurrentPath(ctx, post.ContentHash)
			if err != nil {
				t.Errorf("Failed to get current path for %s: %v", post.Slug, err)
				continue
			}
			if currentPath != expectedPath {
				t.Errorf("Expected path '%s', got '%s'", expectedPath, currentPath)
			}

			// Check we can find post by path
			foundHash, err := service.FindPostByPath(ctx, expectedPath)
			if err != nil {
				t.Errorf("Failed to find post by path %s: %v", expectedPath, err)
				continue
			}
			if foundHash != post.ContentHash {
				t.Errorf("Expected hash '%s', got '%s'", post.ContentHash, foundHash)
			}
		}

		// Test path update scenario (simulating post move)
		newPath := "blog/moved-test-post-1"
		err = service.UpdateCurrentPath(ctx, testPosts[0].ContentHash, newPath)
		if err != nil {
			t.Errorf("Failed to update path: %v", err)
		}

		// Verify old path still works (historical redirect)
		oldPath := service.generateBlogPath(testPosts[0].Slug)
		hash, err := service.FindPostByPath(ctx, oldPath)
		if err != nil {
			t.Errorf("Historical path should still work: %v", err)
		}
		if hash != testPosts[0].ContentHash {
			t.Errorf("Historical path should return same content hash")
		}

		// Verify new path is current
		currentPath, err := service.GetCurrentPath(ctx, testPosts[0].ContentHash)
		if err != nil {
			t.Errorf("Failed to get updated current path: %v", err)
		}
		if currentPath != newPath {
			t.Errorf("Expected updated path '%s', got '%s'", newPath, currentPath)
		}
	})
}

func TestGenerateBlogPath(t *testing.T) {
	service, err := NewSQLiteService(":memory:", nil, "", log.Default(), nil)
	if err != nil {
		t.Fatalf("Failed to create SQLite service: %v", err)
	}

	// Disable foreign key constraints for testing
	_, err = service.db.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		t.Fatalf("Failed to disable foreign keys: %v", err)
	}

	tests := []struct {
		slug     string
		expected string
	}{
		{"test-post", "blog/test-post"},
		{"another-post", "blog/another-post"},
		{"post-with-numbers-123", "blog/post-with-numbers-123"},
		{"", "blog/"},
	}

	for _, test := range tests {
		result := service.generateBlogPath(test.slug)
		if result != test.expected {
			t.Errorf("generateBlogPath(%s) = %s, expected %s", test.slug, result, test.expected)
		}
	}
}
