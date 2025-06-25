package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"blockhead.consulting/internal/bio"
	"blockhead.consulting/internal/blog"
	"blockhead.consulting/internal/config"
	"blockhead.consulting/internal/github"
)

// AboutHandler handles the about page requests
func (h *Handler) AboutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Try to get YAML configuration first, fallback to Markdown
	var aboutConfig *bio.AboutConfig
	var fullBio *bio.Bio
	
	if h.BioService != nil {
		// Try YAML first
		var err error
		aboutConfig, err = h.BioService.GetAboutConfig(ctx)
		if err != nil {
			log.Printf("Warning: Failed to load about YAML config: %v", err)
			// Fallback to Markdown
			fullBio, err = h.BioService.GetFull(ctx)
			if err != nil {
				log.Printf("Warning: Failed to load full bio: %v", err)
			}
		}
	}
	
	// Load services configuration for combined about page
	servicesConfig, err := h.ConfigService.LoadServicesConfig("")
	if err != nil {
		log.Printf("Warning: Failed to load services config for about page: %v", err)
		servicesConfig = nil
	} else {
	}
	
	// Load GitHub stats with timeout to prevent slow page loads
	var githubStats *github.GitHubStats
	if h.GitHubService != nil {
		// Create a timeout context for GitHub API calls (max 2 seconds)
		githubCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		
		githubStats, err = h.GitHubService.GetStats(githubCtx)
		if err != nil {
			log.Printf("Warning: Failed to load GitHub stats (timeout or error): %v", err)
			githubStats = nil
		}
	}
	
	// Load recent blog posts for technical insights section
	var recentPosts []blog.Post
	if h.BlogService != nil {
		posts := h.BlogService.GetAll(ctx)
		if len(posts) > 0 {
			// Get the most recent 3 posts
			limit := 3
			if len(posts) < limit {
				limit = len(posts)
			}
			recentPosts = posts[:limit]
		}
	}
	
	// Determine data structure based on available content
	if aboutConfig != nil {
		data := struct {
			Title          string
			Page           string
			Config         *SiteConfig
			AppConfig      *config.SiteConfig
			AboutConfig    *bio.AboutConfig
			WorkConfig     *config.WorkConfig
			ServicesConfig *config.ServicesConfig
			GitHubStats    *github.GitHubStats
			RecentPosts    []blog.Post
		}{
			Title:          "About Lance Rogers - Blockhead Consulting",
			Page:           "about",
			Config:         h.SiteConfig,
			AppConfig:      h.AppConfig,
			AboutConfig:    aboutConfig,
			WorkConfig:     h.WorkConfig,
			ServicesConfig: servicesConfig,
			GitHubStats:    githubStats,
			RecentPosts:    recentPosts,
		}
		
		if err := h.Templates.ExecuteTemplate(w, "page-about.html", data); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	} else {
		data := struct {
			Title          string
			Page           string
			Config         *SiteConfig
			AppConfig      *config.SiteConfig
			Bio            *bio.Bio
			WorkConfig     *config.WorkConfig
			ServicesConfig *config.ServicesConfig
			GitHubStats    *github.GitHubStats
			RecentPosts    []blog.Post
		}{
			Title:          "About Lance Rogers - Blockhead Consulting",
			Page:           "about",
			Config:         h.SiteConfig,
			AppConfig:      h.AppConfig,
			Bio:            fullBio,
			WorkConfig:     h.WorkConfig,
			ServicesConfig: servicesConfig,
			GitHubStats:    githubStats,
			RecentPosts:    recentPosts,
		}

		if err := h.Templates.ExecuteTemplate(w, "page-about.html", data); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// AboutContentHandler handles HTMX content-only about page requests
func (h *Handler) AboutContentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Try to get YAML configuration first, fallback to Markdown
	var aboutConfig *bio.AboutConfig
	var fullBio *bio.Bio
	
	if h.BioService != nil {
		// Try YAML first
		var err error
		aboutConfig, err = h.BioService.GetAboutConfig(ctx)
		if err != nil {
			log.Printf("Warning: Failed to load about YAML config: %v", err)
			// Fallback to Markdown
			fullBio, err = h.BioService.GetFull(ctx)
			if err != nil {
				log.Printf("Warning: Failed to load full bio: %v", err)
			}
		}
	}
	
	// Load services configuration for combined about page
	servicesConfig, err := h.ConfigService.LoadServicesConfig("")
	if err != nil {
		log.Printf("Warning: Failed to load services config for about content: %v", err)
	}
	
	// Load GitHub stats with timeout to prevent slow page loads
	var githubStats *github.GitHubStats
	if h.GitHubService != nil {
		// Create a timeout context for GitHub API calls (max 2 seconds)
		githubCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		
		githubStats, err = h.GitHubService.GetStats(githubCtx)
		if err != nil {
			log.Printf("Warning: Failed to load GitHub stats (timeout or error): %v", err)
			githubStats = nil
		}
	}
	
	// Load recent blog posts for technical insights section
	var recentPosts []blog.Post
	if h.BlogService != nil {
		posts := h.BlogService.GetAll(ctx)
		if len(posts) > 0 {
			// Get the most recent 3 posts
			limit := 3
			if len(posts) < limit {
				limit = len(posts)
			}
			recentPosts = posts[:limit]
		}
	}
	
	// Use the combined about page content that includes work and services
	if aboutConfig != nil {
		data := struct {
			Config         *SiteConfig
			AppConfig      *config.SiteConfig
			AboutConfig    *bio.AboutConfig
			WorkConfig     *config.WorkConfig
			ServicesConfig *config.ServicesConfig
			GitHubStats    *github.GitHubStats
			RecentPosts    []blog.Post
		}{
			Config:         h.SiteConfig,
			AppConfig:      h.AppConfig,
			AboutConfig:    aboutConfig,
			WorkConfig:     h.WorkConfig,
			ServicesConfig: servicesConfig,
			GitHubStats:    githubStats,
			RecentPosts:    recentPosts,
		}
		
		w.Header().Set("Content-Type", "text/html")
		
		if err := h.Templates.ExecuteTemplate(w, "about-page-content", data); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	} else {
		data := struct {
			Config         *SiteConfig
			AppConfig      *config.SiteConfig
			Bio            *bio.Bio
			WorkConfig     *config.WorkConfig
			ServicesConfig *config.ServicesConfig
			GitHubStats    *github.GitHubStats
			RecentPosts    []blog.Post
		}{
			Config:         h.SiteConfig,
			AppConfig:      h.AppConfig,
			Bio:            fullBio,
			WorkConfig:     h.WorkConfig,
			ServicesConfig: servicesConfig,
			GitHubStats:    githubStats,
			RecentPosts:    recentPosts,
		}

		w.Header().Set("Content-Type", "text/html")
		
		if err := h.Templates.ExecuteTemplate(w, "about-page-content", data); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}