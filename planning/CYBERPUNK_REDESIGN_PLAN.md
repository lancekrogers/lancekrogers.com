# Cyberpunk Redesign Plan for lancekrogers.com

## Design Philosophy
Keep the excellent cyberpunk aesthetic but create a completely distinct visual identity from blockhead.consulting. Fresh start on frontend while preserving the robust backend systems.

## Color Scheme Options (Cyberpunk Alternatives)

### Option 1: Electric Blue/Purple (Recommended)
```css
--primary-blue: #00ccff;        /* Bright electric blue */
--secondary-purple: #cc00ff;    /* Vibrant purple */
--accent-cyan: #66ffff;         /* Light cyan */
--accent-pink: #ff0099;         /* Hot pink */
--warning-orange: #ff6600;      /* Cyber orange */
```
**Vibe**: Tron-like, electric, high-tech

### Option 2: Red/Orange Cyber
```css
--primary-red: #ff3333;         /* Bright red */
--secondary-orange: #ff9900;    /* Cyber orange */
--accent-yellow: #ffff00;       /* Electric yellow */
--accent-pink: #ff0066;         /* Neon pink */
--cool-blue: #0099ff;           /* Contrast blue */
```
**Vibe**: Cyberpunk 2077, aggressive, energetic

### Option 3: Purple/Magenta Matrix
```css
--primary-purple: #9900ff;      /* Deep purple */
--secondary-magenta: #ff00aa;   /* Bright magenta */
--accent-violet: #6600cc;       /* Dark violet */
--accent-lime: #99ff00;         /* Matrix green accent */
--highlight-white: #ffffff;     /* Pure white highlights */
```
**Vibe**: Matrix-inspired, mysterious, sophisticated

## Frontend Rebuild Strategy

### What to Keep (Backend Systems)
✅ **Blog System**: Content management, markdown processing, search
✅ **Messaging System**: Encrypted contact forms, Git storage
✅ **Template Engine**: Go template system works great
✅ **Configuration System**: YAML-based content management
✅ **Build System**: Make file, Docker setup
✅ **Animation Framework**: The glitch/typewriter system is excellent

### What to Rebuild (Frontend)
🔄 **HTML Structure**: New layout and component organization
🔄 **CSS/SCSS**: Complete visual redesign with new color scheme
🔄 **JavaScript**: Keep animations but rebuild interactions
🔄 **Templates**: New template designs for personal branding
🔄 **Assets**: New imagery, icons, logos

## New Visual Identity Elements

### Personal Logo Options
**Option 1: Monogram Style**
```
[ LR ]  or  { LR }  or  <LR/>
```
Code-inspired brackets with initials

**Option 2: Typography Treatment**
```
LANCE ROGERS
────────────
SOFTWARE ENGINEER
```
Clean typography with cyber underlines

**Option 3: Glitch Text Effect**
```
L̲A̲N̲C̲E̲ ̲R̲O̲G̲E̲R̲S̲
```
Cyberpunk-style distorted text

### Animation System Evolution
- **Keep**: Typewriter effect, glitch animations, scan lines
- **Update**: Color scheme to match new palette
- **Add**: New personal touches (maybe code syntax highlighting effects)
- **Improve**: Mobile performance and accessibility

## Frontend Architecture

### New SCSS Structure
```
static/scss/
├── abstracts/
│   ├── _variables-cyberpunk.scss    # New color variables
│   ├── _mixins.scss
│   └── _functions.scss
├── base/
│   ├── _reset.scss
│   ├── _typography.scss             # Updated fonts
│   └── _animations.scss             # Refined glitch effects
├── components/
│   ├── _buttons.scss
│   ├── _cards.scss                  # Project cards
│   ├── _navigation.scss
│   └── _terminal.scss               # New terminal-style elements
├── layout/
│   ├── _hero.scss                   # Personal hero section
│   ├── _grid.scss
│   └── _sections.scss
├── pages/
│   ├── _home.scss                   # Personal portfolio home
│   ├── _about.scss
│   ├── _projects.scss               # New projects showcase
│   └── _blog.scss                   # Keep existing blog styles
└── themes/
    └── _cyberpunk-personal.scss     # New theme
```

### New Component Ideas
- **Terminal Window**: Fake terminal showing code/stats
- **Matrix Rain**: Subtle background effect with initials
- **Holographic Cards**: For project showcases
- **Code Syntax Highlighting**: For technical sections
- **Neon Borders**: Glowing outlines on sections
- **Particle Systems**: Subtle floating code characters

## Content Structure Changes

### New Homepage Sections
1. **Hero**: Personal name with new glitch effect
2. **About**: Quick personal intro
3. **Featured Projects**: 3-4 key projects with cyber cards
4. **Latest Posts**: Recent blog entries
5. **Contact**: "Connect with me" messaging

### Project Showcase Design
- **Holographic Cards**: Each project in a glowing card
- **Tech Stack Tags**: Neon-styled technology badges
- **Live Demo Links**: Cyber-styled buttons
- **GitHub Integration**: Real-time stars/activity display

## Implementation Approach

### Phase 1: Design System (Week 1)
1. **Choose color scheme** (need your preference!)
2. **Create new design variables**
3. **Design component library**
4. **Create personal logo/branding**

### Phase 2: Core Layout (Week 2)
1. **Rebuild hero section** with personal branding
2. **Create project showcase components**
3. **Update navigation** for personal site
4. **Implement new color scheme**

### Phase 3: Content & Polish (Week 3)
1. **Update all content** for personal brand
2. **Add project details** and portfolio items
3. **Refine animations** with new colors
4. **Mobile optimization**

### Phase 4: Integration (Week 4)
1. **Connect to existing backend** systems
2. **Test blog and messaging** functionality
3. **Performance optimization**
4. **Launch preparation**

## Questions for You

### Color Preference
Which cyberpunk color scheme appeals to you most?
- Electric Blue/Purple (Tron-style)
- Red/Orange (Cyberpunk 2077-style)  
- Purple/Magenta (Matrix-style)
- Something else you have in mind?

### Logo Style
What personal branding approach feels right?
- Code brackets with initials [ LR ]
- Clean typography treatment
- Glitch text effects
- Something completely different?

### Project Focus
Which projects should be featured prominently?
- Guild AI Framework
- Claude Code Go SDK
- YouTube Summarizer
- Professional work (Mythical Games, Bank of America)
- Open source contributions
- Personal experiments

### Animation Intensity
How much animation do you want?
- Keep current intensity but new colors
- Tone it down for broader appeal
- Go even more intense/experimental
- Situational (subtle normally, intense on special sections)

This approach lets us keep all the excellent technical work you've done while creating a completely fresh, personal cyberpunk identity. What are your thoughts on the color schemes and overall direction?