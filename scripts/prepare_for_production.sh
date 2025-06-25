#!/bin/bash
# Prepare codebase for production deployment by removing debug logging

echo "🚀 Preparing codebase for production deployment..."

# Function to remove lines containing patterns
remove_lines() {
    local file=$1
    local pattern=$2
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "/$pattern/d" "$file"
    else
        # Linux
        sed -i "/$pattern/d" "$file"
    fi
}

# Create backup
echo "📦 Creating backup..."
backup_dir="backups/pre-production-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$backup_dir"
cp -r static internal/handlers "$backup_dir/"

# Remove DEBUG logs from Go files
echo "🔧 Removing DEBUG logs from Go files..."
find internal/handlers -name "*.go" -type f | while read file; do
    remove_lines "$file" "log\.Printf.*DEBUG:"
done

# Update main.go to set production environment
echo "🔧 Setting production environment in main.go..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' 's/Environment:.*"development"/Environment: "production"/' main.go
else
    sed -i 's/Environment:.*"development"/Environment: "production"/' main.go
fi

# Create environment file template
echo "📝 Creating production environment template..."
cat > .env.production.example << 'EOF'
# Production Environment Configuration
# Copy this to .env and update with your production values

# Server Configuration
PORT=8087
ENVIRONMENT=production

# Email Configuration (Brevo/SendinBlue)
SMTP_HOST=smtp-relay.brevo.com
SMTP_PORT=587
SMTP_USER=your-brevo-smtp-user
SMTP_PASSWORD=your-brevo-smtp-password

# Encryption Key (32-byte key for AES-256)
# Generate with: openssl rand -hex 32
MESSAGE_ENCRYPTION_KEY=your-32-byte-encryption-key-here

# Admin Configuration (if calendar is enabled)
# ADMIN_USERNAME=admin
# ADMIN_PASSWORD=secure-password-here

# GitHub Integration (optional)
GITHUB_USERNAME=lancekrogers
EOF

# Create deployment checklist
echo "📋 Creating deployment checklist..."
cat > DEPLOYMENT_CHECKLIST.md << 'EOF'
# Production Deployment Checklist

## Pre-Deployment Steps ✅

- [ ] Run `./scripts/prepare_for_production.sh`
- [ ] Copy `.env.production.example` to `.env` and update values
- [ ] Set `ENVIRONMENT=production` in environment
- [ ] Generate new `MESSAGE_ENCRYPTION_KEY` with `openssl rand -hex 32`
- [ ] Update SMTP credentials for production email service
- [ ] Remove or secure any test data

## Build Steps

```bash
# Build for production
go build -ldflags="-s -w" -o blockhead-server main.go

# Or use make
make build
```

## Security Verification

- [ ] Verify CSP headers are correct (no unsafe-eval)
- [ ] Check all security headers are present
- [ ] Confirm admin endpoints return 503 without env vars
- [ ] Test rate limiting is working
- [ ] Verify no DEBUG logs in output

## Post-Deployment

- [ ] Monitor server logs for errors
- [ ] Test contact form functionality
- [ ] Verify blog posts load correctly
- [ ] Check that animations work without console errors
- [ ] Confirm GitHub stats are loading (if enabled)

## Rollback Plan

If issues occur, restore from backup:
```bash
cp -r backups/pre-production-*/static/* static/
cp -r backups/pre-production-*/internal/handlers/* internal/handlers/
```
EOF

echo "✅ Production preparation complete!"
echo ""
echo "📋 Next steps:"
echo "1. Review changes with: git diff"
echo "2. Copy .env.production.example to .env and update values"
echo "3. Build with: go build -ldflags=\"-s -w\" -o blockhead-server main.go"
echo "4. Follow DEPLOYMENT_CHECKLIST.md"
echo ""
echo "💾 Backup created in: $backup_dir"