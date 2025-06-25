package booking

import (
	"time"
)

// Booking represents a booking in the system
type Booking struct {
	ID          string    `json:"id"`
	SlotID      string    `json:"slotId"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Company     string    `json:"company"`
	ServiceType string    `json:"serviceType"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

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