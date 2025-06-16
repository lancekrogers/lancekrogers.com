# Git Storage System Guide

## Overview

The Blockhead Consulting website uses an encrypted git-based storage system for contact form messages. This ensures secure, version-controlled, and distributed storage of sensitive customer data.

## How It Works

### 🔄 Contact Form Flow

1. **User submits contact form** → `POST /contact`
2. **Contact service validates** → Form data validation
3. **Git storage encrypts** → AES-256-GCM encryption
4. **Git operations** → Auto-commit and optional push
5. **Email notification** → Notification sent via MailHog/SMTP

### 📁 Repository Structure

```
data/messages/
├── README.md                    # Repository documentation
├── messages/                    # Encrypted messages directory
│   └── 2025/                   # Year-based organization
│       └── 05/                 # Month-based organization
│           ├── 2025-05-28_19-12-11_msg_908ed6b19364d5f0.json.enc
│           └── 2025-05-28_19-20-19_msg_48df7e955d781384.json.enc
```

### 🔐 Encryption Details

- **Algorithm**: AES-256-GCM (Galois/Counter Mode)
- **Key Source**: Environment variable `GIT_ENCRYPTION_KEY` (32 bytes)
- **Nonce**: Cryptographically secure random nonce per message
- **Content**: All form data (name, email, company, message) encrypted

### 📧 Git Configuration

```bash
# Repository settings
GIT_REPO_PATH=./data/messages
GIT_BRANCH=main
GIT_COMMIT_AUTHOR=Blockhead Dev Bot
GIT_COMMIT_EMAIL=dev@blockhead.consulting

# Security settings
GIT_ENCRYPTION_KEY=0123456789abcdef0123456789abcdef  # 32 bytes for AES-256
GIT_PUSH_ON_WRITE=true  # Auto-push to remote after commit

# Remote repository
git remote add origin git@github.com:Blockhead-Consulting/website-messages.git
```

## Development Setup

### Quick Start

```bash
# Start complete development environment
make dev-bg

# This automatically:
# 1. Starts MailHog email testing service
# 2. Starts website with git storage enabled
# 3. Provides status dashboard with all service URLs

# Stop everything
make stop-dev
```

### Available Commands

```bash
# Development services
make dev-services    # Start MailHog and other dev services
make dev            # Start website with all services (foreground)
make dev-bg         # Start website with all services (background)
make stop-dev       # Stop all development services

# Individual services
make start-mailhog  # Start MailHog email testing
make stop-mailhog   # Stop MailHog
make check-docker   # Verify Docker is running

# Server management
make status         # Check server and service status
make logs          # Show server logs (if running in background)
make stop          # Stop website server only
```

### Service URLs

- **Website**: <http://localhost:8087> (port from .env)
- **MailHog UI**: <http://localhost:8025>
- **MailHog SMTP**: localhost:1025

## Message Storage

### Encrypted Message Format

```json
{
  "id": "msg_908ed6b19364d5f0",
  "ciphertext": "2I5nDOkCune6nSdBZj2EzM3y+9KHBPRRyWod9fGp9W6zlwV5...",
  "nonce": "lBZ45/V9RUMRis4/",
  "created_at": "2025-05-28T19:12:11Z",
  "version": "1.0"
}
```

### Git Commit Messages

```bash
Add message msg_908ed6b19364d5f0 from Test User
Add message msg_48df7e955d781384 from Test User
```

### Viewing Messages

```bash
# Check git status
cd data/messages && git status

# View commit history
cd data/messages && git log --oneline

# Push to remote manually
cd data/messages && git push

# View file structure
ls -la data/messages/messages/2025/05/
```

## Security Features

### ✅ What's Secure

- **AES-256-GCM encryption** for all message content
- **Unique nonces** prevent replay attacks
- **Version-controlled storage** with full audit trail
- **Remote backup** to private GitHub repository
- **Environment-based keys** (not committed to code)
- **Structured access control** via git permissions

### ⚠️ Important Security Notes

- **Encryption key** must be exactly 32 bytes for AES-256
- **Never commit** encryption keys to any repository
- **Remote repository** should be private with restricted access
- **Production keys** should be different from development keys
- **Key rotation** should be implemented for production use

## Troubleshooting

### Common Issues

**"encryption key must be 32 bytes"**

```bash
# Check key length
echo "0123456789abcdef0123456789abcdef" | wc -c  # Should be 33 (includes newline)
# The key itself should be exactly 32 characters
```

**"Git storage service initialization failed"**

```bash
# Check .env file has all required variables
grep GIT_ .env

# Check git repository exists
ls -la data/messages/.git

# Initialize if needed
cd data/messages && git init
```

**"Docker daemon is not running"**

```bash
# Start Docker Desktop on macOS
open -a Docker

# Or check if running
docker info
```

**Messages not appearing in repository**

```bash
# Check server logs
tail -f data/server.log

# Test contact form manually
curl -X POST http://localhost:8087/contact \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "name=Test&email=test@example.com&message=Test message"
```

## Production Deployment

### Environment Variables

```bash
# Production git storage
GIT_REPO_PATH=/app/data/messages
GIT_ENCRYPTION_KEY=<secure-32-byte-production-key>
GIT_REMOTE_URL=git@github.com:Blockhead-Consulting/website-messages.git
GIT_PUSH_ON_WRITE=true

# Production email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=<gmail-username>
SMTP_PASSWORD=<gmail-app-password>
SMTP_FROM_ADDRESS=<your-email>
SMTP_TLS_ENABLED=true
```

### Security Checklist

- [ ] Generate secure 32-byte encryption key
- [ ] Set up private remote repository
- [ ] Configure SSH keys for git access
- [ ] Set up production SMTP credentials
- [ ] Test message encryption/decryption
- [ ] Verify git push functionality
- [ ] Set up monitoring for failed messages
- [ ] Plan key rotation strategy

## Content as Submodule (Optional)

The `content/` directory can optionally be a git submodule if you want to:

- Share content across multiple environments
- Separate content management from code changes
- Have different access controls for content vs code

**To set up content as submodule:**

```bash
# Remove existing content directory
mv content content-backup

# Add as submodule
git submodule add git@github.com:Blockhead-Consulting/website-content.git content

# Update submodule
git submodule update --remote content
```

**Pros**: Separate versioning, shared content, granular access
**Cons**: Added complexity, requires submodule knowledge
**Recommendation**: Not necessary for single-site deployment

