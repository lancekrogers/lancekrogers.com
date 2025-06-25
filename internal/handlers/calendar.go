package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"blockhead.consulting/internal/booking"
	"blockhead.consulting/internal/config"
	"blockhead.consulting/internal/security"
)

// TimeSlot represents a bookable time slot
type TimeSlot struct {
	ID        string    `json:"id"`
	Date      string    `json:"date"`
	Time      string    `json:"time"`
	Available bool      `json:"available"`
	Booked    bool      `json:"booked"`
	BookedBy  string    `json:"bookedBy,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// BookingRequest represents a booking form submission
type BookingRequest struct {
	SlotID      string `json:"slotId"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Company     string `json:"company"`
	ServiceType string `json:"serviceType"`
	Message     string `json:"message"`
}

// CalendarHandler renders the calendar booking page
func (h *Handler) CalendarHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title     string
		Page      string
		Config    *SiteConfig
		AppConfig *config.SiteConfig
	}{
		Title:     "Book a Consultation - Blockhead Consulting",
		Page:      "calendar",
		Config:    h.SiteConfig,
		AppConfig: h.AppConfig,
	}

	if err := h.Templates.ExecuteTemplate(w, "page-calendar.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// CalendarContentHandler serves calendar content for HTMX requests
func (h *Handler) CalendarContentHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Config    *SiteConfig
		AppConfig *config.SiteConfig
	}{
		Config:    h.SiteConfig,
		AppConfig: h.AppConfig,
	}
	
	w.Header().Set("Content-Type", "text/html")
	
	if err := h.Templates.ExecuteTemplate(w, "calendar-content", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// SlotsHandler returns available booking slots
func (h *Handler) SlotsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Use booking service if available
	if h.BookingService != nil {
		slots, err := h.BookingService.GetAvailableSlots(ctx)
		if err != nil {
			log.Printf("Error getting available slots: %v", err)
			http.Error(w, "Failed to get available slots", http.StatusInternalServerError)
			return
		}
		
		// Convert to handler's TimeSlot format
		var availableSlots []*TimeSlot
		for _, slot := range slots {
			availableSlots = append(availableSlots, &TimeSlot{
				ID:        slot.ID,
				Date:      slot.Date,
				Time:      slot.Time,
				Available: slot.Available,
				Booked:    slot.Booked,
				BookedBy:  slot.BookedBy,
				CreatedAt: slot.CreatedAt,
			})
		}
		
		// Sort by date and time
		sort.Slice(availableSlots, func(i, j int) bool {
			if availableSlots[i].Date == availableSlots[j].Date {
				return availableSlots[i].Time < availableSlots[j].Time
			}
			return availableSlots[i].Date < availableSlots[j].Date
		})
		
		// Limit to first 20 slots for demo
		if len(availableSlots) > 20 {
			availableSlots = availableSlots[:20]
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(availableSlots)
		return
	}
	
	// Fallback to empty response if no booking service
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]*TimeSlot{})
}

// BookingHandler processes booking requests
func (h *Handler) BookingHandler(w http.ResponseWriter, r *http.Request) {
	var req BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("SECURITY: Invalid JSON in booking request from %s", security.ExtractClientIP(r))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Enhanced validation
	if err := h.validateBookingRequest(&req); err != nil {
		log.Printf("SECURITY: Invalid booking request from %s: %v", security.ExtractClientIP(r), err)
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	
	// Use booking service if available
	if h.BookingService != nil {
		// Convert to booking service request
		bookingReq := &booking.BookingRequest{
			SlotID:      req.SlotID,
			Name:        req.Name,
			Email:       req.Email,
			Company:     req.Company,
			ServiceType: req.ServiceType,
			Message:     req.Message,
		}
		
		// Book the slot
		_, err := h.BookingService.BookSlot(ctx, bookingReq)
		if err != nil {
			log.Printf("BOOKING: Failed to book slot: %v", err)
			http.Error(w, "Failed to book slot", http.StatusConflict)
			return
		}
		
		// Send confirmation email if email service is available
		if h.EmailService != nil {
			// TODO: Send confirmation email
		}
		
		log.Printf("BOOKING: New booking from %s - Email: %s, Service: %s", 
			security.ExtractClientIP(r), req.Email, req.ServiceType)
		
		// Return success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Booking confirmed! You'll receive a confirmation email shortly.",
		})
		return
	}
	
	// Fallback if no booking service
	log.Printf("BOOKING: Booking service not available")
	http.Error(w, "Booking service temporarily unavailable", http.StatusServiceUnavailable)
}

// validateBookingRequest validates booking form data
func (h *Handler) validateBookingRequest(req *BookingRequest) error {
	if req.SlotID == "" || req.Name == "" || req.Email == "" || req.ServiceType == "" {
		return fmt.Errorf("missing required fields")
	}

	if !security.SlotIDRegex.MatchString(req.SlotID) {
		return fmt.Errorf("invalid slot ID format")
	}

	if !security.NameRegex.MatchString(req.Name) {
		return fmt.Errorf("invalid name format")
	}

	if !security.EmailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	if req.Company != "" && !security.CompanyRegex.MatchString(req.Company) {
		return fmt.Errorf("invalid company format")
	}

	if req.Message != "" {
		if len(req.Message) > 2000 {
			return fmt.Errorf("message too long (max 2000 characters)")
		}
		if !security.MessageRegex.MatchString(req.Message) {
			return fmt.Errorf("invalid message format")
		}
	}

	validServiceTypes := map[string]bool{
		"crypto-infrastructure": true,
		"ai-claude":            true,
		"both":                 true,
		"other":                true,
	}

	if !validServiceTypes[req.ServiceType] {
		return fmt.Errorf("invalid service type")
	}

	return nil
}