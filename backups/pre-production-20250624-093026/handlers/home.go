package handlers

import (
	"log"
	"net/http"

	"blockhead.consulting/internal/bio"
	"blockhead.consulting/internal/blog"
	"blockhead.consulting/internal/config"
)

// HomeHandler handles the home page requests
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get brief bio
	var bioBrief *bio.Bio
	if h.BioService != nil {
		var err error
		bioBrief, err = h.BioService.GetBrief(ctx)
		if err != nil {
			log.Printf("Warning: Failed to load brief bio: %v", err)
		}
	}
	
	// Get recent blog posts
	var recentPosts []blog.Post
	if h.BlogService != nil && h.SiteConfig.BlogEnabled {
		allPosts := h.BlogService.GetAll(ctx)
		log.Printf("DEBUG: Found %d blog posts for homepage", len(allPosts))
		// Get latest 3 posts
		if len(allPosts) > 3 {
			recentPosts = allPosts[:3]
		} else {
			recentPosts = allPosts
		}
		log.Printf("DEBUG: Passing %d recent posts to template", len(recentPosts))
	} else {
		log.Printf("DEBUG: Blog service nil: %v, Blog enabled: %v", h.BlogService == nil, h.SiteConfig.BlogEnabled)
	}
	
	// Load services configuration for modal
	servicesConfig, err := h.ConfigService.LoadServicesConfig("")
	if err != nil {
		log.Printf("Warning: Failed to load services config for home page: %v", err)
	}
	
	// Determine title based on config availability
	title := "Blockhead Consulting - Enterprise Blockchain & AI Infrastructure"
	if h.AppConfig != nil && h.AppConfig.Site.Name != "" {
		title = h.AppConfig.Site.Name + " - " + h.AppConfig.Site.Tagline
	}
	
	data := struct {
		Title          string
		Page           string
		Config         *SiteConfig
		AppConfig      *config.SiteConfig
		BioBrief       *bio.Bio
		RecentPosts    []blog.Post
		ServicesConfig *config.ServicesConfig
	}{
		Title:          title,
		Page:           "home",
		Config:         h.SiteConfig,
		AppConfig:      h.AppConfig,
		BioBrief:       bioBrief,
		RecentPosts:    recentPosts,
		ServicesConfig: servicesConfig,
	}

	// Use ExecuteTemplate directly with the specific page template
	if err := h.Templates.ExecuteTemplate(w, "page-home.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// HomeContentHandler handles HTMX content-only home page requests
func (h *Handler) HomeContentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get brief bio
	var bioBrief *bio.Bio
	if h.BioService != nil {
		var err error
		bioBrief, err = h.BioService.GetBrief(ctx)
		if err != nil {
			log.Printf("Warning: Failed to load brief bio: %v", err)
		}
	}
	
	// Get recent blog posts
	var recentPosts []blog.Post
	if h.BlogService != nil && h.SiteConfig.BlogEnabled {
		allPosts := h.BlogService.GetAll(ctx)
		// Get latest 3 posts
		if len(allPosts) > 3 {
			recentPosts = allPosts[:3]
		} else {
			recentPosts = allPosts
		}
	}
	
	// Load services configuration for modal
	servicesConfig, err := h.ConfigService.LoadServicesConfig("")
	if err != nil {
		log.Printf("Warning: Failed to load services config for home content: %v", err)
	}
	
	data := struct {
		Config         *SiteConfig
		AppConfig      *config.SiteConfig
		BioBrief       *bio.Bio
		RecentPosts    []blog.Post
		ServicesConfig *config.ServicesConfig
	}{
		Config:         h.SiteConfig,
		AppConfig:      h.AppConfig,
		BioBrief:       bioBrief,
		RecentPosts:    recentPosts,
		ServicesConfig: servicesConfig,
	}

	w.Header().Set("Content-Type", "text/html")
	
	if err := h.Templates.ExecuteTemplate(w, "home-content", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}