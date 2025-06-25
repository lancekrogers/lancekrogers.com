# Blockhead Consulting Website

## Current State Overview
This is a professional consulting website for blockchain/AI infrastructure services. The design features a clean, cyberpunk-inspired aesthetic with subtle CRT monitor effects.

## Current Features

### **Visual Design**
- **Terminal/CRT Aesthetic**: Clean monospace fonts with subtle scan-line effects throughout the hero section
- **Color Scheme**: 
  - Primary: Black backgrounds (`#0a0a0a`, `#1a1a1a`, `#252525`)
  - Text: White primary text (`#ffffff`) with gray secondaries
  - Accents: Green crypto (`#00ff88`), blue AI (`#00d4ff`), red warning (`#ff6b6b`)
- **Typography**: Mix of system fonts for readability and JetBrains Mono for code/technical elements

### **Animation System**
- **Full-Page Black Mirror Effect (12s duration)**: Coordinated chaos across entire page
  - Page-wide glitch overlay with moving RGB scan lines
  - All sections experience movement, color shifts, and corruption effects
  - Creates immersive tech dystopia atmosphere for maximum attention
- **Main Animation (7s duration)**: Glitch effect on "BLOCKHEAD CONSULTING" heading
  - Uses CSS pseudo-elements (::before, ::after) with different colored text shadows
  - Clips portions of text at different timings for authentic glitch effect
  - Runs simultaneously with full-page effect
- **Typewriter Effect**: JavaScript-driven character-by-character reveal starting after 500ms
- **Periodic Glitch**: Subtle glitch effects trigger every 20-30 seconds after main animation
- **CRT Scan Lines**: Permanent subtle horizontal lines across hero section for monitor aesthetic

### **Technical Implementation**

**CSS Structure:**
- `.glitch` - Main heading with data-text attribute for pseudo-element content
- `.glitch::before` - Red text shadow glitch layer
- `.glitch::after` - Green/blue text shadow glitch layer
- Complex keyframe animations (`glitch-anim`, `glitch-anim2`, `glitch-rotate`) with precise timing
- `hide-pseudo` animation to hide glitch layers after main animation
- `.subtle-glitch` class for periodic effects

**JavaScript Functions:**
- Full-page glitch overlay creation and management
- Global glitch class coordination across all page elements
- Typewriter effect with character-by-character text reveal
- Animation loop using `requestAnimationFrame` for smooth performance
- Periodic glitch scheduling with randomized 20-30 second intervals
- Animation state management (typing, glitch active/inactive, timing)

### **Current Content**
- **Hero Heading**: "BLOCKHEAD CONSULTING"
- **Subtitle**: "Bridging traditional finance with blockchain technology. Building production-grade AI systems that scale."
- **Services**: Crypto Infrastructure and AI/LLM Consulting with detailed service lists
- **Pricing**: Hourly rates ($200-500/hour) and service packages ($2.5k-$50k+)
- **Technical Expertise**: Languages, Blockchain, AI/ML, Infrastructure sections

### **Performance Considerations**
- Uses `requestAnimationFrame` for smooth animations
- Minimal DOM manipulation with efficient text replacement
- Lightweight CSS animations with hardware acceleration hints
- No heavy libraries - vanilla JavaScript implementation

## Files
- `/static/styles.css` - Complete CSS with cyberpunk theme and glitch animations
- `/static/main.js` - Animation system and smooth scrolling
- `/templates/index.html` - Main page structure with hero, services, and contact sections

## Animation Timing
- **Full-Page Glitch**: 12-second Black Mirror-style chaos affecting entire page
- **Typewriter**: Starts 500ms after page load, 50ms per character
- **Main Glitch**: 7-second complex animation sequence (runs during full-page effect)
- **Periodic Glitch**: Every 20-30 seconds, 1-second duration (starts after full-page effect ends)
- **Mobile**: No timing differences (same as desktop)

## Configuration System
The site includes a configuration system for deployment flexibility:

### Calendar Feature Toggle
- Calendar functionality can be disabled via config for rapid deployment
- When disabled, calendar links are completely removed from navigation (not just hidden)
- Prevents hackers from discovering calendar endpoints in HTML/JavaScript
- Environment variable: `CALENDAR_ENABLED=false` disables calendar features

### Mobile Navigation
- Responsive hamburger menu on mobile/tablet devices
- Desktop navigation remains unchanged
- Automatically adapts based on screen size
- Contains: Services, Blog, and Book Consultation (when calendar enabled)

## ⚠️ CRITICAL: Booking System Security Requirements

### **DO NOT IMPLEMENT WITHOUT EXPLICIT APPROVAL**
**The booking system has CRITICAL security vulnerabilities and must NOT be worked on until:**
1. Website content development is completed (current priority)
2. Explicit approval is given by the project owner
3. Security implementation plan is reviewed and approved

### **Current Security Status: HIGH RISK**
The existing booking implementation contains severe vulnerabilities:
- **Public API endpoints** with no authentication (`/api/slots`)
- **Hardcoded admin credentials** (`admin:changeme`) 
- **Unencrypted sensitive data** storage (client emails, project details)
- **No access monitoring** or audit logging
- **Information disclosure** of business patterns

### **Required Security Implementation**
**⚠️ AGENTS: Do NOT implement these fixes without explicit project owner approval**

See detailed security implementation plan: `ai_docs/BOOKING_SECURITY_IMPLEMENTATION.md`

**Summary of required fixes:**
1. **Authentication & Authorization**: JWT tokens, session management, remove hardcoded credentials
2. **Data Protection**: Encryption at rest, input validation, data sanitization  
3. **Monitoring & Logging**: Access logging, security event monitoring, audit trails
4. **Rate Limiting**: Endpoint-specific limits, enhanced protection
5. **Configuration Security**: Environment-based secrets, validation

### **Priority Level: LOW**
- Content development takes precedence
- Security implementation deferred until content completion
- Booking feature should remain disabled in production until security fixes are implemented

## Future Enhancement Ideas
**Subtle Ongoing Interference** (to implement after main site development):
- Random character substitution in text elements
- Occasional chromatic aberration on images/icons
- Micro-glitches on hover states
- Background static during form interactions
- Memory leak simulation in console logs
- Cursor corruption effects
- Loading state anomalies

## Code Organization
- Binaries should live in @bin no binaries should be committed to git tracking
- Use the make file for testing, stop creating binaries in the top level directory