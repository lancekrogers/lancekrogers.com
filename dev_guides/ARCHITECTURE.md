# Blockhead Consulting Website Architecture

## Overview

This document describes the architecture of the Blockhead Consulting website - a modern Go application with HTMX for dynamic UI, Git-based storage for data persistence, and event-driven communication between services.

## System Architecture

### High-Level Architecture

```
┌─────────────────┐
│   Web Browser   │
│      (HTMX)     │
└────────┬────────┘
         │ HTTPS
         ▼
┌─────────────────┐
│   Caddy Server  │ (TLS termination)
│  (Reverse Proxy)│
└────────┬────────┘
         │ HTTP
         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go Application Server                     │
│                                                              │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   Router    │  │  Middleware  │  │    Templates     │  │
│  │  (gorilla)  │  │    Stack     │  │  (HTML + HTMX)  │  │
│  └──────┬──────┘  └──────────────┘  └──────────────────┘  │
│         │                                                    │
│         ▼                                                    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Service Layer                           │   │
│  │                                                      │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌────────┐│   │
│  │  │  Blog   │  │Contact  │  │ Email   │  │Security││   │
│  │  │Service  │  │Service  │  │Service  │  │Service ││   │
│  │  └─────────┘  └─────────┘  └─────────┘  └────────┘│   │
│  └─────────────────────────────────────────────────────┘   │
│         │              │              │                      │
│         ▼              ▼              ▼                      │
│  ┌─────────────┐ ┌─────────────┐ ┌────────────┐           │
│  │   Event     │ │Git Storage  │ │   Gmail    │           │
│  │    Bus      │ │   Layer     │ │   SMTP     │           │
│  └─────────────┘ └─────────────┘ └────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Web Layer

```
HTTP Request Flow:

Browser Request
     │
     ▼
Caddy Server (HTTPS)
     │
     ▼
Security Middleware
├─→ Rate Limiting
├─→ CSP Headers
├─→ Request Validation
└─→ Audit Logging
     │
     ▼
Router (gorilla/mux)
     │
     ├─→ Static Routes    (/static/*)
     ├─→ Page Routes      (/, /blog, /contact)
     ├─→ HTMX Routes      (/content/*)
     └─→ API Routes       (/api/*)
     │
     ▼
Handler Functions
     │
     ▼
Service Layer
```

### 2. Service Registry & Dependency Injection

```
Application Startup:

main()
  │
  ├─→ Load Configuration (.env)
  │
  ├─→ Initialize Services
  │     ├─→ BlogService
  │     ├─→ ContactService
  │     ├─→ EmailService
  │     └─→ GitStorage
  │
  ├─→ Register Services
  │     │
  │     ▼
  │   ServiceRegistry
  │     ├─→ Register("blog", blogService)
  │     ├─→ Register("contact", contactService)
  │     ├─→ Register("email", emailService)
  │     └─→ Register("storage", gitStorage)
  │
  └─→ Start HTTP Server
```

### 3. Event-Driven Architecture

```
Event Flow:

User Action (e.g., Submit Contact Form)
     │
     ▼
Handler validates & creates domain object
     │
     ▼
Publish Event to Event Bus
     │
     ▼
Event Bus (Async)
     │
     ├─→ Handler 1: Git Storage
     │     └─→ Encrypt & Store
     │
     ├─→ Handler 2: Email Service
     │     └─→ Send Notification
     │
     └─→ Handler 3: Audit Logger
           └─→ Log Event
```

### 4. Git Storage System

```
Message Storage Flow:

New Message
     │
     ▼
Validate & Create Message Object
     │
     ▼
Convert to JSON
     │
     ▼
Encrypt (AES-256-GCM)
     │
     ▼
Generate Timestamped Filename
messages/2024/01/2024-01-15_14-30-45_abc123.json.enc
     │
     ▼
Write to Git Repository
     │
     ├─→ git add
     ├─→ git commit -m "New message from John Doe"
     └─→ git push (async)

Repository Structure:
messages-repo/
├── messages/
│   └── 2024/
│       └── 01/
│           ├── 2024-01-15_14-30-45_abc123.json.enc
│           └── 2024-01-15_16-22-10_def456.json.enc
└── README.md
```

### 5. Email Service Architecture

```
Email Sending Flow:

Event: MessageReceived
     │
     ▼
Email Handler Triggered
     │
     ▼
Load Email Template
     │
     ▼
Render HTML with Data
     │
     ▼
Build SMTP Message
     │
     ▼
Connect to Gmail SMTP
├─→ Host: smtp.gmail.com
├─→ Port: 587 (STARTTLS)
└─→ Auth: App Password
     │
     ▼
Send Email
     │
     ├─→ Success: Log & Done
     │
     └─→ Failure: Retry Queue
           │
           ▼
     Exponential Backoff
     (5min → 15min → 45min)
```

### 6. Blog System

```
Blog Content Flow:

Build Time:
Markdown Files → Parse Frontmatter → Generate HTML → Embed in Binary

Runtime:
Request /blog
     │
     ▼
BlogHandler
     │
     ▼
BlogService.GetAllPosts()
     │
     ▼
Return Cached Posts
     │
     ▼
Render Template
     │
     ▼
Return HTML

Blog Structure:
content/
└── blog/
    ├── 2024-01-15-post-title.md
    └── 2024-01-20-another-post.md

Frontmatter:
---
title: "Post Title"
date: 2024-01-15
tags: ["golang", "web"]
summary: "Post summary"
---
```

## Data Flow Patterns

### 1. HTMX Single Page Application Pattern

```
Initial Page Load:
Browser → GET / → Full HTML Page with HTMX

Navigation:
User Clicks Link → HTMX Intercepts → GET /content/blog → HTML Fragment → Swap Content

Benefits:
- No full page reloads
- Smaller payloads
- Better UX
- SEO friendly
```

### 2. Contact Form Submission

```
1. User fills form
     │
2. HTMX POST /contact
     │
3. Server validates
     │
4. Create Message object
     │
5. Publish MessageReceived event
     │
6. Return success HTML fragment
     │
7. HTMX swaps response into page

Parallel (via Event Bus):
├─→ Git: Encrypt & store message
├─→ Email: Send notification
└─→ Log: Audit trail
```

### 3. Security Middleware Stack

```
Request → [Rate Limiter] → [Size Limiter] → [Security Headers] → [CSRF] → [Logger] → Handler

Each middleware can:
- Pass to next (normal flow)
- Return error (block request)
- Modify request/response
- Log security events
```

## Security Architecture

### 1. Defense in Depth

```
Layer 1: Infrastructure
├─→ Caddy (TLS 1.3)
├─→ Firewall rules
└─→ Oracle Cloud security

Layer 2: Application
├─→ Rate limiting
├─→ Input validation
├─→ CSRF protection
└─→ Security headers

Layer 3: Data
├─→ Encryption at rest
├─→ Git access control
└─→ Environment variables
```

### 2. Authentication Flow (Future)

```
Current: Basic Auth for admin
     │
     ▼
Future: JWT-based auth
     │
     ├─→ Login endpoint
     ├─→ Token generation
     ├─→ Token validation
     └─→ Refresh tokens
```

## Deployment Architecture

### Oracle Cloud Infrastructure

```
Oracle Cloud Always Free Tier
     │
     ├─→ Compute: ARM VM (1 OCPU, 6GB RAM)
     ├─→ Storage: 50GB boot + 150GB block
     ├─→ Network: 10TB bandwidth
     └─→ Region: US-West (Phoenix)

VM Setup:
┌─────────────────────────────┐
│      Ubuntu 22.04 ARM64     │
│                             │
│  ┌────────┐  ┌───────────┐ │
│  │ Caddy  │  │    Go     │ │
│  │ :443   │→ │   :8085   │ │
│  └────────┘  └───────────┘ │
│                             │
│  ┌────────────────────────┐ │
│  │     Git Repos          │ │
│  │  ├─→ messages/         │ │
│  │  ├─→ blog/            │ │
│  │  └─→ calendar/        │ │
│  └────────────────────────┘ │
└─────────────────────────────┘
```

### Deployment Process

```
Development:
1. Code changes
2. Run tests
3. Build binary
4. Test locally

Production:
1. Build for ARM64
   GOOS=linux GOARCH=arm64 go build

2. Copy to server
   scp blockhead-server ubuntu@server:/opt/blockhead/

3. Restart service
   sudo systemctl restart blockhead

4. Verify health
   curl https://blockhead.consulting/health
```

## Monitoring & Observability

### 1. Logging Strategy

```
Log Levels:
├─→ ERROR:   System errors, failed operations
├─→ WARN:    Degraded performance, retries
├─→ INFO:    Normal operations, requests
└─→ DEBUG:   Detailed debugging info

Structured Logging:
{
  "time": "2024-01-15T14:30:45Z",
  "level": "INFO",
  "msg": "Contact form submitted",
  "request_id": "abc123",
  "user_ip": "192.168.1.1",
  "email": "user@example.com"
}
```

### 2. Health Checks

```
GET /health

Response:
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "24h35m",
  "checks": {
    "git_repos": "ok",
    "smtp": "ok",
    "disk_space": "ok (45% used)"
  }
}
```

### 3. Metrics (Future)

```
Planned Metrics:
├─→ Request rate
├─→ Response times
├─→ Error rates
├─→ Business metrics
│   ├─→ Messages received
│   ├─→ Blog views
│   └─→ Bookings made
└─→ System metrics
    ├─→ CPU usage
    ├─→ Memory usage
    └─→ Disk usage
```

## Technology Stack

### Core Technologies

| Component | Technology | Why |
|-----------|------------|-----|
| Language | Go 1.24 | Performance, simplicity |
| Web Framework | gorilla/mux | Mature, flexible routing |
| UI Framework | HTMX | SPA without JavaScript complexity |
| Templates | html/template | Built-in, secure |
| Configuration | godotenv | Simple .env files |
| Storage | Git | Version control built-in |
| Encryption | AES-256-GCM | Industry standard |
| Email | Gmail SMTP | Reliable, free |
| Deployment | Oracle Cloud | Generous free tier |
| TLS | Caddy | Automatic HTTPS |

### Development Tools

| Tool | Purpose |
|------|---------|
| Docker Compose | Local testing environment |
| MailHog | SMTP testing |
| Make | Build automation |
| Git | Version control |

## Future Enhancements

### Phase 1: Current Implementation
- [x] Blog system with markdown
- [x] HTMX navigation
- [x] Security middleware
- [ ] Git storage for messages
- [ ] Email notifications
- [ ] Service registry
- [ ] Event bus

### Phase 2: Post-Launch
- [ ] Search functionality
- [ ] RSS feed
- [ ] Admin dashboard
- [ ] Analytics
- [ ] Newsletter signup

### Phase 3: Advanced Features
- [ ] Booking system (with security)
- [ ] Calendar integration
- [ ] Client portal
- [ ] Payment integration

## Development Guidelines

### 1. Code Organization

```
- Keep handlers thin
- Business logic in services
- Interfaces over implementations
- Dependency injection
- Error handling at boundaries
```

### 2. Testing Strategy

```
- Unit tests: 80% coverage
- Integration tests: Critical paths
- E2E tests: User journeys
- Performance tests: Load scenarios
- Security tests: OWASP Top 10
```

### 3. Security First

```
- Validate all inputs
- Sanitize all outputs
- Use prepared statements
- Implement rate limiting
- Log security events
- Encrypt sensitive data
```

## Conclusion

This architecture provides a solid foundation for a professional consulting website with room to grow. The modular design, event-driven architecture, and Git-based storage create a system that is both simple to understand and powerful enough for future enhancements.

Key benefits:
- **Simple**: No external databases or services
- **Secure**: Encryption and security at every layer
- **Scalable**: Event-driven design allows easy extension
- **Maintainable**: Clean architecture with clear boundaries
- **Cost-effective**: Runs on Oracle Cloud's free tier