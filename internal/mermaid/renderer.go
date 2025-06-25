package mermaid

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// Renderer provides Mermaid diagram rendering functionality
type Renderer struct {
	cache      map[string]string
	cacheMutex sync.RWMutex
	tempDir    string
	logger     *log.Logger
}

// NewRenderer creates a new Mermaid renderer
func NewRenderer(tempDir string, logger *log.Logger) *Renderer {
	return &Renderer{
		cache:   make(map[string]string),
		tempDir: tempDir,
		logger:  logger,
	}
}

// RenderServerSide attempts to render a Mermaid diagram on the server
func (r *Renderer) RenderServerSide(mermaidCode string) string {
	// Generate cache key from content hash
	hash := md5.Sum([]byte(mermaidCode))
	cacheKey := hex.EncodeToString(hash[:])
	
	// Check cache first
	r.cacheMutex.RLock()
	if cached, exists := r.cache[cacheKey]; exists {
		r.cacheMutex.RUnlock()
		return cached
	}
	r.cacheMutex.RUnlock()
	
	// Create temporary directory for Mermaid CLI
	os.MkdirAll(r.tempDir, 0755)
	
	inputFile := filepath.Join(r.tempDir, cacheKey+".mmd")
	outputFile := filepath.Join(r.tempDir, cacheKey+".svg")
	
	// Write Mermaid input file
	if err := os.WriteFile(inputFile, []byte(mermaidCode), 0644); err != nil {
		r.logger.Printf("Failed to write Mermaid input file: %v", err)
		return ""
	}
	
	// Try to execute mmdc (Mermaid CLI) with timeout - using simple dark theme for now
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "mmdc", 
		"-i", inputFile, 
		"-o", outputFile,
		"-t", "dark",
		"-b", "transparent")
	
	// Capture both stdout and stderr for debugging
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		r.logger.Printf("Mermaid CLI failed (falling back to client-side):")
		r.logger.Printf("  Error: %v", err)
		r.logger.Printf("  Command: mmdc -i %s -o %s -t dark -b transparent", inputFile, outputFile)
		r.logger.Printf("  Stdout: %s", stdout.String())
		r.logger.Printf("  Stderr: %s", stderr.String())
		r.logger.Printf("  Input file exists: %v", fileExists(inputFile))
		r.logger.Printf("  Output dir exists: %v", fileExists(filepath.Dir(outputFile)))
		// Clean up temp files
		os.Remove(inputFile)
		os.Remove(outputFile)
		return ""
	}
	
	// Read generated SVG
	svgContent, err := os.ReadFile(outputFile)
	if err != nil {
		r.logger.Printf("Failed to read generated SVG: %v", err)
		os.Remove(inputFile)
		os.Remove(outputFile)
		return ""
	}
	
	// Clean up temp files
	os.Remove(inputFile)
	os.Remove(outputFile)
	
	svgString := string(svgContent)
	
	// Cache the result
	r.cacheMutex.Lock()
	r.cache[cacheKey] = svgString
	r.cacheMutex.Unlock()
	
	return svgString
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}