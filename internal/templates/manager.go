package templates

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"strings"
	"time"
)

// NewManager creates a new template manager
func NewManager(templateFS fs.FS, logger *log.Logger) (*Manager, error) {
	funcMap := createFuncMap()
	
	tmpl := template.New("main").Funcs(funcMap)
	
	// Parse all HTML templates with multiple patterns
	// Order matters - pages must be parsed last
	patterns := []string{
		"templates/layouts/partials/*.html",
		"templates/fragments/*.html", 
		"templates/layouts/*.html",
		"templates/pages/*.html",
	}
	
	for _, pattern := range patterns {
		if _, err := tmpl.ParseFS(templateFS, pattern); err != nil {
			// Some patterns might not match any files, that's OK
			logger.Printf("Warning: pattern %s matched no files: %v", pattern, err)
		}
	}
	
	logger.Printf("Loaded %d templates", len(tmpl.Templates()))
	
	return &Manager{
		templates: tmpl,
		funcMap:   funcMap,
	}, nil
}

// GetTemplates returns the underlying template.Template
func (m *Manager) GetTemplates() *template.Template {
	return m.templates
}

// createFuncMap returns the template function map
func createFuncMap() template.FuncMap {
	return template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"replaceAll": strings.ReplaceAll,
		"slug": slug,
		"hasPrefix": strings.HasPrefix,
		"formatURL": formatURL,
		"printf": fmt.Sprintf,
		"add": func(a, b int) int {
			return a + b
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"getActiveTag": getActiveTag,
		"getCurrentYear": getCurrentYear,
		"timeAgo": timeAgo,
		"title": toTitle,
		"formatActivityType": formatActivityType,
	}
}

// Template helper functions

func slug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, "&", "")
	return s
}

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

func getActiveTag(currentPath, linkPath string) string {
	// Normalize paths
	currentPath = strings.TrimSuffix(currentPath, "/")
	linkPath = strings.TrimSuffix(linkPath, "/")
	
	// Exact match or prefix match for nested routes
	if currentPath == linkPath || (linkPath != "/" && strings.HasPrefix(currentPath, linkPath)) {
		return "active"
	}
	return ""
}

func getCurrentYear() int {
	return time.Now().Year()
}

func timeAgo(timestamp time.Time) string {
	now := time.Now()
	diff := now.Sub(timestamp)
	
	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes < 1 {
			return "just now"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	
	weeks := int(diff.Hours() / (24 * 7))
	if weeks == 1 {
		return "1 week ago"
	}
	return fmt.Sprintf("%d weeks ago", weeks)
}

func toTitle(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func formatActivityType(activityType string) string {
	switch strings.ToLower(activityType) {
	case "push":
		return "Push"
	case "pr":
		return "PR Opened"
	case "merge":
		return "PR Merged"
	case "star":
		return "Starred"
	case "fork":
		return "Forked"
	case "issue":
		return "Issue"
	case "release":
		return "Release"
	case "create":
		return "Created"
	case "delete":
		return "Deleted"
	case "comment":
		return "Comment"
	default:
		return toTitle(activityType)
	}
}