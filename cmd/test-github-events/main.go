package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	godotenv.Load()
	
	username := os.Getenv("GITHUB_USERNAME")
	token := os.Getenv("GITHUB_TOKEN")

	if username == "" {
		username = "lancekrogers"
	}

	fmt.Printf("Fetching raw GitHub events for %s...\n\n", username)

	// Make direct API call to see raw events
	url := fmt.Sprintf("https://api.github.com/users/%s/events/public?per_page=10", username)
	
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("User-Agent", "BlockheadConsulting-Test/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	
	var events []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Found %d events\n\n", len(events))
	
	// Analyze each event
	for i, event := range events {
		eventType := event["type"].(string)
		repo := event["repo"].(map[string]interface{})
		repoName := repo["name"].(string)
		createdAt := event["created_at"].(string)
		
		fmt.Printf("Event %d:\n", i+1)
		fmt.Printf("  Type: %s\n", eventType)
		fmt.Printf("  Repo: %s\n", repoName)
		fmt.Printf("  Time: %s\n", createdAt)
		
		if payload, ok := event["payload"].(map[string]interface{}); ok {
			fmt.Printf("  Payload:\n")
			
			// Common payload fields
			if action, ok := payload["action"].(string); ok {
				fmt.Printf("    Action: %s\n", action)
			}
			
			// CreateEvent specific
			if refType, ok := payload["ref_type"].(string); ok {
				fmt.Printf("    Ref Type: %s\n", refType)
			}
			if ref, ok := payload["ref"].(string); ok {
				fmt.Printf("    Ref: %s\n", ref)
			}
			
			// PullRequestEvent specific
			if pr, ok := payload["pull_request"].(map[string]interface{}); ok {
				if title, ok := pr["title"].(string); ok {
					fmt.Printf("    PR Title: %s\n", title)
				}
				if merged, ok := pr["merged"].(bool); ok {
					fmt.Printf("    Merged: %v\n", merged)
				}
				if state, ok := pr["state"].(string); ok {
					fmt.Printf("    State: %s\n", state)
				}
			}
			
			// DeleteEvent specific
			if ref, ok := payload["ref"].(string); ok && eventType == "DeleteEvent" {
				fmt.Printf("    Deleted Ref: %s\n", ref)
			}
			
			// PushEvent specific
			if commits, ok := payload["commits"].([]interface{}); ok {
				fmt.Printf("    Commits: %d\n", len(commits))
			}
		}
		
		fmt.Println()
	}
}