package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"blockhead.consulting/internal/contact"
)

// ContactHandler processes contact form submissions
func (h *Handler) ContactHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Create contact request from form data
	req := &contact.ContactRequest{
		Name:    strings.TrimSpace(r.FormValue("name")),
		Email:   strings.TrimSpace(r.FormValue("email")),
		Company: strings.TrimSpace(r.FormValue("company")),
		Message: strings.TrimSpace(r.FormValue("message")),
	}

	// Process contact form using the contact service
	if h.ContactService != nil {
		log.Printf("CONTACT: Processing form from %s <%s>", req.Name, req.Email)
		_, err := h.ContactService.ProcessContactForm(ctx, req, r)
		if err != nil {
			log.Printf("CONTACT: Error processing form: %v", err)
			// Handle validation errors with HTMX-friendly response
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<div class="alert error">%s</div>`, err.Error())
			return
		}
		log.Printf("CONTACT: Form processed successfully")
	} else {
		// Fallback for when contact service is not available
		log.Printf("Contact form (fallback): Name=%s, Email=%s, Message=%s", req.Name, req.Email, req.Message)
	}

	// Return HTMX success response
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<div class="alert success">Message sent successfully! I'll get back to you within 24 hours.</div>`)
}