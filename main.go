package main

import (
	"embed"
	"log"
	"os"

	"blockhead.consulting/internal/app"
)

//go:embed templates
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

//go:embed content/blog/**/*.md content/blog.yml
var blogFS embed.FS

func main() {
	// Create new application instance
	application := app.New(templateFS, staticFS, blogFS)

	// Initialize all dependencies
	if err := application.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Create handler
	if err := application.CreateHandler(); err != nil {
		log.Fatalf("Failed to create handler: %v", err)
	}

	// Run the application
	if err := application.Run(); err != nil {
		log.Printf("Application error: %v", err)
		os.Exit(1)
	}
}

