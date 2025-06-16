package git

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptor(t *testing.T) {
	// Generate a test key
	key, err := GenerateKey()
	require.NoError(t, err)
	assert.Len(t, key, 32)
	
	// Create encryptor
	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)
	
	// Create test message
	message := &Message{
		ID:        "test-123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Company:   "ACME Corp",
		Message:   "This is a test message with special chars: 你好世界 🎉",
		IP:        "192.168.1.1",
		UserAgent: "Mozilla/5.0",
		Timestamp: time.Now().UTC(),
		Status:    "new",
	}
	
	t.Run("encrypt and decrypt", func(t *testing.T) {
		// Encrypt
		encrypted, err := encryptor.Encrypt(message)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted.Ciphertext)
		assert.NotEmpty(t, encrypted.Nonce)
		assert.Equal(t, message.ID, encrypted.ID)
		assert.Equal(t, "1.0", encrypted.Version)
		
		// Decrypt
		decrypted, err := encryptor.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, message.ID, decrypted.ID)
		assert.Equal(t, message.Name, decrypted.Name)
		assert.Equal(t, message.Email, decrypted.Email)
		assert.Equal(t, message.Company, decrypted.Company)
		assert.Equal(t, message.Message, decrypted.Message)
		assert.Equal(t, message.IP, decrypted.IP)
		assert.Equal(t, message.UserAgent, decrypted.UserAgent)
		assert.Equal(t, message.Status, decrypted.Status)
		assert.WithinDuration(t, message.Timestamp, decrypted.Timestamp, time.Second)
	})
	
	t.Run("different nonce each time", func(t *testing.T) {
		encrypted1, err := encryptor.Encrypt(message)
		require.NoError(t, err)
		
		encrypted2, err := encryptor.Encrypt(message)
		require.NoError(t, err)
		
		// Same plaintext but different ciphertext due to different nonce
		assert.NotEqual(t, encrypted1.Nonce, encrypted2.Nonce)
		assert.NotEqual(t, encrypted1.Ciphertext, encrypted2.Ciphertext)
	})
	
	t.Run("invalid key length", func(t *testing.T) {
		_, err := NewEncryptor([]byte("too-short"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "32 bytes")
	})
	
	t.Run("tampered ciphertext", func(t *testing.T) {
		encrypted, err := encryptor.Encrypt(message)
		require.NoError(t, err)
		
		// Tamper with ciphertext
		encrypted.Ciphertext = "tampered" + encrypted.Ciphertext[8:]
		
		_, err = encryptor.Decrypt(encrypted)
		assert.Error(t, err)
	})
	
	t.Run("wrong key", func(t *testing.T) {
		encrypted, err := encryptor.Encrypt(message)
		require.NoError(t, err)
		
		// Create encryptor with different key
		wrongKey, _ := GenerateKey()
		wrongEncryptor, _ := NewEncryptor(wrongKey)
		
		_, err = wrongEncryptor.Decrypt(encrypted)
		assert.Error(t, err)
	})
}