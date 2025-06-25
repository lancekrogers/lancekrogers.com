package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Service provides booking management functionality
type Service interface {
	// GetAvailableSlots returns available booking slots
	GetAvailableSlots(ctx context.Context) ([]*TimeSlot, error)
	
	// GetSlot returns a specific time slot by ID
	GetSlot(ctx context.Context, slotID string) (*TimeSlot, error)
	
	// BookSlot books a time slot
	BookSlot(ctx context.Context, req *BookingRequest) (*Booking, error)
	
	// GetBookings returns all bookings
	GetBookings(ctx context.Context) ([]*Booking, error)
	
	// GetBooking returns a specific booking by ID
	GetBooking(ctx context.Context, bookingID string) (*Booking, error)
	
	// CancelBooking cancels a booking
	CancelBooking(ctx context.Context, bookingID string) error
}

// FileBasedService implements Service using file storage
type FileBasedService struct {
	mu           sync.RWMutex
	timeSlots    map[string]*TimeSlot
	bookings     map[string]*Booking
	bookingsFile string
	logger       *log.Logger
}

// NewFileBasedService creates a new file-based booking service
func NewFileBasedService(bookingsFile string, logger *log.Logger) (*FileBasedService, error) {
	s := &FileBasedService{
		timeSlots:    make(map[string]*TimeSlot),
		bookings:     make(map[string]*Booking),
		bookingsFile: bookingsFile,
		logger:       logger,
	}
	
	// Initialize time slots
	s.initializeTimeSlots()
	
	// Load existing bookings
	if err := s.loadBookings(); err != nil {
		return nil, fmt.Errorf("failed to load bookings: %w", err)
	}
	
	return s, nil
}

// GetAvailableSlots returns available booking slots
func (s *FileBasedService) GetAvailableSlots(ctx context.Context) ([]*TimeSlot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var available []*TimeSlot
	for _, slot := range s.timeSlots {
		if slot.Available && !slot.Booked {
			available = append(available, slot)
		}
	}
	
	return available, nil
}

// GetSlot returns a specific time slot by ID
func (s *FileBasedService) GetSlot(ctx context.Context, slotID string) (*TimeSlot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	slot, exists := s.timeSlots[slotID]
	if !exists {
		return nil, fmt.Errorf("slot not found: %s", slotID)
	}
	
	return slot, nil
}

// BookSlot books a time slot
func (s *FileBasedService) BookSlot(ctx context.Context, req *BookingRequest) (*Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Check if slot exists and is available
	slot, exists := s.timeSlots[req.SlotID]
	if !exists {
		return nil, fmt.Errorf("slot not found: %s", req.SlotID)
	}
	
	if !slot.Available || slot.Booked {
		return nil, fmt.Errorf("slot not available: %s", req.SlotID)
	}
	
	// Create booking
	booking := &Booking{
		ID:          fmt.Sprintf("booking-%d", time.Now().UnixNano()),
		SlotID:      req.SlotID,
		Name:        req.Name,
		Email:       req.Email,
		Company:     req.Company,
		ServiceType: req.ServiceType,
		Message:     req.Message,
		Status:      "confirmed",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Update slot
	slot.Booked = true
	slot.BookedBy = req.Email
	
	// Save booking
	s.bookings[booking.ID] = booking
	
	// Persist to file
	if err := s.saveBookings(); err != nil {
		// Rollback changes
		slot.Booked = false
		slot.BookedBy = ""
		delete(s.bookings, booking.ID)
		return nil, fmt.Errorf("failed to save booking: %w", err)
	}
	
	s.logger.Printf("Booking created: %s for slot %s", booking.ID, req.SlotID)
	return booking, nil
}

// GetBookings returns all bookings
func (s *FileBasedService) GetBookings(ctx context.Context) ([]*Booking, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	bookings := make([]*Booking, 0, len(s.bookings))
	for _, booking := range s.bookings {
		bookings = append(bookings, booking)
	}
	
	return bookings, nil
}

// GetBooking returns a specific booking by ID
func (s *FileBasedService) GetBooking(ctx context.Context, bookingID string) (*Booking, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	booking, exists := s.bookings[bookingID]
	if !exists {
		return nil, fmt.Errorf("booking not found: %s", bookingID)
	}
	
	return booking, nil
}

// CancelBooking cancels a booking
func (s *FileBasedService) CancelBooking(ctx context.Context, bookingID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	booking, exists := s.bookings[bookingID]
	if !exists {
		return fmt.Errorf("booking not found: %s", bookingID)
	}
	
	// Free up the slot
	if slot, exists := s.timeSlots[booking.SlotID]; exists {
		slot.Booked = false
		slot.BookedBy = ""
	}
	
	// Update booking status
	booking.Status = "cancelled"
	booking.UpdatedAt = time.Now()
	
	// Save changes
	if err := s.saveBookings(); err != nil {
		return fmt.Errorf("failed to save booking cancellation: %w", err)
	}
	
	s.logger.Printf("Booking cancelled: %s", bookingID)
	return nil
}

// initializeTimeSlots creates available slots for the next 30 days
func (s *FileBasedService) initializeTimeSlots() {
	now := time.Now()
	for i := 0; i < 30; i++ {
		date := now.AddDate(0, 0, i)
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue
		}

		dateStr := date.Format("2006-01-02")
		for hour := 10; hour < 17; hour++ {
			timeStr := fmt.Sprintf("%02d:00", hour)
			slotID := fmt.Sprintf("%s-%s", dateStr, timeStr)

			if _, exists := s.timeSlots[slotID]; !exists {
				s.timeSlots[slotID] = &TimeSlot{
					ID:        slotID,
					Date:      dateStr,
					Time:      timeStr,
					Available: true,
					Booked:    false,
					CreatedAt: now,
				}
			}
		}
	}
}

// loadBookings loads bookings from file
func (s *FileBasedService) loadBookings() error {
	data, err := os.ReadFile(s.bookingsFile)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's OK
			return nil
		}
		return err
	}

	var persistedData struct {
		TimeSlots map[string]*TimeSlot `json:"timeSlots"`
		Bookings  map[string]*Booking  `json:"bookings"`
	}
	
	if err := json.Unmarshal(data, &persistedData); err != nil {
		return err
	}

	// Merge with existing slots
	for id, slot := range persistedData.TimeSlots {
		s.timeSlots[id] = slot
	}
	
	// Load bookings
	s.bookings = persistedData.Bookings
	if s.bookings == nil {
		s.bookings = make(map[string]*Booking)
	}
	
	return nil
}

// saveBookings saves bookings to file
func (s *FileBasedService) saveBookings() error {
	// Create data directory if it doesn't exist
	os.MkdirAll(filepath.Dir(s.bookingsFile), 0755)

	persistedData := struct {
		TimeSlots map[string]*TimeSlot `json:"timeSlots"`
		Bookings  map[string]*Booking  `json:"bookings"`
	}{
		TimeSlots: s.timeSlots,
		Bookings:  s.bookings,
	}

	data, err := json.MarshalIndent(persistedData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.bookingsFile, data, 0644)
}