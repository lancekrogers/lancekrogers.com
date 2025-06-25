package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HealthHandler provides health check endpoint for Docker and monitoring
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version": "1.0.0",
		"services": map[string]string{
			"blog": "ok",
			"templates": "ok",
		},
	}
	
	// Check if blog service is working
	if h.BlogService != nil {
		ctx := r.Context()
		posts := h.BlogService.GetAll(ctx)
		status["services"].(map[string]string)["blog"] = fmt.Sprintf("ok (%d posts)", len(posts))
	}
	
	// Check if contact service is working
	if h.ContactService != nil {
		status["services"].(map[string]string)["contact"] = "ok"
	}
	
	// Check if email service is working
	if h.EmailService != nil {
		status["services"].(map[string]string)["email"] = "ok"
	}
	
	// Check if git storage is working
	if h.GitStorageService != nil {
		status["services"].(map[string]string)["storage"] = "ok"
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}