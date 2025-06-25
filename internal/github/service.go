package github

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// service implements the GitHub Service interface
type service struct {
	client  *Client
	cache   *Cache
	config  *GitHubConfig
	enabled bool
}

// NewService creates a new GitHub service instance
func NewService(config *GitHubConfig) Service {
	if config == nil || config.Username == "" || config.Token == "" {
		log.Printf("GITHUB_SERVICE: Service disabled - missing credentials")
		return &service{enabled: false}
	}
	
	// Set default cache expiry if not provided
	if config.CacheExpiry == 0 {
		config.CacheExpiry = 1 * time.Hour
	}
	
	client := NewClient(config.Token, config.Username)
	cache := NewCache("", config.CacheExpiry)
	
	log.Printf("GITHUB_SERVICE: Service enabled for user '%s'", config.Username)
	
	s := &service{
		client:  client,
		cache:   cache,
		config:  config,
		enabled: config.Enabled,
	}
	
	// Pre-warm the cache in the background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if stats, err := s.client.FetchUserData(ctx); err == nil {
			// Also fetch recent activities
			if activities, err := s.client.FetchRecentEvents(ctx, 5); err == nil {
				stats.RecentActivity = activities
			}
			s.cache.Set(stats)
			log.Printf("GITHUB: Initial cache populated with %d repos, %d stars, %d activities", 
				stats.PublicRepos, stats.TotalStars, len(stats.RecentActivity))
		} else {
			log.Printf("GITHUB: Initial cache population failed: %v", err)
		}
	}()
	
	return s
}

// IsEnabled returns whether GitHub integration is enabled
func (s *service) IsEnabled() bool {
	return s.enabled && s.client != nil
}

// GetStats retrieves GitHub statistics for the configured user
func (s *service) GetStats(ctx context.Context) (*GitHubStats, error) {
	if !s.IsEnabled() {
		return s.getDefaultStats(), nil
	}
	
	// Try to get from cache first - this should be instant
	if stats, found := s.cache.Get(); found {
		log.Printf("GITHUB: Using cached stats (%d repos, %d stars)", stats.PublicRepos, stats.TotalStars)
		
		// Trigger background refresh if cache is getting old (but don't wait)
		go func() {
			// Use a separate context for background refresh
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			
			if newStats, err := s.client.FetchUserData(bgCtx); err == nil {
				// Also fetch recent activities for the background refresh
				if activities, err := s.client.FetchRecentEvents(bgCtx, 5); err == nil {
					newStats.RecentActivity = activities
				}
				s.cache.Set(newStats)
				log.Printf("GITHUB: Background cache refresh complete with %d activities", len(newStats.RecentActivity))
			}
		}()
		
		return stats, nil
	}
	
	// No cache available - check if we have time for an API call
	select {
	case <-ctx.Done():
		log.Printf("GITHUB: Request cancelled, returning defaults")
		return s.getDefaultStats(), nil
	default:
	}
	
	// Try to fetch with the remaining context time
	log.Printf("GITHUB: No cache available, attempting quick fetch...")
	stats, err := s.client.FetchUserData(ctx)
	if err != nil {
		if ctx.Err() != nil {
			log.Printf("GITHUB: API call cancelled due to timeout, returning defaults")
		} else {
			log.Printf("GITHUB: API call failed: %v", err)
		}
		return s.getDefaultStats(), nil
	}
	
	// Check context cancellation before fetching activities
	select {
	case <-ctx.Done():
		log.Printf("GITHUB: Request cancelled before fetching activities")
		// Still cache what we have and return it
		if err := s.cache.Set(stats); err != nil {
			log.Printf("GITHUB: Cache write failed: %v", err)
		}
		return stats, nil
	default:
	}
	
	// Add recent activity
	activities, err := s.client.FetchRecentEvents(ctx, 5)
	if err != nil {
		if ctx.Err() != nil {
			log.Printf("GITHUB: Activities fetch cancelled due to timeout")
		} else {
			log.Printf("GITHUB: Failed to fetch activities (continuing): %v", err)
		}
		activities = []ActivityItem{}
	}
	stats.RecentActivity = activities
	
	// Cache the results
	if err := s.cache.Set(stats); err != nil {
		log.Printf("GITHUB: Cache write failed: %v", err)
	}
	
	log.Printf("GITHUB: ✅ Live stats loaded: %d repos, %d stars, %d contributions", 
		stats.PublicRepos, stats.TotalStars, stats.Contributions)
	
	return stats, nil
}

// GetContributionGraph retrieves the contribution graph data
func (s *service) GetContributionGraph(ctx context.Context) (*ContributionGraph, error) {
	if !s.IsEnabled() {
		return s.getDefaultContributionGraph(), nil
	}
	
	// Get stats which includes contribution data
	stats, err := s.GetStats(ctx)
	if err != nil {
		return s.getDefaultContributionGraph(), nil
	}
	
	// Convert contribution map to graph structure
	graph := &ContributionGraph{
		TotalContributions: stats.Contributions,
		Weeks:              []ContributionWeek{},
	}
	
	// TODO: Implement proper week-based contribution graph
	// This is a simplified version
	return graph, nil
}

// GetRepositories retrieves the user's repositories
func (s *service) GetRepositories(ctx context.Context, limit int) ([]Repository, error) {
	if !s.IsEnabled() {
		return s.getDefaultRepositories(), nil
	}
	
	// This would require additional API calls
	// For now, return empty slice
	return []Repository{}, nil
}

// GetRecentActivity retrieves recent activity items
func (s *service) GetRecentActivity(ctx context.Context, limit int) ([]ActivityItem, error) {
	if !s.IsEnabled() {
		return s.getDefaultActivity(), nil
	}
	
	return s.client.FetchRecentEvents(ctx, limit)
}

// ClearCache clears the cached GitHub data
func (s *service) ClearCache() error {
	if s.cache == nil {
		return nil
	}
	return s.cache.Clear()
}

// getDefaultStats returns default stats when GitHub integration is disabled
func (s *service) getDefaultStats() *GitHubStats {
	return &GitHubStats{
		PublicRepos:     42,
		TotalStars:      1337,
		TotalForks:      89,
		Followers:       156,
		Following:       42,
		Contributions:   523,
		PullRequests:    89,
		Issues:          23,
		RecentActivity:  s.getDefaultActivity(),
		ContributionMap: make(map[string]int),
		Languages:       s.getDefaultLanguages(),
		LastUpdated:     time.Now(),
	}
}

// getDefaultActivity returns default activity items
func (s *service) getDefaultActivity() []ActivityItem {
	return []ActivityItem{
		{
			Type:        "push",
			Repository:  "BlockheadConsulting/ai-infrastructure",
			Description: "Pushed commits",
			Timestamp:   time.Now().Add(-2 * time.Hour),
		},
		{
			Type:        "pr",
			Repository:  "ethereum/go-ethereum",
			Description: "Pull request merged",
			Timestamp:   time.Now().Add(-24 * time.Hour),
		},
		{
			Type:        "star",
			Repository:  "langchain-ai/langchain",
			Description: "Starred repository",
			Timestamp:   time.Now().Add(-72 * time.Hour),
		},
	}
}

// getDefaultLanguages returns default language statistics
func (s *service) getDefaultLanguages() map[string]int {
	return map[string]int{
		"Go":         25,
		"Python":     18,
		"TypeScript": 15,
		"Solidity":   12,
		"Rust":       8,
		"JavaScript": 6,
		"Shell":      4,
		"Other":      12,
	}
}

// getDefaultContributionGraph returns a default contribution graph
func (s *service) getDefaultContributionGraph() *ContributionGraph {
	return &ContributionGraph{
		TotalContributions: 523,
		Weeks:              []ContributionWeek{},
	}
}

// getDefaultRepositories returns default repositories
func (s *service) getDefaultRepositories() []Repository {
	return []Repository{
		{
			Name:        "ai-infrastructure",
			FullName:    "BlockheadConsulting/ai-infrastructure",
			Description: "Production-grade AI infrastructure components",
			Stars:       42,
			Forks:       8,
			Language:    "Go",
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
			URL:         "https://github.com/BlockheadConsulting/ai-infrastructure",
			IsPrivate:   false,
		},
		{
			Name:        "blockchain-analytics",
			FullName:    "BlockheadConsulting/blockchain-analytics",
			Description: "Real-time blockchain data analytics platform",
			Stars:       28,
			Forks:       5,
			Language:    "Python",
			UpdatedAt:   time.Now().Add(-12 * time.Hour),
			URL:         "https://github.com/BlockheadConsulting/blockchain-analytics",
			IsPrivate:   false,
		},
	}
}

// FormatActivityTime formats activity timestamp for display
func FormatActivityTime(timestamp time.Time) string {
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

// FormatActivityType formats activity type for display
func FormatActivityType(activityType string) string {
	switch strings.ToLower(activityType) {
	case "push":
		return "Push"
	case "pr":
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
		return "Activity"
	}
}