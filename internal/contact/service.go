package contact

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"blockhead.consulting/internal/email"
	"blockhead.consulting/internal/errors"
	"blockhead.consulting/internal/events"
	"blockhead.consulting/internal/storage/git"
)

// Service provides contact form functionality
type Service interface {
	// ProcessContactForm processes a contact form submission
	ProcessContactForm(ctx context.Context, req *ContactRequest, r *http.Request) (*ContactMessage, error)
	
	// GetMessage retrieves a message by ID
	GetMessage(ctx context.Context, id string) (*ContactMessage, error)
	
	// ListMessages lists all messages with optional status filter
	ListMessages(ctx context.Context, status string) ([]*ContactMessage, error)
	
	// UpdateMessageStatus updates the status of a message
	UpdateMessageStatus(ctx context.Context, id string, status string) error
}

// service implements the contact form service
type service struct {
	validator    *Validator
	storage      git.Service
	eventBus     events.EventBus
	emailService email.Service
	logger       *log.Logger
	adminEmail   string
}

// NewService creates a new contact service
func NewService(storage git.Service, eventBus events.EventBus, emailService email.Service, adminEmail string, logger *log.Logger) Service {
	if logger == nil {
		logger = log.Default()
	}
	
	return &service{
		validator:    NewValidator(),
		storage:      storage,
		eventBus:     eventBus,
		emailService: emailService,
		logger:       logger,
		adminEmail:   adminEmail,
	}
}

// ProcessContactForm processes a contact form submission
func (s *service) ProcessContactForm(ctx context.Context, req *ContactRequest, r *http.Request) (*ContactMessage, error) {
	// Validate request
	if errs := s.validator.Validate(req); len(errs) > 0 {
		return nil, errors.New(errors.ErrCodeValidation, "validation failed").
			WithContext("errors", errs)
	}
	
	// Generate message ID
	id, err := generateID()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to generate message ID")
	}
	
	// Extract request metadata
	ip := extractIP(r)
	userAgent := r.Header.Get("User-Agent")
	
	// Create message
	message := &ContactMessage{
		ID:        id,
		Name:      req.Name,
		Email:     req.Email,
		Company:   req.Company,
		Message:   req.Message,
		IP:        ip,
		UserAgent: userAgent,
		Timestamp: time.Now().UTC(),
		Status:    "new",
	}
	
	// Convert to storage message
	storageMsg := &git.Message{
		ID:        message.ID,
		Name:      message.Name,
		Email:     message.Email,
		Company:   message.Company,
		Message:   message.Message,
		IP:        message.IP,
		UserAgent: message.UserAgent,
		Timestamp: message.Timestamp,
		Status:    message.Status,
	}
	
	// Save to storage (only if storage is available)
	if s.storage != nil {
		if err := s.storage.SaveMessage(ctx, storageMsg); err != nil {
			return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to save message")
		}
		s.logger.Printf("CONTACT: Message %s saved to storage", message.ID)
	} else {
		s.logger.Printf("CONTACT: Storage not available - message %s logged only", message.ID)
	}
	
	// Send email notification
	if s.emailService != nil {
		if err := s.sendNotificationEmail(ctx, message); err != nil {
			// Log but don't fail the request
			s.logger.Printf("CONTACT: Failed to send notification email: %v", err)
		}
	}
	
	// Publish event
	if s.eventBus != nil {
		event := events.NewEventWithContext(ctx,
			events.EventMessageReceived,
			map[string]interface{}{
				"message_id": message.ID,
				"name":       message.Name,
				"email":      message.Email,
				"timestamp":  message.Timestamp,
			},
		)
		
		if err := s.eventBus.Publish(ctx, event); err != nil {
			// Log but don't fail the request
			s.logger.Printf("CONTACT: Failed to publish event: %v", err)
		}
	}
	
	s.logger.Printf("CONTACT: Processed message %s from %s <%s>", message.ID, message.Name, message.Email)
	
	return message, nil
}

// GetMessage retrieves a message by ID
func (s *service) GetMessage(ctx context.Context, id string) (*ContactMessage, error) {
	storageMsg, err := s.storage.GetMessage(ctx, id)
	if err != nil {
		return nil, err
	}
	
	return s.convertFromStorage(storageMsg), nil
}

// ListMessages lists all messages with optional status filter
func (s *service) ListMessages(ctx context.Context, status string) ([]*ContactMessage, error) {
	storageMsgs, err := s.storage.ListMessages(ctx, status)
	if err != nil {
		return nil, err
	}
	
	messages := make([]*ContactMessage, len(storageMsgs))
	for i, msg := range storageMsgs {
		messages[i] = s.convertFromStorage(msg)
	}
	
	return messages, nil
}

// UpdateMessageStatus updates the status of a message
func (s *service) UpdateMessageStatus(ctx context.Context, id string, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"new":     true,
		"read":    true,
		"replied": true,
		"closed":  true,
	}
	
	if !validStatuses[status] {
		return errors.New(errors.ErrCodeValidation, "invalid status")
	}
	
	return s.storage.UpdateStatus(ctx, id, status)
}

// convertFromStorage converts a storage message to a contact message
func (s *service) convertFromStorage(msg *git.Message) *ContactMessage {
	return &ContactMessage{
		ID:        msg.ID,
		Name:      msg.Name,
		Email:     msg.Email,
		Company:   msg.Company,
		Message:   msg.Message,
		IP:        msg.IP,
		UserAgent: msg.UserAgent,
		Timestamp: msg.Timestamp,
		Status:    msg.Status,
	}
}

// generateID generates a unique message ID
func generateID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("msg_%s", hex.EncodeToString(bytes)), nil
}

// extractIP extracts the client IP from the request
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return xff
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	
	return r.RemoteAddr
}

// sendNotificationEmail sends an email notification about a new contact message
func (s *service) sendNotificationEmail(ctx context.Context, message *ContactMessage) error {
	subject := fmt.Sprintf("New Contact Form Submission - %s", message.Name)
	
	// Create plain text body
	plainBody := fmt.Sprintf(`New contact form submission received:

Name: %s
Email: %s
Company: %s
Message ID: %s
Timestamp: %s
IP Address: %s

Message:
%s

---
This is an automated notification from the Blockhead Consulting website.`,
		message.Name,
		message.Email,
		message.Company,
		message.ID,
		message.Timestamp.Format("2006-01-02 15:04:05 UTC"),
		message.IP,
		message.Message,
	)
	
	// Create HTML body
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #1a1a1a; color: #00ff88; padding: 20px; border-radius: 8px 8px 0 0; }
        .content { background: #f8f9fa; padding: 20px; border-radius: 0 0 8px 8px; }
        .field { margin-bottom: 15px; }
        .label { font-weight: bold; color: #555; }
        .value { margin-top: 5px; }
        .message-content { background: white; padding: 15px; border-radius: 4px; border-left: 4px solid #00ff88; }
        .footer { margin-top: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>🔗 New Contact Form Submission</h2>
        </div>
        <div class="content">
            <div class="field">
                <div class="label">Name:</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Email:</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Company:</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Message ID:</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Timestamp:</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">IP Address:</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Message:</div>
                <div class="message-content">%s</div>
            </div>
            <div class="footer">
                This is an automated notification from the Blockhead Consulting website.
            </div>
        </div>
    </div>
</body>
</html>`,
		message.Name,
		message.Email,
		message.Company,
		message.ID,
		message.Timestamp.Format("2006-01-02 15:04:05 UTC"),
		message.IP,
		strings.ReplaceAll(message.Message, "\n", "<br>"),
	)
	
	// Send notification email to configured admin
	recipients := []string{s.adminEmail}
	if s.adminEmail == "" {
		// Skip email if no admin email configured
		return nil
	}
	
	return s.emailService.SendHTML(ctx, recipients, subject, plainBody, htmlBody)
}