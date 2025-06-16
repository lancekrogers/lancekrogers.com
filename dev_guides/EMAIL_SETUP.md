# Email Configuration Guide

## Overview
The Blockhead Consulting website sends email notifications when contact form submissions are received. This guide explains how to configure email sending for both development and production environments.

## Development Setup (MailHog)

For local development, we recommend using MailHog to capture emails without actually sending them.

### 1. Start MailHog
```bash
# Using Docker Compose (recommended)
docker compose up mailhog -d

# Or run MailHog directly
docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog
```

### 2. Configure Environment
Your `.env` file should have these settings for MailHog:
```bash
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM_ADDRESS=noreply@blockhead.consulting
SMTP_FROM_NAME=Blockhead Consulting
SMTP_TLS_ENABLED=false
ADMIN_EMAIL=admin@blockhead.consulting
```

### 3. View Captured Emails
Open http://localhost:8025 in your browser to see all captured emails.

## Production Setup

### Recommended: Third-Party Email Services

For production, we recommend using dedicated email services rather than Gmail/Google Workspace due to authentication complexities with business accounts.

### 1. **Brevo (Recommended - 300 emails/day forever)**
1. Sign up at https://www.brevo.com
2. Go to SMTP & API > SMTP
3. Create SMTP key
4. Configure:
```bash
SMTP_HOST=smtp-relay.brevo.com
SMTP_PORT=587
SMTP_USERNAME=your-login-email@example.com
SMTP_PASSWORD=your-smtp-key
SMTP_FROM_ADDRESS=lance@blockhead.consulting
SMTP_FROM_NAME=Blockhead Consulting
SMTP_TLS_ENABLED=true
ADMIN_EMAIL=lance@blockhead.consulting
```

### 2. **Resend (100 emails/day forever)**
1. Sign up at https://resend.com
2. Create API key
3. Configure:
```bash
SMTP_HOST=smtp.resend.com
SMTP_PORT=587
SMTP_USERNAME=resend
SMTP_PASSWORD=re_xxxxxxxxxxxx
SMTP_FROM_ADDRESS=lance@blockhead.consulting
SMTP_FROM_NAME=Blockhead Consulting
SMTP_TLS_ENABLED=true
ADMIN_EMAIL=lance@blockhead.consulting
```

### 3. **MailerSend (12,000 emails/month forever)**
1. Sign up at https://www.mailersend.com
2. Go to Settings > API Tokens
3. Create SMTP credentials
4. Configure:
```bash
SMTP_HOST=smtp.mailersend.net
SMTP_PORT=587
SMTP_USERNAME=your-smtp-username
SMTP_PASSWORD=your-smtp-password
SMTP_FROM_ADDRESS=lance@blockhead.consulting
SMTP_FROM_NAME=Blockhead Consulting
SMTP_TLS_ENABLED=true
ADMIN_EMAIL=lance@blockhead.consulting
```

### 4. **Gmail (Personal Accounts Only)**
**Note**: App Passwords are not available for Google Workspace/business accounts.

For personal Gmail accounts only:
1. Enable 2-Factor Authentication at https://myaccount.google.com/security
2. Generate App Password at https://myaccount.google.com/apppasswords
3. Configure:
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password-here
SMTP_FROM_ADDRESS=your-email@gmail.com
SMTP_FROM_NAME=Blockhead Consulting
SMTP_TLS_ENABLED=true
ADMIN_EMAIL=your-email@gmail.com
```

### 5. **Your Web Host SMTP**
Many hosting providers include SMTP. Check with:
- DigitalOcean App Platform
- Vercel (via add-ons)
- Railway
- Your VPS provider
- cPanel hosting

Example configuration:
```bash
SMTP_HOST=mail.yourdomain.com
SMTP_PORT=587
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=your-email-password
SMTP_FROM_ADDRESS=noreply@yourdomain.com
SMTP_FROM_NAME=Blockhead Consulting
SMTP_TLS_ENABLED=true
ADMIN_EMAIL=admin@yourdomain.com
```

## Email Notification Flow

1. User submits contact form
2. Form is validated
3. Message is encrypted and saved to Git storage
4. Email notification is sent to `ADMIN_EMAIL` with:
   - Contact person's name
   - Email address
   - Company (if provided)
   - Message content
   - Message ID for tracking
   - Timestamp
   - IP address

## Testing Email Configuration

### 1. Verify Configuration
```bash
# Check if email service can connect
curl -X GET http://localhost:8087/health
```

### 2. Test Contact Form
```bash
# Submit a test contact form
curl -X POST http://localhost:8087/contact \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "name=Test User&email=test@example.com&company=Test Corp&message=This is a test message to verify email notifications are working correctly."
```

### 3. Check Email Delivery
- **Development**: Check MailHog at http://localhost:8025
- **Production**: Check your configured admin email inbox

## Troubleshooting

### "Failed to send notification email" in logs
1. Check SMTP credentials are correct
2. Verify network connectivity to SMTP server
3. Check if ports are blocked by firewall
4. For Gmail, ensure app password is being used (not regular password)

### No emails received but no errors
1. Check `ADMIN_EMAIL` is configured
2. Check spam/junk folder
3. Verify SMTP server logs if available
4. Try with `SMTP_TLS_ENABLED=false` for testing

### Gmail specific issues
1. **Google Workspace/Business accounts**: App Passwords are not available - use a third-party service instead
2. **Personal Gmail**: Ensure "Less secure app access" is NOT being used (deprecated)
3. Use App Passwords instead of regular passwords
4. Check if account has 2FA enabled
5. Verify sending limits haven't been exceeded

### Brevo specific issues
1. Check SMTP key is correct (not API key)
2. Verify your account email as the username
3. Make sure you're using `smtp-relay.brevo.com` not the old `smtp-relay.sendinblue.com`

### Resend specific issues
1. API key must start with `re_`
2. Username is always `resend` (not your email)
3. From address must be from a verified domain

## Security Considerations

1. **Never commit credentials**: Keep SMTP passwords in `.env` file only
2. **Use TLS in production**: Always set `SMTP_TLS_ENABLED=true` for production
3. **Validate sender**: Use your own domain for `SMTP_FROM_ADDRESS` when possible
4. **Rate limiting**: The contact form has built-in rate limiting to prevent abuse
5. **SPF/DKIM**: Configure these DNS records for better deliverability

## Environment Variables Reference

| Variable | Description | Example |
|----------|-------------|---------|
| `SMTP_HOST` | SMTP server hostname | `smtp.gmail.com` |
| `SMTP_PORT` | SMTP server port | `587` |
| `SMTP_USERNAME` | SMTP authentication username | `user@gmail.com` |
| `SMTP_PASSWORD` | SMTP authentication password | `app-specific-password` |
| `SMTP_FROM_ADDRESS` | Email address to send from | `noreply@blockhead.consulting` |
| `SMTP_FROM_NAME` | Display name for sender | `Blockhead Consulting` |
| `SMTP_TLS_ENABLED` | Enable TLS encryption | `true` |
| `ADMIN_EMAIL` | Email to receive notifications | `admin@blockhead.consulting` |