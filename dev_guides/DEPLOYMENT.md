# 🚀 Deployment Guide - Blockhead Consulting Website

## Quick Deployment

### Standard Deployment (Calendar Enabled)
```bash
# Build and run with calendar features
make build
./server
```

### Rapid Deployment (Calendar Disabled)

**Option 1: Using .env file**
```bash
# Create/edit .env file
echo "CALENDAR_ENABLED=false" > .env
make build
./server
```

**Option 2: Using environment variables**
```bash
# Disable calendar for quick deployment
export CALENDAR_ENABLED=false
make build
./server
```

## Environment Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CALENDAR_ENABLED` | `true` | Enable/disable calendar functionality |
| `ENVIRONMENT` | `development` | Runtime environment (development/production) |
| `SITE_NAME` | `Blockhead Consulting` | Site name for branding |
| `PORT` | `8085` | Server port |

### Production Configuration

**Option 1: Using .env file**
```bash
# Create .env file
cat > .env << EOF
CALENDAR_ENABLED=true
ENVIRONMENT=production
SITE_NAME=Blockhead Consulting
PORT=80
EOF
```

**Option 2: Using environment variables**
```bash
export CALENDAR_ENABLED=true
export ENVIRONMENT=production
export SITE_NAME="Blockhead Consulting"
export PORT=80
```

### Development Configuration

**Option 1: Using .env file (Recommended)**
```bash
# Copy example and customize
cp .env.example .env
# Edit .env as needed
```

**Option 2: Using environment variables**
```bash
export CALENDAR_ENABLED=true
export ENVIRONMENT=development
export PORT=8085
```

## Feature Toggles

### Calendar Disable (Secure)
When `CALENDAR_ENABLED=false`:
- Calendar routes are **not registered** (404 on calendar endpoints)
- Calendar navigation links are **completely removed** from templates
- Calendar JavaScript is **not loaded**
- Calendar API endpoints are **not available**
- No traces of calendar functionality in HTML source

**Security Benefit**: Prevents discovery of calendar endpoints by attackers.

### Mobile Navigation
- Automatic hamburger menu on mobile/tablet (≤1024px)
- Desktop navigation unchanged (>1024px)
- Smooth animations and touch-friendly interactions

## Build Commands

### Development
```bash
make dev          # Start development server
make test         # Run all tests
make test-verbose # Detailed test output
```

### Production
```bash
make build        # Build binary
make all          # Full build and test pipeline
```

### Maintenance
```bash
make clean        # Clean build artifacts
make kill-port    # Kill processes on port 8085
make restart      # Kill and restart server
```

## Security Features (Always Enabled)

✅ **Enterprise Security Headers**
- Content Security Policy with nonce-based execution
- XSS, CSRF, Clickjacking protection
- HSTS enforcement

✅ **Rate Limiting & DDoS Protection**
- 100 requests/minute per IP
- Progressive blocking (5-minute timeouts)
- Automatic cleanup

✅ **Input Validation & Sanitization**
- Strict regex-based validation
- XSS and SQL injection prevention
- Attack pattern detection

✅ **Security Monitoring**
- Comprehensive audit logging
- Suspicious activity detection
- Real-time threat monitoring

## Mobile Features (Always Enabled)

📱 **Responsive Design**
- Hamburger menu on mobile/tablet
- Touch-friendly interactions
- Optimized performance

🎨 **Consistent Experience**
- Same cyberpunk aesthetics across devices
- Smooth animations on mobile
- HTMX SPA navigation maintained

## Deployment Examples

### Quick Launch (No Calendar)
```bash
#!/bin/bash
export CALENDAR_ENABLED=false
export ENVIRONMENT=production
export PORT=80
make build
sudo ./server
```

### Full Feature Deployment
```bash
#!/bin/bash
export CALENDAR_ENABLED=true
export ENVIRONMENT=production
export PORT=80
make build
sudo ./server
```

### Docker Deployment
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

ENV CALENDAR_ENABLED=true
ENV ENVIRONMENT=production
ENV PORT=8085

EXPOSE 8085
CMD ["./server"]
```

## Health Checks

The server provides basic health information via logs:
- Configuration status on startup
- Security event monitoring
- Performance metrics
- Error tracking

## Troubleshooting

### Calendar Not Working
- Check `CALENDAR_ENABLED` environment variable
- Verify calendar routes are registered (check logs)
- Ensure templates include calendar navigation

### Mobile Menu Not Appearing
- Check browser width (menu appears ≤1024px)
- Verify JavaScript is loading
- Check for console errors

### Security Headers Missing
- Security middleware should auto-apply
- Check middleware order in main.go
- Verify no proxy stripping headers

---

**Ready for immediate deployment with enterprise-grade security and mobile optimization!**