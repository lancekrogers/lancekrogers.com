package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"blockhead.consulting/internal/github"
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

	fmt.Printf("Testing GitHub API connectivity...\n")
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Token: %s\n", func() string {
		if token != "" {
			return "***provided***"
		}
		return "NOT PROVIDED"
	}())
	fmt.Printf("---\n")

	// Create GitHub config
	config := &github.GitHubConfig{
		Username:    username,
		Token:       token,
		CacheExpiry: 5 * time.Minute,
		Enabled:     token != "",
	}

	// Create service
	service := github.NewService(config)

	fmt.Printf("Service enabled: %v\n", service.IsEnabled())

	// Test getting stats
	ctx := context.Background()
	stats, err := service.GetStats(ctx)
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		return
	}

	fmt.Printf("---\nGitHub Stats Results:\n")
	fmt.Printf("Public Repos: %d\n", stats.PublicRepos)
	fmt.Printf("Stars Received (on your repos): %d\n", stats.StarsReceived)
	fmt.Printf("Total Stars (repos you starred): %d\n", stats.TotalStars)
	fmt.Printf("Total Forks: %d\n", stats.TotalForks)
	fmt.Printf("Followers: %d\n", stats.Followers)
	fmt.Printf("Following: %d\n", stats.Following)
	fmt.Printf("Contributions: %d\n", stats.Contributions)
	fmt.Printf("Pull Requests: %d\n", stats.PullRequests)
	fmt.Printf("Issues: %d\n", stats.Issues)
	fmt.Printf("Last Updated: %s\n", stats.LastUpdated.Format(time.RFC3339))

	// Check if we're getting default values
	if stats.PublicRepos == 42 && stats.TotalStars == 1337 && stats.Contributions == 523 {
		fmt.Printf("\n⚠️  WARNING: These appear to be DEFAULT FALLBACK VALUES\n")
		fmt.Printf("This means the GitHub API is not being called successfully.\n")
		if token == "" {
			fmt.Printf("SOLUTION: Set GITHUB_TOKEN environment variable\n")
		}
	} else {
		fmt.Printf("\n✅ SUCCESS: Live GitHub data retrieved!\n")
	}

	fmt.Printf("\nRecent Activity:\n")
	for i, activity := range stats.RecentActivity {
		fmt.Printf("%d. [%s] %s - %s (%s)\n", 
			i+1, 
			activity.Type, 
			activity.Repository, 
			activity.Description,
			activity.Timestamp.Format("2006-01-02"))
	}

	// Also test fetching recent events directly
	fmt.Printf("\n---\nDirect Event Fetch Test:\n")
	client := github.NewClient(token, username)
	events, err := client.FetchRecentEvents(ctx, 10)
	if err != nil {
		fmt.Printf("Error fetching events: %v\n", err)
	} else {
		fmt.Printf("Found %d events\n", len(events))
		for i, event := range events {
			fmt.Printf("%d. [%s] %s - %s (%s)\n",
				i+1,
				event.Type,
				event.Repository,
				event.Description,
				event.Timestamp.Format("2006-01-02 15:04"))
		}
	}

	fmt.Printf("\nLanguages:\n")
	for lang, count := range stats.Languages {
		fmt.Printf("- %s: %d\n", lang, count)
	}
}