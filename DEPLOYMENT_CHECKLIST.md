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
