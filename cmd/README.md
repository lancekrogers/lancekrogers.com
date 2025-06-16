# Message Management Tools

This directory contains command-line tools for managing encrypted contact form messages.

## Tools

### 1. Generate Encryption Key
Generates a new 32-byte encryption key for message encryption.

```bash
go run cmd/generate-key/main.go
```

Save the generated key in your `.env` file:
```
MESSAGE_ENCRYPTION_KEY=<your-64-char-hex-key>
```

### 2. Decrypt and Read Messages
Lists and decrypts messages from the Git repository.

```bash
# List all messages
go run cmd/decrypt-messages/main.go

# List only new messages
go run cmd/decrypt-messages/main.go -status new

# Read a specific message
go run cmd/decrypt-messages/main.go -id msg_abc123

# Use a specific repo path
go run cmd/decrypt-messages/main.go -repo /path/to/messages
```

### 3. Update Message Status
Changes the status of a message (new, read, replied, closed).

```bash
# Mark as read
go run cmd/message-status/main.go -id msg_abc123 -status read

# Mark as replied and push to remote
go run cmd/message-status/main.go -id msg_abc123 -status replied -push
```

## Environment Variables

You can set these environment variables instead of using flags:

- `MESSAGE_ENCRYPTION_KEY`: The encryption key in hex format
- `MESSAGE_REPO_PATH`: Path to the messages repository (default: data/messages)

## Security Notes

1. **Never commit the encryption key** to version control
2. Store the key securely (e.g., in a password manager)
3. The same key must be used for encryption and decryption
4. Without the key, messages cannot be decrypted

## Git Repository Setup

For automatic push to work, you need to:

1. Initialize a private Git repository for messages
2. Configure Git credentials for push access
3. Set up the remote URL in the storage configuration

Example:
```bash
cd data/messages
git remote add origin git@github.com:yourusername/private-messages.git
```