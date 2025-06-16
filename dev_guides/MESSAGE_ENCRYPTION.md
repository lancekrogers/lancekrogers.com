# Message Encryption Documentation

## Overview

Contact form messages are encrypted using **AES-256-GCM** (Galois/Counter Mode) symmetric encryption before being stored in Git. This provides both confidentiality and authenticity.

## Encryption Details

### Algorithm: AES-256-GCM
- **Type**: Symmetric encryption (same key for encrypt/decrypt)
- **Key Size**: 256 bits (32 bytes)
- **Mode**: GCM (Galois/Counter Mode)
- **Nonce**: 12 bytes, randomly generated for each message
- **Authentication**: Built-in authentication tag prevents tampering

### Key Format
- **Raw**: 32 bytes of random data
- **Storage Format**: 64 character hexadecimal string
- **Example**: `a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd`

## Key Management

### 1. Generate a New Key
```bash
# Using make command
make generate-key

# Or directly
go run cmd/generate-key/main.go
```

This will output:
```
Generated new encryption key:
============================
Hex format: a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd

Usage:
1. Add to .env file:
   MESSAGE_ENCRYPTION_KEY=a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd
```

### 2. Configure the Key

Add to your `.env` file:
```bash
# Message Encryption Key (64 hex characters = 32 bytes)
MESSAGE_ENCRYPTION_KEY=your-64-character-hex-key-here
GIT_ENCRYPTION_KEY=same-key-as-above  # Alternative name
```

### 3. Security Best Practices

**DO:**
- ✅ Generate a cryptographically secure random key
- ✅ Store the key in a password manager
- ✅ Use environment variables or secure key storage
- ✅ Keep a secure backup of the key
- ✅ Use different keys for dev/staging/production

**DON'T:**
- ❌ Commit the key to version control
- ❌ Use a weak or predictable key
- ❌ Share the key via insecure channels
- ❌ Lose the key (messages cannot be decrypted without it)

## Using the Decryption Tools

### List Messages
```bash
# List all messages
make messages

# List only new messages
make messages-new

# With custom repo path
go run cmd/decrypt-messages/main.go -repo /path/to/messages
```

### Read a Specific Message
```bash
# Using make
make message-read ID=msg_abc123

# Direct command
go run cmd/decrypt-messages/main.go -id msg_abc123
```

### Update Message Status
```bash
# Mark as read
make message-status ID=msg_abc123 STATUS=read

# Mark as replied
make message-status ID=msg_abc123 STATUS=replied
```

## Message Storage Format

### Encrypted File Structure
```json
{
  "id": "msg_abc123",
  "ciphertext": "base64-encoded-encrypted-data",
  "nonce": "base64-encoded-12-byte-nonce",
  "created_at": "2024-01-15T10:30:00Z",
  "version": "1.0"
}
```

### Decrypted Message Structure
```json
{
  "id": "msg_abc123",
  "name": "John Doe",
  "email": "john@example.com",
  "company": "ACME Corp",
  "message": "I'd like to discuss a project...",
  "ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "timestamp": "2024-01-15T10:30:00Z",
  "status": "new"
}
```

## Troubleshooting

### "Encryption key required" Error
- Ensure `MESSAGE_ENCRYPTION_KEY` is set in `.env` or environment
- Key must be exactly 64 hexadecimal characters

### "Invalid hex key" Error
- Key must contain only 0-9 and a-f characters
- Check for spaces or special characters

### "Failed to decrypt message" Error
- Ensure you're using the same key that encrypted the message
- Check if the message file is corrupted
- Verify the nonce and ciphertext are intact

## Technical Implementation

### Encryption Process
1. Generate random 12-byte nonce
2. Create AES-256-GCM cipher with key
3. Encrypt message JSON with nonce
4. Base64 encode ciphertext and nonce
5. Store in Git with metadata

### Decryption Process
1. Read encrypted file from Git
2. Base64 decode ciphertext and nonce
3. Create AES-256-GCM cipher with key
4. Decrypt and authenticate ciphertext
5. Parse JSON to message struct

## Key Rotation

If you need to rotate keys:

1. Generate new key
2. Decrypt all messages with old key
3. Re-encrypt with new key
4. Update `.env` with new key
5. Securely destroy old key

Note: There's currently no automated key rotation tool.

## Recovery

**If you lose the encryption key:**
- Messages CANNOT be decrypted
- There is no backdoor or recovery method
- This is by design for security

**Backup Strategy:**
1. Store key in password manager
2. Keep encrypted backup in secure location
3. Document key in secure ops documentation
4. Consider key escrow for business continuity