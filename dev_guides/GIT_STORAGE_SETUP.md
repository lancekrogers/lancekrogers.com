# Git Storage Setup Guide

## Overview
Contact form messages are encrypted and stored in a Git repository. This guide explains how to set up automatic pushing to a remote repository.

## Prerequisites
1. A private Git repository (GitHub, GitLab, etc.)
2. Git credentials configured for push access
3. The encryption key configured in `.env`

## Setup Steps

### 1. Create Private Repository
Create a **private** repository on your Git hosting service:
- GitHub: https://github.com/new (select "Private")
- GitLab: Create new project with "Private" visibility
- Bitbucket: Create repository with "Private" access

**Important**: Use a PRIVATE repository to protect customer data!

### 2. Configure Git Remote
```bash
# Navigate to messages directory
cd data/messages

# Add remote (if not already added)
git remote add origin git@github.com:yourusername/your-private-messages.git

# Verify remote is configured
git remote -v
```

### 3. Configure Authentication

#### Option A: SSH Key (Recommended)
```bash
# Check if you have SSH key
ls ~/.ssh/id_rsa.pub

# If not, generate one
ssh-keygen -t rsa -b 4096 -C "your-email@example.com"

# Add to SSH agent
ssh-add ~/.ssh/id_rsa

# Copy public key and add to GitHub/GitLab
cat ~/.ssh/id_rsa.pub
```

#### Option B: HTTPS with Token
```bash
# For GitHub, use Personal Access Token
git remote set-url origin https://your-token@github.com/username/repo.git
```

### 4. Update Environment Configuration
Edit `.env` file:
```bash
# Git Storage Configuration
GIT_REPO_PATH=./data/messages
GIT_ENCRYPTION_KEY=your-64-character-hex-key
GIT_REMOTE_URL=origin              # Remote name
GIT_BRANCH=main                    # Branch to push to
GIT_COMMIT_AUTHOR=Your Bot Name    
GIT_COMMIT_EMAIL=bot@yourdomain.com
GIT_PUSH_ON_WRITE=true            # Enable automatic push
```

### 5. Test Push Access
```bash
cd data/messages

# Create test file
echo "test" > test.txt
git add test.txt
git commit -m "Test push access"
git push origin main

# If successful, remove test
git rm test.txt
git commit -m "Remove test file"
git push origin main
```

## Troubleshooting

### Push Permission Denied
```
Permission denied (publickey).
fatal: Could not read from remote repository.
```
**Solution**: Configure SSH key or HTTPS token (see Step 3)

### Remote Not Found
```
fatal: 'origin' does not appear to be a git repository
```
**Solution**: Add remote with `git remote add origin <url>`

### Branch Doesn't Exist
```
error: src refspec main does not match any
```
**Solution**: 
```bash
git checkout -b main
git push -u origin main
```

### Authentication Required for Every Push
**Solution for SSH**:
```bash
# Add SSH key to agent
ssh-add ~/.ssh/id_rsa

# For permanent fix, add to ~/.ssh/config:
Host github.com
  AddKeysToAgent yes
  UseKeychain yes
  IdentityFile ~/.ssh/id_rsa
```

## Production Deployment

### 1. Use Deploy Keys (Recommended)
For production, use repository-specific deploy keys:
```bash
# Generate deploy key
ssh-keygen -t ed25519 -C "blockhead-website-prod" -f ~/.ssh/blockhead_deploy

# Add to repository settings as deploy key with write access
cat ~/.ssh/blockhead_deploy.pub
```

### 2. Configure Server
On production server:
```bash
# Add to ~/.ssh/config
Host github.com
  HostName github.com
  User git
  IdentityFile ~/.ssh/blockhead_deploy
  IdentitiesOnly yes
```

### 3. Use Environment Variables
Set in production environment:
```bash
GIT_ENCRYPTION_KEY=production-key-from-secure-storage
GIT_PUSH_ON_WRITE=true
GIT_COMMIT_AUTHOR=Blockhead Production
GIT_COMMIT_EMAIL=noreply@blockhead.consulting
```

## Security Considerations

1. **Private Repository**: Always use private repositories
2. **Encryption Key**: Never commit the encryption key
3. **Access Control**: Use deploy keys or machine users
4. **Audit Trail**: Git provides complete audit history
5. **Backup**: Regular backups of the Git repository

## Monitoring

Check push status in logs:
```bash
# Check if pushes are working
tail -f data/server.log | grep "GIT:"

# Successful push:
# GIT: Successfully pushed to remote

# Failed push:
# GIT: Push failed: <error>
```

## Manual Operations

If automatic push fails, manually push:
```bash
cd data/messages
git status
git push origin main
```

To pull messages from another machine:
```bash
git pull origin main
make messages  # List all messages
```