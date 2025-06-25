package blog

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"time"

	"blockhead.consulting/internal/config"
)

// SocialMetaTags represents all the meta tags needed for social media sharing
type SocialMetaTags struct {
	// Basic meta tags
	Title       string
	Description string

	// Open Graph tags
	OGType        string
	OGURL         string
	OGTitle       string
	OGDescription string
	OGImage       string
	OGSiteName    string

	// Twitter Card tags
	TwitterCard        string
	TwitterURL         string
	TwitterTitle       string
	TwitterDescription string
	TwitterImage       string
	TwitterCreator     string
	TwitterSite        string

	// Article-specific tags
	ArticleAuthor        string
	ArticlePublishedTime string
	ArticleModifiedTime  string
	ArticleTags          []string
}

// GenerateSocialMetaTags creates social media meta tags for a blog post
func GenerateSocialMetaTags(post *Post, siteConfig *config.SiteConfig) *SocialMetaTags {
	baseURL := strings.TrimSuffix(siteConfig.Site.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://blockheadconsulting.com" // Default fallback
	}

	postURL := fmt.Sprintf("%s/blog/%s", baseURL, post.Slug)

	// Use post's OG description if available, otherwise summary, otherwise truncated content
	description := post.OGDescription
	if description == "" {
		description = post.Summary
	}
	if description == "" && post.Content != "" {
		// Extract first 160 characters from content (remove HTML)
		contentStr := string(post.Content)
		contentStr = stripHTML(contentStr)
		if len(contentStr) > 160 {
			description = contentStr[:157] + "..."
		} else {
			description = contentStr
		}
	}
	if description == "" {
		description = siteConfig.Site.Description
	}

	// Determine featured image
	featuredImage := post.FeaturedImage
	if featuredImage == "" {
		featuredImage = siteConfig.Site.Social.DefaultImage
	}
	if featuredImage != "" && !strings.HasPrefix(featuredImage, "http") {
		// Convert relative path to absolute URL
		featuredImage = fmt.Sprintf("%s%s", baseURL, featuredImage)
	}

	// Determine Twitter handle
	twitterCreator := post.TwitterHandle
	if twitterCreator == "" {
		twitterCreator = siteConfig.Site.Social.AuthorTwitter
	}
	if twitterCreator != "" && !strings.HasPrefix(twitterCreator, "@") {
		twitterCreator = "@" + twitterCreator
	}

	twitterSite := siteConfig.Site.Social.TwitterHandle
	if twitterSite != "" && !strings.HasPrefix(twitterSite, "@") {
		twitterSite = "@" + twitterSite
	}

	// Format dates for article tags
	publishedTime := post.Date.Format(time.RFC3339)
	modifiedTime := post.UpdatedAt.Format(time.RFC3339)
	if post.UpdatedAt.IsZero() {
		modifiedTime = publishedTime
	}

	return &SocialMetaTags{
		// Basic meta tags
		Title:       post.Title,
		Description: description,

		// Open Graph tags
		OGType:        "article",
		OGURL:         postURL,
		OGTitle:       post.Title,
		OGDescription: description,
		OGImage:       featuredImage,
		OGSiteName:    siteConfig.Site.Name,

		// Twitter Card tags
		TwitterCard:        "summary_large_image",
		TwitterURL:         postURL,
		TwitterTitle:       post.Title,
		TwitterDescription: description,
		TwitterImage:       featuredImage,
		TwitterCreator:     twitterCreator,
		TwitterSite:        twitterSite,

		// Article-specific tags
		ArticleAuthor:        siteConfig.Site.Social.AuthorName,
		ArticlePublishedTime: publishedTime,
		ArticleModifiedTime:  modifiedTime,
		ArticleTags:          post.Tags,
	}
}

// RenderMetaTags generates HTML meta tags for social media
func (s *SocialMetaTags) RenderMetaTags() template.HTML {
	var builder strings.Builder

	// Basic meta tags
	if s.Title != "" {
		builder.WriteString(fmt.Sprintf(`<title>%s</title>%s`, escapeHTML(s.Title), "\n"))
	}
	if s.Description != "" {
		builder.WriteString(fmt.Sprintf(`<meta name="description" content="%s">%s`, escapeHTML(s.Description), "\n"))
	}

	// Open Graph tags
	if s.OGType != "" {
		builder.WriteString(fmt.Sprintf(`<meta property="og:type" content="%s">%s`, escapeHTML(s.OGType), "\n"))
	}
	if s.OGURL != "" {
		builder.WriteString(fmt.Sprintf(`<meta property="og:url" content="%s">%s`, escapeHTML(s.OGURL), "\n"))
	}
	if s.OGTitle != "" {
		builder.WriteString(fmt.Sprintf(`<meta property="og:title" content="%s">%s`, escapeHTML(s.OGTitle), "\n"))
	}
	if s.OGDescription != "" {
		builder.WriteString(fmt.Sprintf(`<meta property="og:description" content="%s">%s`, escapeHTML(s.OGDescription), "\n"))
	}
	if s.OGImage != "" {
		builder.WriteString(fmt.Sprintf(`<meta property="og:image" content="%s">%s`, escapeHTML(s.OGImage), "\n"))
	}
	if s.OGSiteName != "" {
		builder.WriteString(fmt.Sprintf(`<meta property="og:site_name" content="%s">%s`, escapeHTML(s.OGSiteName), "\n"))
	}

	// Twitter Card tags
	if s.TwitterCard != "" {
		builder.WriteString(fmt.Sprintf(`<meta name="twitter:card" content="%s">%s`, escapeHTML(s.TwitterCard), "\n"))
	}
	if s.TwitterURL != "" {
		builder.WriteString(fmt.Sprintf(`<meta name="twitter:url" content="%s">%s`, escapeHTML(s.TwitterURL), "\n"))
	}
	if s.TwitterTitle != "" {
		builder.WriteString(fmt.Sprintf(`<meta name="twitter:title" content="%s">%s`, escapeHTML(s.TwitterTitle), "\n"))
	}
	if s.TwitterDescription != "" {
		builder.WriteString(fmt.Sprintf(`<meta name="twitter:description" content="%s">%s`, escapeHTML(s.TwitterDescription), "\n"))
	}
	if s.TwitterImage != "" {
		builder.WriteString(fmt.Sprintf(`<meta name="twitter:image" content="%s">%s`, escapeHTML(s.TwitterImage), "\n"))
	}
	if s.TwitterCreator != "" {
		builder.WriteString(fmt.Sprintf(`<meta name="twitter:creator" content="%s">%s`, escapeHTML(s.TwitterCreator), "\n"))
	}
	if s.TwitterSite != "" {
		builder.WriteString(fmt.Sprintf(`<meta name="twitter:site" content="%s">%s`, escapeHTML(s.TwitterSite), "\n"))
	}

	// Article-specific tags
	if s.ArticleAuthor != "" {
		builder.WriteString(fmt.Sprintf(`<meta property="article:author" content="%s">%s`, escapeHTML(s.ArticleAuthor), "\n"))
	}
	if s.ArticlePublishedTime != "" {
		builder.WriteString(fmt.Sprintf(`<meta property="article:published_time" content="%s">%s`, escapeHTML(s.ArticlePublishedTime), "\n"))
	}
	if s.ArticleModifiedTime != "" {
		builder.WriteString(fmt.Sprintf(`<meta property="article:modified_time" content="%s">%s`, escapeHTML(s.ArticleModifiedTime), "\n"))
	}

	// Article tags
	for _, tag := range s.ArticleTags {
		if tag != "" {
			builder.WriteString(fmt.Sprintf(`<meta property="article:tag" content="%s">%s`, escapeHTML(tag), "\n"))
		}
	}

	return template.HTML(builder.String())
}

// stripHTML removes HTML tags from content (basic implementation)
func stripHTML(content string) string {
	// Simple HTML stripping - remove everything between < and >
	result := strings.Builder{}
	inTag := false

	for _, char := range content {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}

	// Clean up whitespace
	cleaned := strings.TrimSpace(result.String())
	// Replace multiple spaces/newlines with single space
	cleaned = strings.ReplaceAll(cleaned, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\t", " ")
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}

	return cleaned
}

// escapeHTML escapes HTML characters for use in attributes
func escapeHTML(s string) string {
	return template.HTMLEscapeString(s)
}

// ValidateImageURL checks if the image URL is properly formatted
func ValidateImageURL(imageURL string) bool {
	if imageURL == "" {
		return false
	}

	// Check if it's a valid URL
	if _, err := url.Parse(imageURL); err != nil {
		return false
	}

	// Check if it's an image file
	lowerURL := strings.ToLower(imageURL)
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	for _, ext := range validExtensions {
		if strings.HasSuffix(lowerURL, ext) {
			return true
		}
	}

	return false
}

// TruncateDescription ensures description fits within social media limits
func TruncateDescription(description string, limit int) string {
	if len(description) <= limit {
		return description
	}

	// Try to break at word boundary
	truncated := description[:limit-3]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > limit/2 { // Only use word boundary if it's not too short
		truncated = description[:lastSpace]
	}

	return truncated + "..."
}
