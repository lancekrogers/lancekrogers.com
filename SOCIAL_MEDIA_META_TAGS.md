# Social Media Meta Tags Implementation

This document describes the comprehensive social media meta tags implementation for your blog posts, ensuring they display beautifully when shared on Twitter, Facebook, LinkedIn, Discord, Slack, and other platforms.

## ✅ **What's Been Implemented**

### **1. Database Schema Updates**
Added new fields to the `Post` struct and database schema:
- `FeaturedImage` - URL to post's featured image
- `OGDescription` - Custom Open Graph description
- `TwitterHandle` - Post-specific Twitter handle

### **2. Template Meta Tags**
Complete meta tags in `templates/pages/blog-post.html`:
- **Open Graph tags** (Facebook, LinkedIn, Discord, etc.)
- **Twitter Card tags** (summary_large_image format)
- **Article-specific tags** (author, publish date, tags)
- **RSS feed discovery** link

### **3. Configuration Support**
Extended `content/site.yml` with social media configuration:
```yaml
site:
  base_url: "https://blockheadconsulting.com"
  social:
    twitter_handle: "@lancerogers"
    default_image: "/static/images/blockhead-og-default.png"
    author_name: "Lance Rogers"
    author_twitter: "@lancerogers"
```

### **4. Helper Functions**
Added template helper functions:
- `formatURL` - Converts relative URLs to absolute
- `hasPrefix` - String prefix checking
- `printf` - String formatting

## 🚀 **How to Use**

### **1. Add Social Media Fields to Blog Posts**
You can now add these optional fields to your markdown frontmatter:

```yaml
---
title: "Your Amazing Blog Post"
date: 2025-06-17
summary: "A short summary of your post"
tags: ["web development", "social media"]
featured_image: "/static/images/my-post-image.png"
og_description: "Custom description for social media (can be different from summary)"
twitter_handle: "@customhandle"
---
```

### **2. Image Requirements**
For optimal social media display:
- **Recommended size**: 1200 × 630 pixels (1.91:1 ratio)
- **File formats**: PNG or JPG
- **File size**: Under 5MB
- **Location**: Place in `/static/images/` directory

### **3. Testing Your Implementation**
Use these tools to validate your meta tags:
- [Facebook Sharing Debugger](https://developers.facebook.com/tools/debug/)
- [Twitter Card Validator](https://cards-dev.twitter.com/validator)  
- [LinkedIn Post Inspector](https://www.linkedin.com/post-inspector/)

### **4. Example Meta Tags Generated**
For a post with featured image, the following tags are automatically generated:

```html
<!-- Basic meta tags -->
<title>Your Amazing Blog Post - Blockhead Consulting</title>
<meta name="description" content="Custom description for social media">

<!-- Open Graph tags -->
<meta property="og:type" content="article">
<meta property="og:url" content="https://blockheadconsulting.com/blog/your-amazing-blog-post">
<meta property="og:title" content="Your Amazing Blog Post">
<meta property="og:description" content="Custom description for social media">
<meta property="og:image" content="https://blockheadconsulting.com/static/images/my-post-image.png">
<meta property="og:site_name" content="Blockhead Consulting">

<!-- Twitter Card tags -->
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:url" content="https://blockheadconsulting.com/blog/your-amazing-blog-post">
<meta name="twitter:title" content="Your Amazing Blog Post">
<meta name="twitter:description" content="Custom description for social media">
<meta name="twitter:image" content="https://blockheadconsulting.com/static/images/my-post-image.png">
<meta name="twitter:creator" content="@customhandle">
<meta name="twitter:site" content="@lancerogers">

<!-- Article-specific tags -->
<meta property="article:author" content="Lance Rogers">
<meta property="article:published_time" content="2025-06-17T00:00:00Z">
<meta property="article:tag" content="web development">
<meta property="article:tag" content="social media">
```

## 🎯 **Smart Fallbacks**

The implementation includes intelligent fallbacks:

1. **Image fallback**: If no `featured_image` is specified, uses `site.social.default_image`
2. **Description fallback**: Uses `og_description` → `summary` → `site.description`
3. **Twitter handle fallback**: Uses post-specific `twitter_handle` → site `author_twitter`

## 📝 **Example Blog Post**

See the example post at `content/blog/social-media-meta-tags-example.md` for a complete demonstration of all features.

## 🔧 **Technical Details**

### **Database Schema**
The SQLite schema has been updated to include:
```sql
CREATE TABLE posts (
  -- ... existing fields ...
  featured_image TEXT,
  og_description TEXT,
  twitter_handle TEXT
);
```

### **Go Structs**
```go
type Post struct {
  // ... existing fields ...
  FeaturedImage string `json:"featured_image,omitempty" db:"featured_image"`
  OGDescription string `json:"og_description,omitempty" db:"og_description"`
  TwitterHandle string `json:"twitter_handle,omitempty" db:"twitter_handle"`
}

type Frontmatter struct {
  // ... existing fields ...
  FeaturedImage string `yaml:"featured_image"`
  OGDescription string `yaml:"og_description"`
  TwitterHandle string `yaml:"twitter_handle"`
}
```

## 🎨 **Best Practices**

### **Image Creation Tips**
1. **Design for readability** - Use high contrast text on images
2. **Include your brand** - Add your logo or brand colors
3. **Mobile-friendly** - Ensure text is readable on small screens
4. **Consistent style** - Maintain visual consistency across posts

### **Description Writing**
1. **Keep it engaging** - Write for humans, not just search engines
2. **Stay within limits** - 110-140 characters for Facebook, 200 for Twitter/LinkedIn
3. **Include keywords** - But naturally, don't stuff
4. **Call to action** - Give people a reason to click

### **Testing Workflow**
1. **Write your post** with social media fields
2. **Deploy your changes**
3. **Test with social media debugging tools**
4. **Share and verify** the appearance

## 🚨 **Important Notes**

1. **Use absolute URLs** - All image and link URLs must be absolute (include https://)
2. **Test regularly** - Social platforms cache meta tags, so test after changes
3. **Monitor performance** - Track click-through rates from social media
4. **Update default image** - Create and set a compelling default image for posts without featured images

## 🎉 **Result**

When you share your blog posts on social media, they will now display with:
- ✅ Compelling title and description
- ✅ Eye-catching featured image
- ✅ Proper attribution and branding
- ✅ Professional appearance across all platforms
- ✅ Higher click-through rates and engagement

Your blog posts will look amazing when shared, leading to better engagement and more traffic! 🚀