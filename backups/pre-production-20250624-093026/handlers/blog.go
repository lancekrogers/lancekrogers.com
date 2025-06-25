package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"blockhead.consulting/internal/blog"
	"blockhead.consulting/internal/config"
	"github.com/gorilla/mux"
)

// BlogHandler handles the main blog page
func (h *Handler) BlogHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Load services configuration for modal
	servicesConfig, err := h.ConfigService.LoadServicesConfig("")
	if err != nil {
		log.Printf("Warning: Failed to load services config for blog page: %v", err)
	}
	
	// Parse query parameters for pagination and filtering
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	category := r.URL.Query().Get("category")
	tag := r.URL.Query().Get("tag")
	search := r.URL.Query().Get("search")
	
	if page < 1 {
		page = 1
	}
	
	var posts []blog.Post
	var pagination *blog.PaginatedPosts
	
	// Use SQLite service if available for pagination
	if extendedService, ok := h.BlogService.(blog.ExtendedService); ok {
		params := blog.PaginationParams{
			Page:     page,
			PerPage:  12,
			Category: category,
			Tag:      tag,
			Search:   search,
		}
		
		result, err := extendedService.GetPaginatedPosts(ctx, params)
		if err != nil {
			log.Printf("Pagination error: %v", err)
			// Fallback to all posts
			posts = h.BlogService.GetAll(ctx)
		} else {
			posts = result.Posts
			pagination = result
		}
	} else {
		// Fallback to legacy behavior
		posts = h.BlogService.GetAll(ctx)
	}
	
	data := struct {
		Title          string
		Page           string
		Posts          []blog.Post
		Config         *SiteConfig
		SiteConfig     *config.SiteConfig
		AppConfig      *config.SiteConfig
		WorkConfig     *config.WorkConfig
		ServicesConfig *config.ServicesConfig
		BlogConfig     *blog.BlogConfig
		Pagination     *blog.PaginatedPosts
		Filters        map[string]string
	}{
		Title:          "Blog - Blockhead Consulting",
		Page:           "blog",
		Posts:          posts,
		Config:         h.SiteConfig,
		SiteConfig:     h.AppConfig,
		AppConfig:      h.AppConfig,
		WorkConfig:     h.WorkConfig,
		ServicesConfig: servicesConfig,
		BlogConfig:     h.BlogService.GetBlogConfig(),
		Pagination:     pagination,
		Filters: map[string]string{
			"category": category,
			"tag":      tag,
			"search":   search,
		},
	}

	if err := h.Templates.ExecuteTemplate(w, "page-blog.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// BlogPostHandler handles individual blog post pages
func (h *Handler) BlogPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	ctx := r.Context()
	
	// Load services configuration for modal
	servicesConfig, err := h.ConfigService.LoadServicesConfig("")
	if err != nil {
		log.Printf("Warning: Failed to load services config for blog post: %v", err)
	}
	
	// Debug: Check if blogService is nil
	if h.BlogService == nil {
		log.Printf("ERROR: blogService is nil in blogPostHandler")
		http.NotFound(w, r)
		return
	}
	
	// Use the blog service to get the post
	servicePost, err := h.BlogService.GetBySlug(ctx, slug)
	
	if err != nil || servicePost == nil {
		log.Printf("ERROR: Blog post '%s' not found", slug)
		http.NotFound(w, r)
		return
	}
	
	data := struct {
		Title          string
		Page           string
		Post           *blog.Post
		Config         *SiteConfig
		SiteConfig     *config.SiteConfig
		AppConfig      *config.SiteConfig
		ServicesConfig *config.ServicesConfig
	}{
		Title:          servicePost.Title + " - " + h.AppConfig.Site.Name,
		Page:           "blog",
		Post:           servicePost,
		Config:         h.SiteConfig,
		SiteConfig:     h.AppConfig,
		AppConfig:      h.AppConfig,
		ServicesConfig: servicesConfig,
	}

	if err := h.Templates.ExecuteTemplate(w, "blog-post.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// BlogContentHandler serves blog content for HTMX requests
func (h *Handler) BlogContentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Load services configuration for modal
	servicesConfig, err := h.ConfigService.LoadServicesConfig("")
	if err != nil {
		log.Printf("Warning: Failed to load services config for blog content: %v", err)
	}
	
	// Parse query parameters for pagination and filtering
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	category := r.URL.Query().Get("category")
	tag := r.URL.Query().Get("tag")
	search := r.URL.Query().Get("search")
	
	if page < 1 {
		page = 1
	}
	
	var posts []blog.Post
	var pagination *blog.PaginatedPosts
	
	// Use SQLite service if available for pagination
	if extendedService, ok := h.BlogService.(blog.ExtendedService); ok {
		params := blog.PaginationParams{
			Page:     page,
			PerPage:  12,
			Category: category,
			Tag:      tag,
			Search:   search,
		}
		
		result, err := extendedService.GetPaginatedPosts(ctx, params)
		if err != nil {
			log.Printf("Pagination error: %v", err)
			// Fallback to all posts
			posts = h.BlogService.GetAll(ctx)
		} else {
			posts = result.Posts
			pagination = result
		}
	} else {
		// Fallback to legacy behavior
		posts = h.BlogService.GetAll(ctx)
	}
	
	data := struct {
		Posts          []blog.Post
		Config         *SiteConfig
		AppConfig      *config.SiteConfig
		WorkConfig     *config.WorkConfig
		ServicesConfig *config.ServicesConfig
		BlogConfig     *blog.BlogConfig
		Pagination     *blog.PaginatedPosts
	}{
		Posts:          posts,
		Config:         h.SiteConfig,
		AppConfig:      h.AppConfig,
		WorkConfig:     h.WorkConfig,
		ServicesConfig: servicesConfig,
		BlogConfig:     h.BlogService.GetBlogConfig(),
		Pagination:     pagination,
	}

	w.Header().Set("Content-Type", "text/html")
	
	if err := h.Templates.ExecuteTemplate(w, "blog-content", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// BlogSearchHandler handles search API requests
func (h *Handler) BlogSearchHandler(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	ctx := r.Context()
	
	// Type assert to ExtendedService if available
	if extendedService, ok := h.BlogService.(blog.ExtendedService); ok {
		results, err := extendedService.SearchPosts(ctx, query, page, 12)
		if err != nil {
			log.Printf("Search error: %v", err)
			http.Error(w, "Search failed", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"posts":        results.Posts,
			"total":        results.TotalPosts,
			"total_pages":  results.TotalPages,
			"current_page": results.CurrentPage,
			"has_next":     results.HasNext,
			"has_prev":     results.HasPrev,
		})
	} else {
		// Fallback to basic search
		results := h.BlogService.Search(ctx, query)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"posts": results,
			"total": len(results),
		})
	}
}

// BlogPaginationHandler handles paginated posts API
func (h *Handler) BlogPaginationHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	category := r.URL.Query().Get("category")
	tag := r.URL.Query().Get("tag")
	search := r.URL.Query().Get("search")
	
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 12
	}
	
	ctx := r.Context()
	
	// Type assert to ExtendedService if available
	if extendedService, ok := h.BlogService.(blog.ExtendedService); ok {
		params := blog.PaginationParams{
			Page:     page,
			PerPage:  perPage,
			Category: category,
			Tag:      tag,
			Search:   search,
		}
		
		result, err := extendedService.GetPaginatedPosts(ctx, params)
		if err != nil {
			log.Printf("Pagination error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		
		// Check if HTMX request (partial update)
		if r.Header.Get("HX-Request") == "true" {
			// Load services configuration for modal
			servicesConfig, err := h.ConfigService.LoadServicesConfig("")
			if err != nil {
				log.Printf("Warning: Failed to load services config for blog pagination: %v", err)
			}
			
			// Return just the posts partial for HTMX
			data := struct {
				Posts          []blog.Post
				Config         *SiteConfig
				AppConfig      *config.SiteConfig
				ServicesConfig *config.ServicesConfig
				BlogConfig     *blog.BlogConfig
				Pagination     *blog.PaginatedPosts
			}{
				Posts:          result.Posts,
				Config:         h.SiteConfig,
				AppConfig:      h.AppConfig,
				ServicesConfig: servicesConfig,
				BlogConfig:     h.BlogService.GetBlogConfig(),
				Pagination:     result,
			}
			
			w.Header().Set("Content-Type", "text/html")
			err = h.Templates.ExecuteTemplate(w, "blog-posts-partial", data)
			if err != nil {
				log.Printf("Template error: %v", err)
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
		} else {
			// Return JSON for API calls
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)
		}
	} else {
		// Fallback to basic posts
		posts := h.BlogService.GetAll(ctx)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"posts": posts,
			"total": len(posts),
		})
	}
}

// BlogTagsHandler returns all tags with counts
func (h *Handler) BlogTagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if extendedService, ok := h.BlogService.(blog.ExtendedService); ok {
		tagCounts, err := extendedService.GetTagCounts(ctx)
		if err != nil {
			log.Printf("Tags error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tagCounts)
	} else {
		// Fallback to basic tags
		tags := h.BlogService.GetTags(ctx)
		tagCounts := make([]map[string]interface{}, len(tags))
		for i, tag := range tags {
			tagCounts[i] = map[string]interface{}{
				"tag":   tag,
				"count": 1, // Can't calculate without extended service
			}
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tagCounts)
	}
}

// BlogRelatedHandler returns related posts for a given post
func (h *Handler) BlogRelatedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 3
	}
	
	ctx := r.Context()
	
	if extendedService, ok := h.BlogService.(blog.ExtendedService); ok {
		related, err := extendedService.GetRelatedPosts(ctx, slug, limit)
		if err != nil {
			log.Printf("Related posts error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(related)
	} else {
		// Fallback - return empty array
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
	}
}

// BlogNavigationHandler returns navigation links for a post
func (h *Handler) BlogNavigationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	
	ctx := r.Context()
	
	if extendedService, ok := h.BlogService.(blog.ExtendedService); ok {
		navigation, err := extendedService.GetPostNavigation(ctx, slug)
		if err != nil {
			log.Printf("Navigation error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(navigation)
	} else {
		// Fallback - return empty navigation
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"previous": nil,
			"current":  nil,
			"next":     nil,
		})
	}
}

// BlogRSSHandler generates RSS feed
func (h *Handler) BlogRSSHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	posts := h.BlogService.GetAll(ctx)
	
	// Limit to 20 most recent posts
	if len(posts) > 20 {
		posts = posts[:20]
	}
	
	// Get site config for feed metadata
	feedTitle := "Blockhead Consulting Blog"
	feedDescription := "Technical insights on blockchain, AI, and production engineering"
	feedLink := "https://blockheadconsulting.com/blog"
	
	if h.AppConfig != nil {
		feedTitle = h.AppConfig.Site.Name + " Blog"
		feedDescription = h.AppConfig.Site.Tagline
		feedLink = h.AppConfig.Site.BaseURL + "/blog"
	}
	
	// Generate RSS XML
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>%s</title>
    <description>%s</description>
    <link>%s</link>
    <atom:link href="%s/feed.xml" rel="self" type="application/rss+xml"/>
    <language>en-us</language>
    <lastBuildDate>%s</lastBuildDate>
`, feedTitle, feedDescription, feedLink, feedLink, time.Now().Format(time.RFC1123Z))
	
	for _, post := range posts {
		// Use title as-is (RSS readers handle special characters properly)
		title := post.Title
		
		// Use clean description from social meta or cleaned summary
		description := post.OGDescription
		if description == "" {
			description = post.Summary
		}
		// Strip markdown formatting from description
		description = stripMarkdownFormatting(description)
		
		postURL := feedLink + "/" + post.Slug
		
		fmt.Fprintf(w, `    <item>
      <title>%s</title>
      <description>%s</description>
      <link>%s</link>
      <guid>%s</guid>
      <pubDate>%s</pubDate>
    </item>
`, title, description, postURL, postURL, post.Date.Format(time.RFC1123Z))
	}
	
	fmt.Fprintf(w, `  </channel>
</rss>`)
}

// stripMarkdownFormatting removes markdown formatting from text
func stripMarkdownFormatting(text string) string {
	// Remove bold/italic formatting
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "_", "")
	
	// Remove inline code
	text = strings.ReplaceAll(text, "`", "")
	
	// Remove links - [text](url) -> text
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	text = linkRegex.ReplaceAllString(text, "$1")
	
	// Remove simple links - [text] -> text
	simpleLinkRegex := regexp.MustCompile(`\[([^\]]+)\]`)
	text = simpleLinkRegex.ReplaceAllString(text, "$1")
	
	// Remove headers
	text = strings.ReplaceAll(text, "###", "")
	text = strings.ReplaceAll(text, "##", "")
	text = strings.ReplaceAll(text, "#", "")
	
	// Clean up whitespace and newlines
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	
	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")
	
	return strings.TrimSpace(text)
}