package templates

import (
	"html/template"
)

// Manager handles template loading and management
type Manager struct {
	templates *template.Template
	funcMap   template.FuncMap
}

// Config holds template configuration
type Config struct {
	Patterns []string
}