package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type service struct {
	config *EmailConfig
}

func NewService(config *EmailConfig) Service {
	return &service{
		config: config,
	}
}

func (s *service) Send(ctx context.Context, email *Email) error {
	if email.ID == "" {
		email.ID = uuid.New().String()
	}
	if email.Timestamp.IsZero() {
		email.Timestamp = time.Now()
	}

	if err := s.validateEmail(email); err != nil {
		return &EmailError{
			Code:    ErrCodeValidation,
			Message: "email validation failed",
			Err:     err,
		}
	}

	message := s.buildMessage(email)

	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)
	
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	var err error
	if s.config.TLSEnabled {
		err = s.sendWithTLS(ctx, addr, auth, email.From, email.To, message)
	} else {
		err = smtp.SendMail(addr, auth, email.From, email.To, []byte(message))
	}

	if err != nil {
		return &EmailError{
			Code:    ErrCodeSending,
			Message: "failed to send email",
			Err:     err,
		}
	}

	return nil
}

func (s *service) SendHTML(ctx context.Context, to []string, subject, body, htmlBody string) error {
	email := &Email{
		To:       to,
		From:     s.getFromAddress(),
		Subject:  subject,
		Body:     body,
		HTMLBody: htmlBody,
	}

	return s.Send(ctx, email)
}

func (s *service) SendPlain(ctx context.Context, to []string, subject, body string) error {
	email := &Email{
		To:      to,
		From:    s.getFromAddress(),
		Subject: subject,
		Body:    body,
	}

	return s.Send(ctx, email)
}

func (s *service) ValidateConfig(ctx context.Context) error {
	if s.config.SMTPHost == "" {
		return &EmailError{
			Code:    ErrCodeConfiguration,
			Message: "SMTP host is required",
		}
	}

	if s.config.SMTPPort <= 0 {
		return &EmailError{
			Code:    ErrCodeConfiguration,
			Message: "SMTP port must be positive",
		}
	}

	if s.config.Username == "" {
		return &EmailError{
			Code:    ErrCodeConfiguration,
			Message: "SMTP username is required",
		}
	}

	if s.config.Password == "" {
		return &EmailError{
			Code:    ErrCodeConfiguration,
			Message: "SMTP password is required",
		}
	}

	if s.config.FromAddress == "" {
		return &EmailError{
			Code:    ErrCodeConfiguration,
			Message: "from address is required",
		}
	}

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)

	client, err := smtp.Dial(addr)
	if err != nil {
		return &EmailError{
			Code:    ErrCodeSMTPConnection,
			Message: "failed to connect to SMTP server",
			Err:     err,
		}
	}
	defer client.Close()

	if s.config.TLSEnabled {
		tlsConfig := &tls.Config{
			ServerName: s.config.SMTPHost,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return &EmailError{
				Code:    ErrCodeSMTPConnection,
				Message: "failed to start TLS",
				Err:     err,
			}
		}
	}

	if err := client.Auth(auth); err != nil {
		return &EmailError{
			Code:    ErrCodeAuthentication,
			Message: "SMTP authentication failed",
			Err:     err,
		}
	}

	return nil
}

func (s *service) sendWithTLS(ctx context.Context, addr string, auth smtp.Auth, from string, to []string, message string) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP server: %w", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		ServerName: s.config.SMTPHost,
	}

	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		w.Close()
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

func (s *service) validateEmail(email *Email) error {
	if len(email.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	if email.From == "" {
		return fmt.Errorf("from address is required")
	}

	if email.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if email.Body == "" && email.HTMLBody == "" {
		return fmt.Errorf("email body is required")
	}

	for _, recipient := range email.To {
		if !strings.Contains(recipient, "@") {
			return fmt.Errorf("invalid email address: %s", recipient)
		}
	}

	if !strings.Contains(email.From, "@") {
		return fmt.Errorf("invalid from address: %s", email.From)
	}

	return nil
}

func (s *service) buildMessage(email *Email) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("From: %s\r\n", s.getFromAddressWithName()))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	builder.WriteString(fmt.Sprintf("Date: %s\r\n", email.Timestamp.Format(time.RFC1123Z)))
	builder.WriteString(fmt.Sprintf("Message-ID: <%s@%s>\r\n", email.ID, s.config.SMTPHost))

	for key, value := range email.Headers {
		builder.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	if email.HTMLBody != "" {
		builder.WriteString("MIME-Version: 1.0\r\n")
		builder.WriteString("Content-Type: multipart/alternative; boundary=\"boundary123\"\r\n\r\n")
		
		if email.Body != "" {
			builder.WriteString("--boundary123\r\n")
			builder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
			builder.WriteString(email.Body)
			builder.WriteString("\r\n\r\n")
		}
		
		builder.WriteString("--boundary123\r\n")
		builder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
		builder.WriteString(email.HTMLBody)
		builder.WriteString("\r\n\r\n--boundary123--\r\n")
	} else {
		builder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
		builder.WriteString(email.Body)
		builder.WriteString("\r\n")
	}

	return builder.String()
}

func (s *service) getFromAddress() string {
	if s.config.FromAddress != "" {
		return s.config.FromAddress
	}
	return s.config.Username
}

func (s *service) getFromAddressWithName() string {
	fromAddr := s.getFromAddress()
	if s.config.FromName != "" {
		return fmt.Sprintf("%s <%s>", s.config.FromName, fromAddr)
	}
	return fromAddr
}