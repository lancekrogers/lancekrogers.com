# Content Management Guide

This guide explains how to manage and update content for your Blockhead Consulting website using markdown files.

## Content Structure

Your site content is organized in the `/content/` directory:

```
content/
├── bio-brief.md      # Homepage bio snippet
├── about.md          # Full about page
├── site-config.md    # Documentation template (not active)
└── blog/             # Blog posts
    ├── post1.md
    ├── post2.md
    └── post3.md
```

## Bio Content Management

### Homepage Bio (`content/bio-brief.md`)

This appears in the bio section between Technical Expertise and Contact Form.

**Format:**

```markdown
---
title: "About Lance Rogers"
---

**Strategic systems architect with 9+ years engineering complex financial systems** at Fortune 500 institutions including Bank of America, Mythical Games, and Shutterfly. Lance Rogers specializes in premium crypto infrastructure and enterprise AI integrations, uniquely combining deep technical expertise with strategic business insight—architecting systems that drive measurable business outcomes, not just technical solutions.
```

**Guidelines:**

- Keep to 2-3 sentences maximum
- Use **bold text** for key phrases
- Focus on authority and value proposition
- No headers needed (just paragraph text)

### About Page (`content/about.md`)

This is the full biography page with structured content.

**Format:**

```markdown
---
title: "About Lance Rogers"
subtitle: "Strategic Systems Architect & Technical Consultant"
---

## Strategic Technology Leadership

**Lance Rogers architects enterprise-grade systems** that transform complex technologies into high-value business outcomes. With 9+ years engineering experience at Fortune 500 institutions including Bank of America, Mythical Games, and Shutterfly, he specializes in crypto infrastructure and enterprise AI integrations.

## Proven Track Record

Your track record content here...

### Core Expertise

- **Crypto Infrastructure**: Blockchain integration, custodial wallets
- **Enterprise AI Systems**: Claude SDK development, AI agent orchestration
- **Strategic Systems Thinking**: Workflow optimization, technical audits

### Innovation Leadership

- **Proprietary Frameworks**: Creator of Guild—an enterprise AI agent orchestration system
- **Open Source Contributions**: Active contributor to blockchain and AI development communities
- **Strategic Architecture**: Designer of modular blockchain systems

## Business-Focused Approach

Your business approach content...

### Client Value Delivered

- Reduced development cycles through modular blockchain architectures
- Enabled new revenue streams via innovative NFT and crypto payment systems
- Accelerated AI adoption through production-ready integration frameworks

## Strategic Consulting Services

**Crypto Infrastructure Consulting** - Full-stack blockchain solutions including smart contracts, secure wallet systems, and scalable platform architecture.

**Enterprise AI & Claude SDK Consulting** - Leveraging cutting-edge AI capabilities for business transformation through custom integrations and agent orchestration.

**Strategic Systems Thinking Engagements** - Executive advisory focused on identifying system-level improvements and aligning technology investments with strategic goals.
```

**Supported Markdown Features:**

- `# ## ###` Headers (create page structure)
- `**bold text**` Strong emphasis
- `- item` Bullet lists
- `[link text](url)` Links
- Regular paragraphs
- Auto-generated heading IDs for navigation

## Blog Content Management

### Adding New Blog Posts

Create a new file in `/content/blog/` with this format:

**Filename**: `your-post-slug.md`

**Content Format:**

```markdown
---
title: "Your Blog Post Title"
date: 2025-01-28
summary: "Brief description of the post for listings"
tags: ["blockchain", "ai", "consulting"]
readingTime: 5
---

# Your Blog Post Title

Your blog content here with full markdown support...

## Section Headers

Regular paragraphs with **bold text** and _italic text_.

### Subsections

- Bullet points
- More bullet points

`code blocks`

[Links to other content](https://example.com)
```

**Guidelines:**

- Use descriptive filenames (becomes the URL slug)
- Include all frontmatter fields
- Keep summaries under 160 characters
- Add relevant tags for categorization
- Estimate reading time in minutes

### Managing Existing Posts

To update existing blog posts:

1. Edit the `.md` file in `/content/blog/`
2. Save the file
3. Changes appear immediately (no restart needed)

## Customizing Page Titles and Metadata

### About Page Title/Subtitle

Edit the frontmatter in `/content/about.md`:

```markdown
---
title: "About Lance Rogers" # Page title and H1
subtitle: "Strategic Systems Architect & Technical Consultant" # Appears under title
---
```

### Homepage Bio Title

Edit the frontmatter in `/content/bio-brief.md`:

```markdown
---
title: "About Lance Rogers" # Used in navigation and meta
---
```

## Content Guidelines

### Writing Style

- **Professional tone** for enterprise credibility
- **Strategic language** (architect, transform, optimize, scale)
- **Quantified results** when possible (9+ years, Fortune 500)
- **Value-focused** language over technical jargon

### SEO Best Practices

- Use descriptive page titles
- Include key terms naturally in content
- Add meta descriptions via frontmatter
- Use header hierarchy (H1 → H2 → H3)

### Markdown Tips

**Headers:**

```markdown
# Page Title (H1)

## Main Section (H2)

### Subsection (H3)
```

**Emphasis:**

```markdown
**Bold for key points**
_Italic for emphasis_
```

**Lists:**

```markdown
- Bullet point
- Another point
  - Nested point

1. Numbered list
2. Second item
```

**Links:**

```markdown
[Link text](https://example.com)
[Internal link](/about)
```

## File Management

### Safe Editing

- Always backup files before major changes
- Test changes on development environment first
- Use version control (git) to track changes

### File Naming

- Use lowercase with hyphens: `my-blog-post.md`
- Avoid spaces and special characters
- Be descriptive but concise

### Content Updates Process

1. **Edit** the markdown file
2. **Save** the changes
3. **Test** on local development server
4. **Deploy** to production

## Troubleshooting

### Common Issues

**HTML not rendering properly:**

- Check markdown syntax
- Ensure proper frontmatter format
- Verify file encoding (UTF-8)

**Content not updating:**

- Check file permissions
- Verify file location
- Restart application if needed

**Broken links:**

- Use relative paths: `/about` not `https://site.com/about`
- Check file exists at target location

### Getting Help

If you encounter issues:

1. Check this documentation first
2. Verify markdown syntax online
3. Test with minimal content first
4. Contact technical support if needed

## Future Enhancements

Planned content management improvements:

- Web-based content editor
- Live preview functionality
- Automated deployment on content changes
- Media file management system
- Content scheduling capabilities

---

This content management system provides professional-grade website management while maintaining simplicity and flexibility for rapid updates.
