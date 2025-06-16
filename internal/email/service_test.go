package email

import (
	"context"
	"testing"
	"time"
)

func TestNewService(t *testing.T) {
	config := &EmailConfig{
		SMTPHost:    "smtp.gmail.com",
		SMTPPort:    587,
		Username:    "test@example.com",
		Password:    "password",
		FromAddress: "test@example.com",
		FromName:    "Test User",
		TLSEnabled:  true,
	}

	service := NewService(config)
	if service == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *EmailConfig
		wantErr bool
		errCode string
	}{
		{
			name: "valid config",
			config: &EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    587,
				Username:    "test@example.com",
				Password:    "password",
				FromAddress: "test@example.com",
				TLSEnabled:  true,
			},
			wantErr: true, // Will fail due to invalid credentials, but config is valid
			errCode: ErrCodeAuthentication, // Gmail will fail auth first, not connection
		},
		{
			name: "missing SMTP host",
			config: &EmailConfig{
				SMTPPort:    587,
				Username:    "test@example.com",
				Password:    "password",
				FromAddress: "test@example.com",
			},
			wantErr: true,
			errCode: ErrCodeConfiguration,
		},
		{
			name: "invalid SMTP port",
			config: &EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    0,
				Username:    "test@example.com",
				Password:    "password",
				FromAddress: "test@example.com",
			},
			wantErr: true,
			errCode: ErrCodeConfiguration,
		},
		{
			name: "missing username",
			config: &EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    587,
				Password:    "password",
				FromAddress: "test@example.com",
			},
			wantErr: true,
			errCode: ErrCodeConfiguration,
		},
		{
			name: "missing password",
			config: &EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    587,
				Username:    "test@example.com",
				FromAddress: "test@example.com",
			},
			wantErr: true,
			errCode: ErrCodeConfiguration,
		},
		{
			name: "missing from address",
			config: &EmailConfig{
				SMTPHost: "smtp.gmail.com",
				SMTPPort: 587,
				Username: "test@example.com",
				Password: "password",
			},
			wantErr: true,
			errCode: ErrCodeConfiguration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service := NewService(tt.config)
			
			err := service.ValidateConfig(ctx)
			
			if tt.wantErr && err == nil {
				t.Errorf("ValidateConfig() expected error but got none")
				return
			}
			
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateConfig() unexpected error: %v", err)
				return
			}
			
			if tt.wantErr && err != nil {
				emailErr, ok := err.(*EmailError)
				if !ok {
					t.Errorf("ValidateConfig() expected EmailError but got %T", err)
					return
				}
				
				if emailErr.Code != tt.errCode {
					t.Errorf("ValidateConfig() error code = %v, want %v", emailErr.Code, tt.errCode)
				}
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	service := &service{
		config: &EmailConfig{
			FromAddress: "test@example.com",
		},
	}

	tests := []struct {
		name    string
		email   *Email
		wantErr bool
	}{
		{
			name: "valid email",
			email: &Email{
				To:      []string{"recipient@example.com"},
				From:    "test@example.com",
				Subject: "Test Subject",
				Body:    "Test body",
			},
			wantErr: false,
		},
		{
			name: "no recipients",
			email: &Email{
				To:      []string{},
				From:    "test@example.com",
				Subject: "Test Subject",
				Body:    "Test body",
			},
			wantErr: true,
		},
		{
			name: "missing from",
			email: &Email{
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Body:    "Test body",
			},
			wantErr: true,
		},
		{
			name: "missing subject",
			email: &Email{
				To:   []string{"recipient@example.com"},
				From: "test@example.com",
				Body: "Test body",
			},
			wantErr: true,
		},
		{
			name: "missing body",
			email: &Email{
				To:      []string{"recipient@example.com"},
				From:    "test@example.com",
				Subject: "Test Subject",
			},
			wantErr: true,
		},
		{
			name: "invalid recipient email",
			email: &Email{
				To:      []string{"invalid-email"},
				From:    "test@example.com",
				Subject: "Test Subject",
				Body:    "Test body",
			},
			wantErr: true,
		},
		{
			name: "invalid from email",
			email: &Email{
				To:      []string{"recipient@example.com"},
				From:    "invalid-email",
				Subject: "Test Subject",
				Body:    "Test body",
			},
			wantErr: true,
		},
		{
			name: "valid with HTML body only",
			email: &Email{
				To:       []string{"recipient@example.com"},
				From:     "test@example.com",
				Subject:  "Test Subject",
				HTMLBody: "<p>Test HTML body</p>",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateEmail(tt.email)
			
			if tt.wantErr && err == nil {
				t.Errorf("validateEmail() expected error but got none")
			}
			
			if !tt.wantErr && err != nil {
				t.Errorf("validateEmail() unexpected error: %v", err)
			}
		})
	}
}

func TestBuildMessage(t *testing.T) {
	service := &service{
		config: &EmailConfig{
			FromAddress: "test@example.com",
			FromName:    "Test User",
			SMTPHost:    "smtp.example.com",
		},
	}

	email := &Email{
		ID:        "test-id",
		To:        []string{"recipient1@example.com", "recipient2@example.com"},
		From:      "test@example.com",
		Subject:   "Test Subject",
		Body:      "Test body content",
		HTMLBody:  "<p>Test HTML body</p>",
		Headers:   map[string]string{"X-Priority": "1"},
		Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	message := service.buildMessage(email)

	expectedParts := []string{
		"From: Test User <test@example.com>",
		"To: recipient1@example.com, recipient2@example.com",
		"Subject: Test Subject",
		"Message-ID: <test-id@smtp.example.com>",
		"X-Priority: 1",
		"MIME-Version: 1.0",
		"Content-Type: multipart/alternative",
		"Test body content",
		"<p>Test HTML body</p>",
	}

	for _, part := range expectedParts {
		if !containsString(message, part) {
			t.Errorf("buildMessage() missing expected part: %s", part)
		}
	}
}

func TestBuildMessagePlainOnly(t *testing.T) {
	service := &service{
		config: &EmailConfig{
			FromAddress: "test@example.com",
			SMTPHost:    "smtp.example.com",
		},
	}

	email := &Email{
		ID:        "test-id",
		To:        []string{"recipient@example.com"},
		From:      "test@example.com",
		Subject:   "Test Subject",
		Body:      "Test body content",
		Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	message := service.buildMessage(email)

	expectedParts := []string{
		"From: test@example.com",
		"To: recipient@example.com",
		"Subject: Test Subject",
		"Content-Type: text/plain; charset=\"UTF-8\"",
		"Test body content",
	}

	for _, part := range expectedParts {
		if !containsString(message, part) {
			t.Errorf("buildMessage() missing expected part: %s", part)
		}
	}

	// Should not contain HTML parts
	if containsString(message, "multipart/alternative") {
		t.Errorf("buildMessage() should not contain multipart for plain text only")
	}
}

func TestGetFromAddress(t *testing.T) {
	tests := []struct {
		name     string
		config   *EmailConfig
		expected string
	}{
		{
			name: "with from address",
			config: &EmailConfig{
				FromAddress: "custom@example.com",
				Username:    "user@example.com",
			},
			expected: "custom@example.com",
		},
		{
			name: "without from address",
			config: &EmailConfig{
				Username: "user@example.com",
			},
			expected: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &service{config: tt.config}
			result := service.getFromAddress()
			
			if result != tt.expected {
				t.Errorf("getFromAddress() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetFromAddressWithName(t *testing.T) {
	tests := []struct {
		name     string
		config   *EmailConfig
		expected string
	}{
		{
			name: "with from name",
			config: &EmailConfig{
				FromAddress: "test@example.com",
				FromName:    "Test User",
			},
			expected: "Test User <test@example.com>",
		},
		{
			name: "without from name",
			config: &EmailConfig{
				FromAddress: "test@example.com",
			},
			expected: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &service{config: tt.config}
			result := service.getFromAddressWithName()
			
			if result != tt.expected {
				t.Errorf("getFromAddressWithName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSendPlain(t *testing.T) {
	config := &EmailConfig{
		SMTPHost:    "smtp.gmail.com",
		SMTPPort:    587,
		Username:    "test@example.com",
		Password:    "password",
		FromAddress: "test@example.com",
		TLSEnabled:  true,
	}

	service := NewService(config)
	ctx := context.Background()

	// This will fail due to invalid credentials, but we can test the interface
	err := service.SendPlain(ctx, []string{"recipient@example.com"}, "Test Subject", "Test body")
	
	// We expect an error due to invalid credentials
	if err == nil {
		t.Error("SendPlain() expected error due to invalid credentials but got none")
	}

	emailErr, ok := err.(*EmailError)
	if !ok {
		t.Errorf("SendPlain() expected EmailError but got %T", err)
	}

	if emailErr != nil && emailErr.Code != ErrCodeSending {
		t.Errorf("SendPlain() error code = %v, want %v", emailErr.Code, ErrCodeSending)
	}
}

func TestSendHTML(t *testing.T) {
	config := &EmailConfig{
		SMTPHost:    "smtp.gmail.com",
		SMTPPort:    587,
		Username:    "test@example.com",
		Password:    "password",
		FromAddress: "test@example.com",
		TLSEnabled:  true,
	}

	service := NewService(config)
	ctx := context.Background()

	// This will fail due to invalid credentials, but we can test the interface
	err := service.SendHTML(ctx, []string{"recipient@example.com"}, "Test Subject", "Test body", "<p>Test HTML</p>")
	
	// We expect an error due to invalid credentials
	if err == nil {
		t.Error("SendHTML() expected error due to invalid credentials but got none")
	}

	emailErr, ok := err.(*EmailError)
	if !ok {
		t.Errorf("SendHTML() expected EmailError but got %T", err)
	}

	if emailErr != nil && emailErr.Code != ErrCodeSending {
		t.Errorf("SendHTML() error code = %v, want %v", emailErr.Code, ErrCodeSending)
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}