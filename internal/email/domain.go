package email

import (
	"context"
	"time"
)

type Email struct {
	ID          string            `json:"id"`
	To          []string          `json:"to"`
	From        string            `json:"from"`
	Subject     string            `json:"subject"`
	Body        string            `json:"body"`
	HTMLBody    string            `json:"html_body,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	Username     string
	Password     string
	FromAddress  string
	FromName     string
	TLSEnabled   bool
}

type Service interface {
	Send(ctx context.Context, email *Email) error
	SendHTML(ctx context.Context, to []string, subject, body, htmlBody string) error
	SendPlain(ctx context.Context, to []string, subject, body string) error
	ValidateConfig(ctx context.Context) error
}

type EmailError struct {
	Code    string
	Message string
	Err     error
}

func (e *EmailError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *EmailError) Unwrap() error {
	return e.Err
}

const (
	ErrCodeSMTPConnection = "SMTP_CONNECTION_FAILED"
	ErrCodeAuthentication = "SMTP_AUTH_FAILED"
	ErrCodeSending        = "EMAIL_SEND_FAILED"
	ErrCodeValidation     = "EMAIL_VALIDATION_FAILED"
	ErrCodeConfiguration  = "EMAIL_CONFIG_INVALID"
)