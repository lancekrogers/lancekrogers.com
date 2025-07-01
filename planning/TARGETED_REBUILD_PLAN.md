# Targeted Rebuild Plan: New Homepage + Style Evolution

## What We're Keeping (Don't Touch)
✅ **Blog System**: Entire blog with markdown, mermaid, syntax highlighting - working perfectly
✅ **About Page**: Layout, content, GitHub stats, resume - just needs style updates later
✅ **Backend Systems**: All Go backend, messaging, content management
✅ **Template Engine**: Works great, just need new homepage template

## What We're Rebuilding
🔄 **Homepage Only**: Complete redesign from scratch for personal branding
🔄 **Color Scheme**: New electric blue/purple or purple/magenta cyberpunk palette
🔄 **Navigation**: Update to reflect personal site (remove consulting CTAs)
🔄 **Hero System**: Simplify to single personal hero (no multiple modes)

## Color Scheme Decision

### Electric Blue/Purple (Option 1)
```css
--primary-electric: #00ccff;      /* Bright electric blue */
--secondary-purple: #cc00ff;      /* Vibrant purple */
--accent-cyan: #66ffff;           /* Light cyan highlights */
--accent-pink: #ff0099;           /* Hot pink accents */
--neon-white: #ffffff;            /* Pure white for text */
--cyber-dark: #0a0a0a;            /* Keep dark backgrounds */
```
**Feel**: Clean, high-tech, Tron-like

### Purple/Magenta (Option 2)  
```css
--primary-purple: #9900ff;        /* Deep purple */
--secondary-magenta: #ff00aa;     /* Bright magenta */
--accent-violet: #6600cc;         /* Dark violet */
--accent-cyan: #00ffaa;           /* Matrix green accent */
--neon-white: #ffffff;            /* Pure white highlights */
--cyber-dark: #0a0a0a;            /* Keep dark backgrounds */
```
**Feel**: Matrix-inspired, mysterious, sophisticated

## New Homepage Structure

### Simplified Personal Homepage Sections
```
1. Hero Section
   - "Lance Rogers" (personal name)
   - "Software Engineer & Systems Architect" 
   - Brief tagline about building scalable systems
   - Personal stats (years experience, projects, impact)
   - Simplified call-to-action

2. Featured Projects (3-4 cards)
   - Guild AI Framework
   - Claude Code Go SDK  
   - Key professional highlights
   - Clean cyber-styled project cards

3. Latest Writing (blog integration)
   - 2-3 recent blog posts
   - Link to full blog (keep existing blog completely)

4. Connect Section
   - Personal contact approach
   - Links to GitHub, social media
   - Simple contact form
```

## Implementation Strategy

### Phase 1: Design System (Week 1)
1. **Choose between the two color schemes** (blue/purple vs purple/magenta)
2. **Create new CSS variables** for the chosen palette
3. **Design simple personal logo** (probably `[ LR ]` or typography)
4. **Create homepage wireframe**

### Phase 2: Homepage Rebuild (Week 1-2)
1. **Create new homepage template** (`templates/pages/home-personal.html`)
2. **Build new homepage styles** (`static/scss/pages/_home-personal.scss`)
3. **Update hero animation** with new colors and personal name
4. **Create project showcase components**

### Phase 3: Integration (Week 2)
1. **Update navigation** to remove consulting CTAs
2. **Connect new homepage** to existing backend
3. **Update site configuration** to use new homepage
4. **Preserve all existing functionality**

### Phase 4: Style Evolution (Future)
1. **Gradually update about page** styling to match new colors
2. **Update blog styling** to coordinate (but keep functionality)
3. **Refine animations** and polish details

## File Structure for New Homepage

### New Files to Create
```
templates/pages/home-personal.html          # New homepage template
static/scss/pages/_home-personal.scss       # New homepage styles  
static/scss/themes/_cyberpunk-personal.scss # New color scheme
static/js/modules/personal-animations.js    # New animation logic
```

### Files to Modify Minimally
```
content/site.yml                    # Update personal branding info
templates/layouts/partials/nav.html # Remove consulting CTAs
internal/handlers/home.go           # Point to new template
```

### Files to Leave Alone
```
All blog-related files              # Don't touch the blog!
templates/pages/about-full.html     # Keep existing about page
All backend Go code                 # Working perfectly
All messaging/contact systems       # Keep as-is
```

## Specific Questions to Finalize

### 1. Color Scheme Choice
Which feels more "you"?
- **Electric Blue/Purple**: Clean, high-tech, approachable
- **Purple/Magenta**: Mysterious, sophisticated, unique

### 2. Personal Branding
For the hero section:
- Name: "Lance Rogers" 
- Title: "Software Engineer & Systems Architect"
- Tagline: Something like "Building scalable blockchain and AI infrastructure"?

### 3. Hero Animation
- Keep the glitch effect but with new colors?
- Simplify to just typewriter effect?
- New animation style entirely?

### 4. Project Showcase
Which 3-4 projects to feature on homepage?
- Guild AI Framework (definitely)
- Claude Code Go SDK (definitely)  
- Professional highlight (Mythical Games? Bank of America?)
- Personal project (YouTube Summarizer? ShinySwap?)

## Benefits of This Approach

✅ **Low Risk**: We're not touching any of the complex, working systems
✅ **Fast Implementation**: Only rebuilding homepage, not entire site
✅ **Iterative**: Can evolve about page and blog styling later
✅ **Focused**: Solve the immediate need (personal vs corporate branding)
✅ **Preserves Investment**: All your hard work on blog/about remains intact

This gives you a completely fresh personal homepage while keeping all the sophisticated functionality you've built. What do you think about this targeted approach? And which color scheme is calling to you?