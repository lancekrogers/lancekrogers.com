package git

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"blockhead.consulting/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEventBus struct {
	publishedEvents []events.Event
}

func (m *mockEventBus) Subscribe(eventType events.EventType, handler events.EventHandler) string {
	return "mock-subscription-id"
}

func (m *mockEventBus) Unsubscribe(subscriptionID string) {
}

func (m *mockEventBus) Publish(ctx context.Context, event events.Event) error {
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *mockEventBus) Start(ctx context.Context) error {
	return nil
}

func (m *mockEventBus) Stop() error {
	return nil
}

func createTestService(t *testing.T) (Service, *mockEventBus, string) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "git-storage-test-*")
	require.NoError(t, err)
	
	// Generate test key
	key, err := GenerateKey()
	require.NoError(t, err)
	
	// Create config
	config := StorageConfig{
		RepoPath:      tempDir,
		Branch:        "main",
		CommitAuthor:  "Test User",
		CommitEmail:   "test@example.com",
		EncryptionKey: string(key),
		PushOnWrite:   false, // Don't push in tests
	}
	
	// Create mock event bus
	mockBus := &mockEventBus{}
	
	// Create logger
	logger := log.New(os.Stdout, "[git-test] ", log.LstdFlags)
	
	// Create service
	svc, err := NewService(config, logger, mockBus)
	require.NoError(t, err)
	
	return svc, mockBus, tempDir
}

func TestGitStorageService(t *testing.T) {
	svc, mockBus, tempDir := createTestService(t)
	defer os.RemoveAll(tempDir)
	
	ctx := context.Background()
	
	// Create test message
	message := &Message{
		ID:        "test-msg-123",
		Name:      "Jane Doe",
		Email:     "jane@example.com",
		Company:   "Test Corp",
		Message:   "This is a test message",
		IP:        "10.0.0.1",
		UserAgent: "Test/1.0",
		Timestamp: time.Now().UTC(),
		Status:    "new",
	}
	
	t.Run("save message", func(t *testing.T) {
		err := svc.SaveMessage(ctx, message)
		assert.NoError(t, err)
		
		// Check event was published
		assert.Len(t, mockBus.publishedEvents, 1)
		event := mockBus.publishedEvents[0]
		assert.Equal(t, events.EventMessageStored, event.Type())
		
		// Check file was created
		year := message.Timestamp.Format("2006")
		month := message.Timestamp.Format("01")
		pattern := filepath.Join(tempDir, "messages", year, month, "*test-msg-123*.json.enc")
		matches, err := filepath.Glob(pattern)
		require.NoError(t, err)
		assert.Len(t, matches, 1)
	})
	
	t.Run("get message", func(t *testing.T) {
		retrieved, err := svc.GetMessage(ctx, "test-msg-123")
		require.NoError(t, err)
		assert.Equal(t, message.ID, retrieved.ID)
		assert.Equal(t, message.Name, retrieved.Name)
		assert.Equal(t, message.Email, retrieved.Email)
		assert.Equal(t, message.Message, retrieved.Message)
		assert.Equal(t, message.Status, retrieved.Status)
	})
	
	t.Run("list messages", func(t *testing.T) {
		// Save another message
		message2 := &Message{
			ID:        "test-msg-456",
			Name:      "Bob Smith",
			Email:     "bob@example.com",
			Message:   "Another test",
			IP:        "10.0.0.2",
			UserAgent: "Test/2.0",
			Timestamp: time.Now().Add(time.Hour).UTC(),
			Status:    "read",
		}
		err := svc.SaveMessage(ctx, message2)
		require.NoError(t, err)
		
		// List all messages
		messages, err := svc.ListMessages(ctx, "")
		require.NoError(t, err)
		assert.Len(t, messages, 2)
		
		// Should be sorted newest first
		assert.Equal(t, "test-msg-456", messages[0].ID)
		assert.Equal(t, "test-msg-123", messages[1].ID)
		
		// List by status
		newMessages, err := svc.ListMessages(ctx, "new")
		require.NoError(t, err)
		assert.Len(t, newMessages, 1)
		assert.Equal(t, "test-msg-123", newMessages[0].ID)
		
		readMessages, err := svc.ListMessages(ctx, "read")
		require.NoError(t, err)
		assert.Len(t, readMessages, 1)
		assert.Equal(t, "test-msg-456", readMessages[0].ID)
	})
	
	t.Run("update status", func(t *testing.T) {
		err := svc.UpdateStatus(ctx, "test-msg-123", "replied")
		require.NoError(t, err)
		
		// Verify status was updated
		retrieved, err := svc.GetMessage(ctx, "test-msg-123")
		require.NoError(t, err)
		assert.Equal(t, "replied", retrieved.Status)
	})
	
	t.Run("message not found", func(t *testing.T) {
		_, err := svc.GetMessage(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestGitRepository(t *testing.T) {
	_, _, tempDir := createTestService(t)
	defer os.RemoveAll(tempDir)
	
	t.Run("repository initialized", func(t *testing.T) {
		// Check .git directory exists
		gitDir := filepath.Join(tempDir, ".git")
		info, err := os.Stat(gitDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
		
		// Check README exists
		readme := filepath.Join(tempDir, "README.md")
		_, err = os.Stat(readme)
		assert.NoError(t, err)
		
		// Check messages directory exists
		messagesDir := filepath.Join(tempDir, "messages")
		_, err = os.Stat(messagesDir)
		assert.NoError(t, err)
	})
}

func TestServiceValidation(t *testing.T) {
	logger := log.New(os.Stdout, "[test] ", log.LstdFlags)
	mockBus := &mockEventBus{}
	
	t.Run("missing repo path", func(t *testing.T) {
		config := StorageConfig{
			EncryptionKey: string(make([]byte, 32)),
		}
		_, err := NewService(config, logger, mockBus)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repo path")
	})
	
	t.Run("missing encryption key", func(t *testing.T) {
		config := StorageConfig{
			RepoPath: "/tmp/test",
		}
		_, err := NewService(config, logger, mockBus)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "encryption key")
	})
}