package server

import (
	"context"
	"net/http"
	"strings"

	"blockhead.consulting/internal/blog"
)

// PathRedirectMiddleware handles redirects for old blog post paths
func PathRedirectMiddleware(blogService blog.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only process blog-related requests
			if !strings.HasPrefix(r.URL.Path, "/blog/") {
				next.ServeHTTP(w, r)
				return
			}

			// Check if this is a path that needs redirection
			if pathHistoryService, ok := blogService.(blog.PathHistoryService); ok {
				if err := handlePathRedirect(w, r, pathHistoryService); err != nil {
					// If there's an error or no redirect needed, continue with normal flow
					next.ServeHTTP(w, r)
					return
				}
				// If redirect was handled, return without calling next
				return
			}

			// If no path history service available, continue normally
			next.ServeHTTP(w, r)
		})
	}
}

// handlePathRedirect checks if a path needs redirection and handles it
func handlePathRedirect(w http.ResponseWriter, r *http.Request, pathService blog.PathHistoryService) error {
	ctx := r.Context()

	// Extract the path from the URL (remove leading slash)
	requestPath := strings.TrimPrefix(r.URL.Path, "/")

	// Try to find a post by this exact path
	contentHash, err := pathService.FindPostByPath(ctx, requestPath)
	if err != nil {
		// No post found for this path, let normal handling continue
		return err
	}

	// Get the current canonical path for this content
	currentPath, err := pathService.GetCurrentPath(ctx, contentHash)
	if err != nil {
		// Can't get current path, let normal handling continue
		return err
	}

	// If the requested path is already the current path, no redirect needed
	if requestPath == currentPath {
		return nil
	}

	// Redirect to the current canonical path
	redirectURL := "/" + currentPath

	// Preserve query parameters
	if r.URL.RawQuery != "" {
		redirectURL += "?" + r.URL.RawQuery
	}

	// Use 301 (Moved Permanently) for SEO benefits
	http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)

	return nil
}

// GetPostByAnyPath attempts to find a post by any of its historical paths
func GetPostByAnyPath(ctx context.Context, pathService blog.PathHistoryService, blogService blog.Service, path string) (*blog.Post, error) {
	// Try to find content hash by path
	contentHash, err := pathService.FindPostByPath(ctx, path)
	if err != nil {
		return nil, err
	}

	// Extract slug from current path to get the post
	currentPath, err := pathService.GetCurrentPath(ctx, contentHash)
	if err != nil {
		return nil, err
	}

	// Extract slug from path (assuming format "blog/slug")
	parts := strings.Split(currentPath, "/")
	if len(parts) < 2 {
		return nil, err
	}
	slug := parts[len(parts)-1]

	// Get the post by its current slug
	return blogService.GetBySlug(ctx, slug)
}
