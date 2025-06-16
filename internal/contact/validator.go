package contact

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	// nameRegex allows letters, spaces, hyphens, and apostrophes
	nameRegex = regexp.MustCompile(`^[a-zA-Z\s'-]{1,100}$`)
	
	// companyRegex allows alphanumeric, spaces, and common business characters
	companyRegex = regexp.MustCompile(`^[a-zA-Z0-9\s&.,'()-]{0,200}$`)
	
	// messageRegex allows most printable characters including newlines
	messageRegex = regexp.MustCompile(`^[\p{L}\p{N}\p{P}\p{S}\s\n\r]+$`)
)

// Validator validates contact form submissions
type Validator struct {
	minMessageLength int
	maxMessageLength int
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		minMessageLength: 10,
		maxMessageLength: 5000,
	}
}

// Validate validates a contact request
func (v *Validator) Validate(req *ContactRequest) ValidationErrors {
	var errors ValidationErrors
	
	// Validate name
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Name is required",
		})
	} else if !nameRegex.MatchString(req.Name) {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Name contains invalid characters",
		})
	} else if utf8.RuneCountInString(req.Name) < 2 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Name must be at least 2 characters long",
		})
	}
	
	// Validate email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Email is required",
		})
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Invalid email address",
		})
	}
	
	// Validate company (optional)
	req.Company = strings.TrimSpace(req.Company)
	if req.Company != "" && !companyRegex.MatchString(req.Company) {
		errors = append(errors, ValidationError{
			Field:   "company",
			Message: "Company name contains invalid characters",
		})
	}
	
	// Validate message
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		errors = append(errors, ValidationError{
			Field:   "message",
			Message: "Message is required",
		})
	} else if utf8.RuneCountInString(req.Message) < v.minMessageLength {
		errors = append(errors, ValidationError{
			Field:   "message",
			Message: "Message must be at least 10 characters long",
		})
	} else if utf8.RuneCountInString(req.Message) > v.maxMessageLength {
		errors = append(errors, ValidationError{
			Field:   "message",
			Message: "Message must not exceed 5000 characters",
		})
	} else if !messageRegex.MatchString(req.Message) {
		errors = append(errors, ValidationError{
			Field:   "message",
			Message: "Message contains invalid characters",
		})
	}
	
	return errors
}

// SanitizeHTML removes potentially dangerous HTML from a string
func SanitizeHTML(input string) string {
	// Basic HTML escape - in production, use a proper HTML sanitizer
	replacer := strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"&", "&amp;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(input)
}