package git

import (
	"time"
)

// Message represents a contact form message
type Message struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Company   string    `json:"company,omitempty"`
	Message   string    `json:"message"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // "new", "read", "replied"
}

// EncryptedMessage represents an encrypted message stored in Git
type EncryptedMessage struct {
	ID         string `json:"id"`
	Ciphertext string `json:"ciphertext"` // Base64 encoded
	Nonce      string `json:"nonce"`      // Base64 encoded
	CreatedAt  string `json:"created_at"`
	Version    string `json:"version"`    // Encryption version
}

// StorageConfig holds configuration for Git storage
type StorageConfig struct {
	RepoPath       string
	RemoteURL      string
	Branch         string
	CommitAuthor   string
	CommitEmail    string
	EncryptionKey  string // 32-byte key for AES-256
	PushOnWrite    bool   // Whether to push after each write
}