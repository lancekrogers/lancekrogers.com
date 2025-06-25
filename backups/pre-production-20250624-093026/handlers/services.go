package handlers

import (
	"log"
	"net/http"

	"blockhead.consulting/internal/config"
)

// ServicesModalHandler serves the services modal content
func (h *Handler) ServicesModalHandler(w http.ResponseWriter, r *http.Request) {
	// Get service type from query parameter
	serviceType := r.URL.Query().Get("type")
	
	data := struct {
		ServiceType string
		Config      *SiteConfig
		AppConfig   *config.SiteConfig
	}{
		ServiceType: serviceType,
		Config:      h.SiteConfig,
		AppConfig:   h.AppConfig,
	}
	
	w.Header().Set("Content-Type", "text/html")
	
	if err := h.Templates.ExecuteTemplate(w, "services-modal", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}