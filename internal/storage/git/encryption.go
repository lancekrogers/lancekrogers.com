package git

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"blockhead.consulting/internal/errors"
)

// Encryptor handles message encryption/decryption
type Encryptor struct {
	key []byte
}

// NewEncryptor creates a new encryptor with the given key
func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != 32 {
		return nil, errors.New(errors.ErrCodeValidation, "encryption key must be 32 bytes for AES-256")
	}
	
	return &Encryptor{
		key: key,
	}, nil
}

// Encrypt encrypts a message using AES-256-GCM
func (e *Encryptor) Encrypt(message *Message) (*EncryptedMessage, error) {
	// Marshal message to JSON
	plaintext, err := json.Marshal(message)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeSerialization, "failed to marshal message")
	}
	
	// Create cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to create cipher")
	}
	
	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to create GCM")
	}
	
	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to generate nonce")
	}
	
	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	
	// Create encrypted message
	encrypted := &EncryptedMessage{
		ID:         message.ID,
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		CreatedAt:  message.Timestamp.Format(time.RFC3339),
		Version:    "1.0",
	}
	
	return encrypted, nil
}

// Decrypt decrypts an encrypted message
func (e *Encryptor) Decrypt(encrypted *EncryptedMessage) (*Message, error) {
	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted.Ciphertext)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInvalidFormat, "failed to decode ciphertext")
	}
	
	nonce, err := base64.StdEncoding.DecodeString(encrypted.Nonce)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInvalidFormat, "failed to decode nonce")
	}
	
	// Create cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to create cipher")
	}
	
	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to create GCM")
	}
	
	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDecryption, "failed to decrypt message")
	}
	
	// Unmarshal message
	var message Message
	if err := json.Unmarshal(plaintext, &message); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeSerialization, "failed to unmarshal message")
	}
	
	return &message, nil
}

// GenerateKey generates a new 32-byte encryption key
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}