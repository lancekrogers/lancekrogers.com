package blog

import (
	"html/template"
	"strings"
)

// GetTemplateFunctions returns template functions for blog-specific operations
func GetTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"hasPrefix":    strings.HasPrefix,
		"hasSuffix":    strings.HasSuffix,
		"contains":     strings.Contains,
		"toLower":      strings.ToLower,
		"toUpper":      strings.ToUpper,
		"trimSpace":    strings.TrimSpace,
		"joinStrings":  strings.Join,
		"splitString":  strings.Split,
		"formatURL":    formatURL,
		"truncateText": truncateText,
		"safeHTML":     func(s string) template.HTML { return template.HTML(s) },
		"safeHTMLAttr": func(s string) template.HTMLAttr { return template.HTMLAttr(s) },
		"safeJS":       func(s string) template.JS { return template.JS(s) },
		"safeURL":      func(s string) template.URL { return template.URL(s) },
	}
}

// formatURL ensures URLs are properly formatted with protocol
func formatURL(baseURL, path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	baseURL = strings.TrimSuffix(baseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return baseURL + path
}

// truncateText truncates text to specified length with ellipsis
func truncateText(text string, length int) string {
	if len(text) <= length {
		return text
	}

	// Try to break at word boundary
	truncated := text[:length-3]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > length/2 { // Only use word boundary if it's not too short
		truncated = text[:lastSpace]
	}

	return truncated + "..."
}
