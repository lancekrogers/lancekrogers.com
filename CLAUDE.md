# Lance Rogers Personal Website

## Technical Overview
Personal website built with Go backend and HTMX frontend, featuring a cyberpunk-inspired design for networking and personal branding.

## Backend Architecture

### **Go Web Server**
- **Framework**: Native Go HTTP server with custom routing
- **Templates**: Go template engine for server-side rendering
- **Configuration**: YAML-based configuration system (`content/site.yml`)
- **Content Management**: Markdown files for blog posts with frontmatter
- **Static Assets**: CSS/SCSS compilation and JavaScript modules

### **Key Go Packages**
- `internal/handlers/` - HTTP request handlers
- `internal/config/` - Configuration management
- `internal/blog/` - Blog post processing
- `internal/bio/` - About page content
- `internal/templates/` - Template management

### **Data Storage**
- **Blog**: SQLite database for posts and metadata
- **Configuration**: YAML files for site content
- **Static Files**: File system serving

## Frontend Framework

### **HTMX Integration**
- **Page Navigation**: HTMX handles SPA-style navigation without full page reloads
- **Content Loading**: Dynamic content loading with `hx-get` attributes
- **URL Management**: `hx-push-url` for proper browser history
- **Target Updates**: `hx-target="#main-content"` for content replacement

### **JavaScript Architecture**
- **Modular Design**: ES6 modules in `/static/js/modules/`
- **Animation System**: Custom animations for cyberpunk effects
- **Home Enhancements**: Terminal typing effects, floating particles, 3D transforms
- **No Heavy Frameworks**: Vanilla JavaScript for performance

### **CSS/SCSS Structure**
- **SCSS Compilation**: Automated build process with `make build-css`
- **Component-Based**: Separate SCSS files for different page sections
- **CSS Variables**: Cyberpunk color scheme with CSS custom properties
- **Responsive Design**: Mobile-first approach with breakpoints

## File Structure
```
/templates/
  /layouts/ - Base templates and partials
  /pages/ - Full page templates
  /fragments/ - Reusable template components

/static/
  /scss/ - Source SCSS files
  /css/ - Compiled CSS output
  /js/ - JavaScript modules and main files

/content/ - YAML configuration files
/internal/ - Go backend code
```

## Development Workflow

### **Build Commands**
- `make dev` - Start development server
- `make build-css` - Compile SCSS to CSS
- `make build` - Build Go binary
- `make test` - Run test suite

### **Configuration**
- All content managed through YAML files in `/content/`
- Personal homepage content in `site.yml`
- Blog configuration and posts in separate files
- No hardcoded content in templates

### **Security Features**
- Content Security Policy (CSP) headers
- CSRF protection
- Rate limiting
- Secure headers and HTTPS enforcement

## Current Features

### **Visual Design**
- Cyberpunk aesthetic with terminal/CRT styling
- Color scheme: Black backgrounds with blue/purple accents
- Typography: Monospace fonts for technical elements
- Glassmorphism effects and subtle animations

### **Interactive Elements**
- Terminal typing animation for hero text
- Floating code particles background
- 3D project card hover effects
- Animated stat counters
- Smooth scrolling navigation

### **Content Management**
- YAML-driven content for easy updates
- Blog system with markdown support
- Dynamic GitHub stats integration
- Project showcase with tech stack tags

## Code Organization
- Binaries in `/bin/` - not committed to git
- Use Makefile for all build/test operations
- Modular Go code with clear separation of concerns
- Component-based frontend architecture