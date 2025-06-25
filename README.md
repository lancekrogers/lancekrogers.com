# Blockhead Consulting Website

A professional consulting website for blockchain and AI infrastructure services, built with Go, HTMX, and a cyberpunk aesthetic.

## Features

### **Core Technology**

- **Go + HTMX SPA**: Lightning-fast single-page application with server-side rendering
- **Embedded Templates**: All assets bundled in single binary for easy deployment
- **Enterprise Security**: Rate limiting, input validation, security headers, and audit logging
- **Responsive Design**: Mobile hamburger navigation, desktop-optimized layouts

### **Content Management System**

A powerful file-based CMS that requires no database:

- **YAML Configuration**: Site-wide settings, work portfolio, and blog configuration via YAML files
- **Markdown Content**: Write blog posts and pages in markdown with YAML frontmatter
- **Automatic Processing**: Converts markdown to HTML with syntax highlighting
- **Portfolio Management**: Showcase projects with descriptions, links, and tags
- **Bio/About Pages**: Maintain brief and full bio versions for different contexts
- **Blog Features**: Search, tag filtering, and automatic reading time calculation

### **Professional Features**

- **Layout Inheritance**: Organized template system for easy maintenance and consistency
- **Calendar Booking**: Self-hosted consultation booking (security-hardened)
- **GitHub Integration**: Real-time stats and activity display with intelligent caching
- **Configuration System**: Environment-based feature toggles and settings
- **Cyberpunk Aesthetic**: Unique design with glitch effects and terminal styling
- **Encrypted Message Storage**: Contact form submissions stored encrypted with Git versioning

### **Production Ready**

- **Single Binary Deployment**: All templates and assets embedded
- **Graceful Shutdown**: Proper signal handling and cleanup
- **Comprehensive Testing**: Full test coverage with automated workflows
- **Security Best Practices**: Built-in protection against common vulnerabilities

## Quick Start

```bash
# Clone and setup
git clone https://github.com/[yourusername]/blockhead-consulting.git
cd blockhead-consulting

# Install dependencies
go mod download

# Create environment config
cp .env.example .env
# IMPORTANT: Edit .env with your own values before running!

# Build and run
make build
./bin/blockhead-server

# Visit http://localhost:8085
```

## 📚 Documentation

Documentation is organized in the `dev_guides/` directory:

- **Architecture Overview** - System design and components
- **Email Setup** - Configure notifications with Brevo/Gmail
- **Message Encryption** - Encrypted contact form storage
- **Security Guide** - Security implementation details
- **Deployment Guide** - Production deployment instructions

## Project Structure

```
blockhead-consulting/
├── main.go                    # Application entry point
├── main_test.go              # Comprehensive test suite
├── go.mod                    # Go dependencies
├── Makefile                  # Build automation
├── .env.example              # Configuration template
├── bin/                      # Build artifacts (git-ignored)
├── templates/                # Organized template system
│   ├── layouts/
│   │   ├── base.html        # Main page layout
│   │   ├── fragment.html    # HTMX fragment layout
│   │   └── partials/
│   │       ├── nav.html     # Navigation component
│   │       └── footer.html  # Footer component
│   ├── pages/               # Full page templates
│   │   ├── home.html
│   │   ├── blog.html        # Blog with search/filtering
│   │   ├── blog-post.html
│   │   └── calendar.html
│   └── fragments/           # HTMX content fragments
│       ├── home-content.html
│       ├── blog-content.html
│       └── calendar-content.html
├── static/                   # Static assets
│   ├── styles.css           # Cyberpunk styling
│   ├── main.js              # Navigation and animations
│   └── logos/               # Brand assets
├── content/
│   └── blog/                # Markdown blog posts
│       ├── ai_trading_agent.md
│       ├── claude-code-go-sdk-announcement.md
│       ├── smart_contract_audit.md
│       └── why-irc174-not-ai-killed-tech-jobs.md
└── data/                    # Application data (git-ignored)
    └── bookings.json
```

## Content Management

### **Writing Blog Posts**

Create markdown files in `content/blog/` with YAML frontmatter:

````markdown
---
title: "Infrastructure Patterns for CBDCs"
date: 2024-05-26
summary: "Exploring technical architecture requirements for CBDC implementations."
tags: ["blockchain", "cbdc", "infrastructure", "fintech"]
readingTime: 12
---

# Infrastructure Patterns for CBDCs

Your blog content here with full markdown support...

## Code Examples

```go
func main() {
    // Syntax highlighting included
    fmt.Println("Hello, World!")
}
```
````

## Features

- Automatic HTML conversion
- Syntax highlighting
- Reading time calculation
- Tag-based organization
- Search functionality

### **Blog Features**

- **Automatic Processing**: Drop markdown files in `content/blog/`, restart server
- **Search & Filtering**: Live search by title/content, filter by tags
- **Reading Time**: Automatically calculated based on content length
- **Responsive Design**: Optimized for all device sizes
- **Tag System**: Organize posts by blockchain, AI, golang, etc.

### **Content Management System**

The file-based CMS allows you to manage all site content without a database:

#### **Configuration Files**

1. **Site Configuration** (`content/site.yml`):
   ```yaml
   title: "Blockhead Consulting"
   tagline: "Bridging traditional finance with blockchain technology"
   hero_style: "professional"  # or "playful" for different hero animations
   features:
     calendar_enabled: true
     blog_enabled: true
   ```

2. **Work Portfolio** (`content/work.yml`):
   ```yaml
   projects:
     - title: "DeFi Trading Platform"
       description: "Built automated trading system processing $10M+ daily"
       tags: ["blockchain", "golang", "aws"]
       link: "https://example.com"
       featured: true
   ```

3. **Blog Configuration** (`content/blog.yml`):
   ```yaml
   posts_per_page: 10
   enable_comments: false
   default_author: "Lance Rogers"
   featured_tags: ["blockchain", "ai", "golang"]
   ```

4. **Bio/About Content**:
   - `content/bio-brief.md` - Short bio for homepage
   - `content/about.md` - Full about page with detailed background

#### **Managing Content**

- **Simple Updates**: Edit YAML/Markdown files and restart the server
- **Version Control Friendly**: All content in plain text files, perfect for Git
- **Flexible Structure**: Organize content in subdirectories as needed
- **Preview Support**: Test content changes locally before deploying

### **Template System**

The organized template structure supports:

- **Layout Inheritance**: Define navigation/footer once, use everywhere
- **Component Reuse**: Shared partials for consistent UI elements
- **Page Context**: Active navigation states and conditional content
- **HTMX Fragments**: Separate templates for SPA content loading

### **Message System**

The contact form uses an innovative Git-based storage system for encrypted message archival:

#### **How It Works**

1. When a visitor submits the contact form, their message is encrypted using AES-256-GCM
2. The encrypted message is saved to `data/messages/` with a timestamped filename
3. The system automatically creates a Git commit with message metadata
4. Optionally pushes to a remote repository for backup and team access

#### **Setup Instructions**

1. **Generate Encryption Key**:

   ```bash
   go run cmd/generate-key/main.go
   # Copy the generated key to your .env file
   ```

2. **Initialize Message Repository**:

   ```bash
   cd data/messages
   git init
   git remote add origin your-private-repo-url  # Optional: for remote backup
   cd ../..
   ```

3. **Configure Environment Variables**:

   ```bash
   # Required
   GIT_REPO_PATH=./data/messages
   GIT_ENCRYPTION_KEY=your-generated-32-byte-key

   # Optional - for automatic remote backup
   GIT_REMOTE_URL=git@github.com:yourorg/contact-messages.git
   GIT_PUSH_ON_WRITE=true  # Enable automatic push after each message
   GIT_BRANCH=main
   GIT_COMMIT_AUTHOR=Blockhead Bot
   GIT_COMMIT_EMAIL=bot@blockhead.consulting
   ```

4. **Managing Messages**:

   ```bash
   # View message status
   go run cmd/message-status/main.go

   # Decrypt and read messages
   go run cmd/decrypt-messages/main.go

   # Messages are organized by date: data/messages/YYYY/MM/
   ```

#### **Security Benefits**

- **Encryption at Rest**: Messages are never stored in plain text
- **Audit Trail**: Git history provides complete audit log of all submissions
- **Team Access**: Share encrypted messages via private Git repository
- **Backup**: Automatic versioning and optional remote backup
- **Compliance**: Maintains data integrity and access logs for compliance

## Configuration

### **Environment Variables**

Create `.env` file or set environment variables:

```bash
# Core Settings
PORT=8085
ENVIRONMENT=production
SITE_NAME="Blockhead Consulting"

# Feature Toggles
CALENDAR_ENABLED=true

# Security (REQUIRED for admin interface)
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your-secure-password

# Email Configuration (optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Message System Configuration
GIT_REPO_PATH=./data/messages           # Where encrypted messages are stored
GIT_ENCRYPTION_KEY=your-32-byte-key     # Generate with: go run cmd/generate-key/main.go
GIT_REMOTE_URL=                         # Optional: Push messages to remote repo
GIT_PUSH_ON_WRITE=false                 # Auto-push on new messages
GIT_BRANCH=main                         # Branch for message storage
GIT_COMMIT_AUTHOR=Blockhead Bot         # Git commit author
GIT_COMMIT_EMAIL=bot@blockhead.consulting

# GitHub Integration (optional)
GITHUB_USERNAME=your-github-username   # Your GitHub username for stats display
GITHUB_TOKEN=your-github-token         # GitHub Personal Access Token
GITHUB_ENABLED=true                    # Enable/disable GitHub stats (auto-detected if username/token provided)
```

### **Configuration Features**

- **Calendar Toggle**: Completely disable booking system when `CALENDAR_ENABLED=false`
- **GitHub Integration**: Display real-time GitHub stats on the about page with automatic fallback to defaults
- **Environment Detection**: Different behavior for development/production
- **Security Settings**: Environment-based secrets management

### **GitHub Integration Setup**

To enable live GitHub stats on your about page:

1. **Create a GitHub Personal Access Token**:
   - Go to GitHub Settings → Developer settings → Personal access tokens → Fine-grained tokens
   - Click "Generate new token"
   - Set expiration and select your repositories
   - Required permissions: `metadata:read`, `contents:read` for public repositories
   - For more comprehensive stats, you may also want `pull_requests:read`

2. **Configure Environment Variables**:
   ```bash
   GITHUB_USERNAME=your-github-username
   GITHUB_TOKEN=github_pat_your_token_here
   ```

3. **Features**:
   - **Real-time Stats**: Public repos, stars, contributions, pull requests
   - **Recent Activity**: Latest pushes, PR merges, stars, etc.
   - **Intelligent Caching**: 1-hour cache to respect GitHub API limits
   - **Graceful Fallback**: Shows default values if GitHub is unavailable
   - **Security**: Only public data is accessed, token stored securely

If no GitHub credentials are provided, the about page will display default placeholder values instead.

## Deployment

### **Single Binary Deployment (Recommended)**

```bash
# Build for production
make build

# Deploy single file
scp bin/blockhead-server user@server:/opt/blockhead/
ssh user@server 'sudo systemctl restart blockhead'
```

### **Docker Deployment**

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bin/blockhead-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/blockhead-server .
EXPOSE 8085
CMD ["./blockhead-server"]
```

### **Systemd Service**

```ini
[Unit]
Description=Blockhead Consulting Website
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/blockhead
ExecStart=/opt/blockhead/blockhead-server
Restart=always
RestartSec=5
Environment=PORT=8085
Environment=ENVIRONMENT=production

[Install]
WantedBy=multi-user.target
```

## Security Features

### **Built-in Protection**

- **Rate Limiting**: Per-IP request limits with automatic blocking
- **Input Validation**: Comprehensive sanitization and validation
- **Security Headers**: CSP, HSTS, XSS protection, and more
- **Audit Logging**: Security event monitoring and logging

### **Production Security**

1. **Reverse Proxy**: Use Nginx/Caddy with Let's Encrypt
2. **Firewall**: Restrict access to necessary ports only
3. **Secrets Management**: Use environment variables for sensitive data
4. **Regular Updates**: Keep dependencies and system updated

### **Calendar Security Warning**

⚠️ **The booking system requires security implementation before production use.** The booking feature should remain disabled (`CALENDAR_ENABLED=false`) until proper authentication and security measures are implemented.

## Development

### **Local Development**

```bash
# Install dependencies
go mod download

# Run tests
make test

# Run with hot reload (using air or similar)
air

# Format code
go fmt ./...

# Build and test
make build
make test
```

### **Adding Features**

1. **New Pages**: Create templates in `templates/pages/` using layout inheritance
2. **New Components**: Add to `templates/layouts/partials/`
3. **Blog Posts**: Drop markdown files in `content/blog/`
4. **Static Assets**: Add to `static/` directory

### **Template Development**

```html
<!-- New page template -->
{{template "base" .}} {{define "content"}}
<section class="your-section">
  <div class="container">
    <h1>{{.Title}}</h1>
    <!-- Your content here -->
  </div>
</section>
{{end}}
```

## Testing

```bash
# Run all tests
make test

# Run specific test suite
go test -run TestSecurity -v

# Test with coverage
go test -cover ./...

# Integration tests
go test -tags=integration ./...
```

## Performance

- **Sub-10ms Response Times**: Optimized Go handlers
- **Embedded Assets**: No file I/O for templates/static files
- **Minimal JavaScript**: HTMX for SPA without heavy frameworks
- **Efficient Templating**: Compiled templates with layout inheritance
- **CDN Ready**: Static assets can be served from CDN

## Roadmap

### **Content Features**

- [ ] RSS/Atom feeds for blog
- [ ] Blog post categories and series
- [ ] Related posts suggestions
- [ ] Comment system integration

### **Business Features**

- [ ] Payment integration (Stripe/crypto)
- [ ] Client portal with project tracking
- [ ] Service package configurator
- [ ] Automated proposal generation

### **Technical Improvements**

- [ ] GraphQL API for headless usage
- [ ] WebSocket real-time features
- [ ] Advanced caching strategies
- [ ] Multi-language support

## License

Copyright 2025 Blockhead Consulting. All rights reserved.

---

**Built with ❤️ and focus by [Lance Rogers](https://github.com/lancekrogers)**

_Ready to modernize your infrastructure? [Book a consultation](/calendar) to discuss your blockchain and AI needs._
