---
# ⚠️ DOCUMENTATION TEMPLATE ONLY
# This file is NOT currently used by the application
# It serves as a template for future site cloning and Guild Ventures

# Site Configuration Template
# Copy this file and customize for new sites

site_name: "Blockhead Consulting"
site_description: "Enterprise Blockchain & AI Infrastructure"
hero_style: "professional" # or "cyberpunk"

# About page configuration
about_title: "About Lance Rogers"
about_subtitle: "Strategic Systems Architect & Technical Consultant"

# Contact information
contact_email: "lance@blockhead.consulting"
contact_phone: "+1 (555) 123-4567"

# Social links
linkedin: "https://linkedin.com/in/your-profile"
twitter: "https://twitter.com/your-handle"
github: "https://github.com/your-username"

# Business details
company_name: "Blockhead Consulting LLC"
tagline: "Bridging traditional finance with blockchain technology"
subtitle: "Building production-grade AI systems that scale"

# Services configuration
service_1_title: "Crypto Infrastructure"
service_1_rate: "$300-500/hour"
service_2_title: "AI/LLM Consulting"
service_2_rate: "$200-400/hour"

# Package pricing
assessment_price: "$2,500"
pilot_price: "$15,000"
enterprise_price: "$50,000+"

# Features
calendar_enabled: false
blog_enabled: true
analytics_enabled: false

# Branding
logo_path: "/static/logos/svg/blockhead-three-blocks-green.svg"
profile_image: "/static/images/lance_profile.jpg"
---

# Site Configuration Guide

**🚨 IMPORTANT: This file is currently DOCUMENTATION ONLY**

This file serves as a template for configuring new sites built on this infrastructure. The settings in this file do NOT currently affect the running website - they are hardcoded in the Go application and templates.

## Current Status

**What this file IS:**

- Documentation template for future site cloning
- Reference for Guild Ventures setup
- Specification for potential future configuration system

**What this file is NOT:**

- An active configuration that changes site behavior
- Currently read by the application
- A way to modify the running site settings

## To Actually Change Site Settings

Currently, to modify site settings you need to:

1. **Content**: Edit `/content/bio-brief.md` and `/content/about.md`
2. **Visual Elements**: Edit CSS in `/static/styles.css`
3. **Site Title/Meta**: Edit templates in `/templates/`
4. **Boot Sequence**: Edit `/static/boot-sequence.js`
5. **Services/Pricing**: Edit `/templates/pages/home.html`

## Quick Start for New Sites

1. **Copy this entire website repository**
2. **Update content files**:
   - `content/bio-brief.md` - Homepage bio snippet
   - `content/about.md` - Full about page
   - `content/blog/` - Blog posts
3. **Update site configuration** (this file)
4. **Replace images**:
   - Logo files in `/static/logos/`
   - Profile image in `/static/images/`
5. **Update domain and deployment settings**

## Customization Points

### Content Management

- All content is managed via markdown files in `/content/`
- Bio content supports full markdown: headers, lists, links, bold text
- Blog posts automatically generate from `/content/blog/*.md`
- New pages can be added to `/content/pages/` (future feature)

### Styling

- Colors defined in CSS variables in `/static/styles.css`
- Two hero styles: "professional" and "cyberpunk"
- Responsive design works on mobile and desktop

### Features

- Contact form with email integration
- HTMX for smooth navigation
- Git-based content storage (optional)
- Calendar booking system (optional)
- Security middleware with rate limiting

## Guild Ventures Example

For Guild Ventures site:

1. Update `site_name` to "Guild Ventures"
2. Change `about_title` to "About Guild Ventures"
3. Update hero messages in `/static/boot-sequence.js`
4. Replace logo and profile images
5. Modify service offerings in templates
6. Update contact information and social links

## Technical Architecture

This system provides:

- **Modular Go backend** with clean architecture
- **Markdown-based content management**
- **Template-driven page generation**
- **Security middleware**
- **Event system for integrations**
- **Git storage for messages**
- **Email notification system**
- **Docker development environment**

Perfect for rapid deployment of professional consulting websites.
