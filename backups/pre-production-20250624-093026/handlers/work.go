package handlers

import (
	"net/http"
)

// WorkRedirectHandler redirects /work to /about
func (h *Handler) WorkRedirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/about", http.StatusMovedPermanently)
}

// WorkContentRedirectHandler redirects /content/work to /content/about
func (h *Handler) WorkContentRedirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/content/about", http.StatusMovedPermanently)
}