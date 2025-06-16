package contact

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"blockhead.consulting/internal/email"
	"blockhead.consulting/internal/errors"
	"blockhead.consulting/internal/events"
	"blockhead.consulting/internal/storage/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock storage service
type mockStorage struct {
	messages map[string]*git.Message
	saveErr  error
	getErr   error
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		messages: make(map[string]*git.Message),
	}
}

func (m *mockStorage) SaveMessage(ctx context.Context, message *git.Message) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.messages[message.ID] = message
	return nil
}

func (m *mockStorage) GetMessage(ctx context.Context, id string) (*git.Message, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	msg, exists := m.messages[id]
	if !exists {
		return nil, errors.NotFound("message")
	}
	return msg, nil
}

func (m *mockStorage) ListMessages(ctx context.Context, status string) ([]*git.Message, error) {
	var messages []*git.Message
	for _, msg := range m.messages {
		if status == "" || msg.Status == status {
			messages = append(messages, msg)
		}
	}
	return messages, nil
}

func (m *mockStorage) UpdateStatus(ctx context.Context, id string, status string) error {
	msg, exists := m.messages[id]
	if !exists {
		return errors.NotFound("message")
	}
	msg.Status = status
	return nil
}

// Mock event bus
type mockEventBus struct {
	publishedEvents []events.Event
}

func (m *mockEventBus) Subscribe(eventType events.EventType, handler events.EventHandler) string {
	return "mock-subscription"
}

func (m *mockEventBus) Unsubscribe(subscriptionID string) {}

func (m *mockEventBus) Publish(ctx context.Context, event events.Event) error {
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *mockEventBus) Start(ctx context.Context) error { return nil }
func (m *mockEventBus) Stop() error { return nil }

// Mock email service
type mockEmailService struct {
	sentEmails []email.Email
	sendErr    error
}

func (m *mockEmailService) Send(ctx context.Context, email *email.Email) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sentEmails = append(m.sentEmails, *email)
	return nil
}

func (m *mockEmailService) SendHTML(ctx context.Context, to []string, subject, body, htmlBody string) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	email := email.Email{
		To:       to,
		Subject:  subject,
		Body:     body,
		HTMLBody: htmlBody,
	}
	m.sentEmails = append(m.sentEmails, email)
	return nil
}

func (m *mockEmailService) SendPlain(ctx context.Context, to []string, subject, body string) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	email := email.Email{
		To:      to,
		Subject: subject,
		Body:    body,
	}
	m.sentEmails = append(m.sentEmails, email)
	return nil
}

func (m *mockEmailService) ValidateConfig(ctx context.Context) error {
	return nil
}

func createTestService() (Service, *mockStorage, *mockEventBus, *mockEmailService) {
	storage := newMockStorage()
	eventBus := &mockEventBus{}
	emailService := &mockEmailService{}
	logger := log.New(os.Stdout, "[contact-test] ", log.LstdFlags)
	
	service := NewService(storage, eventBus, emailService, "test@example.com", logger)
	return service, storage, eventBus, emailService
}

func TestContactService(t *testing.T) {
	svc, storage, eventBus, emailService := createTestService()
	ctx := context.Background()
	
	t.Run("process valid contact form", func(t *testing.T) {
		req := &ContactRequest{
			Name:    "Jane Doe",
			Email:   "jane@example.com",
			Company: "Tech Corp",
			Message: "I need help with blockchain integration.",
		}
		
		// Create HTTP request
		r := httptest.NewRequest("POST", "/contact", nil)
		r.RemoteAddr = "192.168.1.100:12345"
		r.Header.Set("User-Agent", "Test Browser/1.0")
		r.Header.Set("X-Forwarded-For", "10.0.0.1, 192.168.1.100")
		
		// Process form
		msg, err := svc.ProcessContactForm(ctx, req, r)
		require.NoError(t, err)
		assert.NotEmpty(t, msg.ID)
		assert.Equal(t, "Jane Doe", msg.Name)
		assert.Equal(t, "jane@example.com", msg.Email)
		assert.Equal(t, "Tech Corp", msg.Company)
		assert.Equal(t, "I need help with blockchain integration.", msg.Message)
		assert.Equal(t, "10.0.0.1", msg.IP) // Should extract from X-Forwarded-For
		assert.Equal(t, "Test Browser/1.0", msg.UserAgent)
		assert.Equal(t, "new", msg.Status)
		assert.WithinDuration(t, time.Now().UTC(), msg.Timestamp, 2*time.Second)
		
		// Check message was saved
		assert.Len(t, storage.messages, 1)
		
		// Check event was published
		assert.Len(t, eventBus.publishedEvents, 1)
		event := eventBus.publishedEvents[0]
		assert.Equal(t, events.EventMessageReceived, event.Type())
		
		// Check email notification was sent
		assert.Len(t, emailService.sentEmails, 1)
		sentEmail := emailService.sentEmails[0]
		assert.Contains(t, sentEmail.Subject, "Jane Doe")
		assert.Contains(t, sentEmail.Body, "jane@example.com")
		assert.Contains(t, sentEmail.HTMLBody, "Tech Corp")
		assert.Equal(t, []string{"test@example.com"}, sentEmail.To)
	})
	
	t.Run("validation errors", func(t *testing.T) {
		req := &ContactRequest{
			Name:    "",
			Email:   "invalid-email",
			Message: "Hi", // Too short
		}
		
		r := httptest.NewRequest("POST", "/contact", nil)
		
		_, err := svc.ProcessContactForm(ctx, req, r)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
		
		// Check no message was saved
		assert.Len(t, storage.messages, 1) // Still just the one from previous test
	})
	
	t.Run("get message", func(t *testing.T) {
		// First save a message
		testMsg := &git.Message{
			ID:        "test-123",
			Name:      "Test User",
			Email:     "test@example.com",
			Message:   "Test message content",
			Status:    "new",
			Timestamp: time.Now().UTC(),
		}
		storage.messages["test-123"] = testMsg
		
		// Get the message
		msg, err := svc.GetMessage(ctx, "test-123")
		require.NoError(t, err)
		assert.Equal(t, "test-123", msg.ID)
		assert.Equal(t, "Test User", msg.Name)
		assert.Equal(t, "test@example.com", msg.Email)
	})
	
	t.Run("list messages", func(t *testing.T) {
		// Create fresh storage for this test
		storage := newMockStorage()
		eventBus := &mockEventBus{}
		emailService := &mockEmailService{}
		logger := log.New(os.Stdout, "[contact-test] ", log.LstdFlags)
		svc := NewService(storage, eventBus, emailService, "test@example.com", logger)
		
		// Add test messages
		storage.messages["msg-1"] = &git.Message{
			ID:     "msg-1",
			Name:   "User 1",
			Status: "new",
		}
		storage.messages["msg-2"] = &git.Message{
			ID:     "msg-2",
			Name:   "User 2",
			Status: "read",
		}
		storage.messages["msg-3"] = &git.Message{
			ID:     "msg-3",
			Name:   "User 3",
			Status: "replied",
		}
		storage.messages["msg-4"] = &git.Message{
			ID:     "msg-4",
			Name:   "User 4",
			Status: "new",
		}
		
		// List all messages
		messages, err := svc.ListMessages(ctx, "")
		require.NoError(t, err)
		assert.Len(t, messages, 4)
		
		// List by status
		newMessages, err := svc.ListMessages(ctx, "new")
		require.NoError(t, err)
		assert.Len(t, newMessages, 2)
		
		readMessages, err := svc.ListMessages(ctx, "read")
		require.NoError(t, err)
		assert.Len(t, readMessages, 1)
	})
	
	t.Run("update message status", func(t *testing.T) {
		// Add a message to update
		storage.messages["update-test"] = &git.Message{
			ID:     "update-test",
			Name:   "Update Test",
			Status: "new",
		}
		
		// Update status
		err := svc.UpdateMessageStatus(ctx, "update-test", "read")
		require.NoError(t, err)
		
		// Verify status was updated
		assert.Equal(t, "read", storage.messages["update-test"].Status)
		
		// Invalid status
		err = svc.UpdateMessageStatus(ctx, "update-test", "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*http.Request)
		expected string
	}{
		{
			name: "X-Forwarded-For single IP",
			setup: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "10.0.0.1")
			},
			expected: "10.0.0.1",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			setup: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "10.0.0.1, 192.168.1.1, 172.16.0.1")
			},
			expected: "10.0.0.1",
		},
		{
			name: "X-Real-IP",
			setup: func(r *http.Request) {
				r.Header.Set("X-Real-IP", "10.0.0.2")
			},
			expected: "10.0.0.2",
		},
		{
			name: "RemoteAddr with port",
			setup: func(r *http.Request) {
				r.RemoteAddr = "192.168.1.100:12345"
			},
			expected: "192.168.1.100",
		},
		{
			name: "RemoteAddr without port",
			setup: func(r *http.Request) {
				r.RemoteAddr = "192.168.1.100"
			},
			expected: "192.168.1.100",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/test", nil)
			tt.setup(r)
			
			ip := extractIP(r)
			assert.Equal(t, tt.expected, ip)
		})
	}
}