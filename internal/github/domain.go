package github

import "context"

// Service defines the interface for GitHub service operations
type Service interface {
	// GetStats retrieves GitHub statistics for the configured user
	GetStats(ctx context.Context) (*GitHubStats, error)
	
	// GetContributionGraph retrieves the contribution graph data
	GetContributionGraph(ctx context.Context) (*ContributionGraph, error)
	
	// GetRepositories retrieves the user's repositories
	GetRepositories(ctx context.Context, limit int) ([]Repository, error)
	
	// GetRecentActivity retrieves recent activity items
	GetRecentActivity(ctx context.Context, limit int) ([]ActivityItem, error)
	
	// IsEnabled returns whether GitHub integration is enabled
	IsEnabled() bool
	
	// ClearCache clears the cached GitHub data
	ClearCache() error
}