package git

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"blockhead.consulting/internal/errors"
	"blockhead.consulting/internal/events"
)

// Service provides Git-based message storage
type Service interface {
	// SaveMessage saves an encrypted message to the Git repository
	SaveMessage(ctx context.Context, message *Message) error
	
	// GetMessage retrieves a message by ID
	GetMessage(ctx context.Context, id string) (*Message, error)
	
	// ListMessages lists all messages with optional filters
	ListMessages(ctx context.Context, status string) ([]*Message, error)
	
	// UpdateStatus updates the status of a message
	UpdateStatus(ctx context.Context, id string, status string) error
}

// service implements the Git storage service
type service struct {
	config    StorageConfig
	encryptor *Encryptor
	mu        sync.Mutex
	logger    *log.Logger
	eventBus  events.EventBus
}

// NewService creates a new Git storage service
func NewService(config StorageConfig, logger *log.Logger, eventBus events.EventBus) (Service, error) {
	// Validate config
	if config.RepoPath == "" {
		return nil, errors.New(errors.ErrCodeValidation, "repo path is required")
	}
	
	if config.EncryptionKey == "" {
		return nil, errors.New(errors.ErrCodeValidation, "encryption key is required")
	}
	
	// Create encryptor
	encryptor, err := NewEncryptor([]byte(config.EncryptionKey))
	if err != nil {
		return nil, err
	}
	
	// Create service
	svc := &service{
		config:    config,
		encryptor: encryptor,
		logger:    logger,
		eventBus:  eventBus,
	}
	
	// Initialize repository if needed
	if err := svc.initRepository(context.Background()); err != nil {
		return nil, err
	}
	
	return svc, nil
}

// initRepository initializes the Git repository if it doesn't exist
func (s *service) initRepository(ctx context.Context) error {
	// Check if repo exists
	if _, err := os.Stat(filepath.Join(s.config.RepoPath, ".git")); err == nil {
		s.logger.Printf("GIT: Repository already exists at %s", s.config.RepoPath)
		return nil
	}
	
	// Create directory
	if err := os.MkdirAll(s.config.RepoPath, 0755); err != nil {
		return errors.Wrap(err, errors.ErrCodeIO, "failed to create repo directory")
	}
	
	// Initialize Git repo
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = s.config.RepoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrap(err, errors.ErrCodeIO, fmt.Sprintf("failed to init git repo: %s", output))
	}
	
	// Configure Git
	s.gitConfig(ctx, "user.name", s.config.CommitAuthor)
	s.gitConfig(ctx, "user.email", s.config.CommitEmail)
	
	// Create README
	readme := filepath.Join(s.config.RepoPath, "README.md")
	content := `# Encrypted Messages Repository

This repository contains encrypted contact form messages.
DO NOT commit unencrypted messages or encryption keys.
`
	if err := os.WriteFile(readme, []byte(content), 0644); err != nil {
		return errors.Wrap(err, errors.ErrCodeIO, "failed to create README")
	}
	
	// Create messages directory
	messagesDir := filepath.Join(s.config.RepoPath, "messages")
	if err := os.MkdirAll(messagesDir, 0755); err != nil {
		return errors.Wrap(err, errors.ErrCodeIO, "failed to create messages directory")
	}
	
	// Initial commit
	s.gitAdd(ctx, ".")
	s.gitCommit(ctx, "Initial commit")
	
	// Add remote if configured
	if s.config.RemoteURL != "" {
		s.gitRemote(ctx, "add", "origin", s.config.RemoteURL)
	}
	
	s.logger.Printf("GIT: Repository initialized at %s", s.config.RepoPath)
	return nil
}

// SaveMessage saves an encrypted message to the Git repository
func (s *service) SaveMessage(ctx context.Context, message *Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Encrypt message
	encrypted, err := s.encryptor.Encrypt(message)
	if err != nil {
		return err
	}
	
	// Create directory structure (year/month)
	year := message.Timestamp.Format("2006")
	month := message.Timestamp.Format("01")
	dir := filepath.Join(s.config.RepoPath, "messages", year, month)
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, errors.ErrCodeIO, "failed to create directory")
	}
	
	// Create filename
	filename := fmt.Sprintf("%s_%s.json.enc",
		message.Timestamp.Format("2006-01-02_15-04-05"),
		message.ID,
	)
	filePath := filepath.Join(dir, filename)
	
	// Marshal encrypted message
	data, err := json.MarshalIndent(encrypted, "", "  ")
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeSerialization, "failed to marshal encrypted message")
	}
	
	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return errors.Wrap(err, errors.ErrCodeIO, "failed to write encrypted message")
	}
	
	// Git operations
	relPath, _ := filepath.Rel(s.config.RepoPath, filePath)
	if err := s.gitAdd(ctx, relPath); err != nil {
		return err
	}
	
	commitMsg := fmt.Sprintf("Add message %s from %s", message.ID, message.Name)
	if err := s.gitCommit(ctx, commitMsg); err != nil {
		return err
	}
	
	// Push if configured
	if s.config.PushOnWrite && s.config.RemoteURL != "" {
		go s.gitPush(context.Background()) // Async push
	}
	
	// Publish event
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, events.NewEventWithContext(ctx,
			events.EventMessageStored,
			map[string]interface{}{
				"message_id": message.ID,
				"path":       relPath,
			},
		))
	}
	
	s.logger.Printf("GIT: Saved message %s to %s", message.ID, relPath)
	return nil
}

// GetMessage retrieves a message by ID
func (s *service) GetMessage(ctx context.Context, id string) (*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Search for file containing the ID
	messagesDir := filepath.Join(s.config.RepoPath, "messages")
	var foundPath string
	
	err := filepath.Walk(messagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if strings.Contains(info.Name(), id) && strings.HasSuffix(info.Name(), ".json.enc") {
			foundPath = path
			return filepath.SkipDir
		}
		
		return nil
	})
	
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeIO, "failed to search for message")
	}
	
	if foundPath == "" {
		return nil, errors.NotFound("message")
	}
	
	// Read file
	data, err := os.ReadFile(foundPath)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeIO, "failed to read message file")
	}
	
	// Unmarshal encrypted message
	var encrypted EncryptedMessage
	if err := json.Unmarshal(data, &encrypted); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeSerialization, "failed to unmarshal encrypted message")
	}
	
	// Decrypt
	message, err := s.encryptor.Decrypt(&encrypted)
	if err != nil {
		return nil, err
	}
	
	return message, nil
}

// ListMessages lists all messages with optional status filter
func (s *service) ListMessages(ctx context.Context, status string) ([]*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	var messages []*Message
	messagesDir := filepath.Join(s.config.RepoPath, "messages")
	
	err := filepath.Walk(messagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(info.Name(), ".json.enc") {
			return nil
		}
		
		// Read and decrypt each message
		data, err := os.ReadFile(path)
		if err != nil {
			s.logger.Printf("GIT: Warning - failed to read %s: %v", path, err)
			return nil
		}
		
		var encrypted EncryptedMessage
		if err := json.Unmarshal(data, &encrypted); err != nil {
			s.logger.Printf("GIT: Warning - failed to unmarshal %s: %v", path, err)
			return nil
		}
		
		message, err := s.encryptor.Decrypt(&encrypted)
		if err != nil {
			s.logger.Printf("GIT: Warning - failed to decrypt %s: %v", path, err)
			return nil
		}
		
		// Filter by status if specified
		if status == "" || message.Status == status {
			messages = append(messages, message)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeIO, "failed to list messages")
	}
	
	// Sort by timestamp (newest first)
	for i := 0; i < len(messages)-1; i++ {
		for j := i + 1; j < len(messages); j++ {
			if messages[i].Timestamp.Before(messages[j].Timestamp) {
				messages[i], messages[j] = messages[j], messages[i]
			}
		}
	}
	
	return messages, nil
}

// UpdateStatus updates the status of a message
func (s *service) UpdateStatus(ctx context.Context, id string, status string) error {
	// Get the message
	message, err := s.GetMessage(ctx, id)
	if err != nil {
		return err
	}
	
	// Update status
	message.Status = status
	
	// Save updated message
	return s.SaveMessage(ctx, message)
}

// Git helper methods

func (s *service) gitConfig(ctx context.Context, key, value string) error {
	cmd := exec.CommandContext(ctx, "git", "config", key, value)
	cmd.Dir = s.config.RepoPath
	return cmd.Run()
}

func (s *service) gitAdd(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "git", "add", path)
	cmd.Dir = s.config.RepoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrap(err, errors.ErrCodeIO, fmt.Sprintf("git add failed: %s", output))
	}
	return nil
}

func (s *service) gitCommit(ctx context.Context, message string) error {
	cmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.Dir = s.config.RepoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrap(err, errors.ErrCodeIO, fmt.Sprintf("git commit failed: %s", output))
	}
	return nil
}

func (s *service) gitPush(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "push", "origin", s.config.Branch)
	cmd.Dir = s.config.RepoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		s.logger.Printf("GIT: Push failed: %s", output)
		return errors.Wrap(err, errors.ErrCodeIO, "git push failed")
	}
	s.logger.Printf("GIT: Successfully pushed to remote")
	return nil
}

func (s *service) gitRemote(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", append([]string{"remote"}, args...)...)
	cmd.Dir = s.config.RepoPath
	return cmd.Run()
}