package github

import "time"

// GitHubStats represents GitHub statistics for a user
type GitHubStats struct {
	PublicRepos     int               `json:"public_repos"`
	TotalStars      int               `json:"total_stars"`      // Stars on projects the user has starred
	StarsReceived   int               `json:"stars_received"`    // Stars received on user's own repos
	TotalForks      int               `json:"total_forks"`
	Followers       int               `json:"followers"`
	Following       int               `json:"following"`
	Contributions   int               `json:"contributions"`
	PullRequests    int               `json:"pull_requests"`
	Issues          int               `json:"issues"`
	RecentActivity  []ActivityItem    `json:"recent_activity"`
	ContributionMap map[string]int    `json:"contribution_map"`
	Languages       map[string]int    `json:"languages"`
	LastUpdated     time.Time         `json:"last_updated"`
}

// ActivityItem represents a single GitHub activity item
type ActivityItem struct {
	Type        string    `json:"type"`
	Repository  string    `json:"repository"`
	Title       string    `json:"title,omitempty"`
	URL         string    `json:"url,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description,omitempty"`
}

// ContributionDay represents a single day's contribution count
type ContributionDay struct {
	Date         string `json:"date"`
	Count        int    `json:"count"`
	Level        int    `json:"level"`
}

// ContributionWeek represents a week of contributions
type ContributionWeek struct {
	Days []ContributionDay `json:"days"`
}

// ContributionGraph represents the full contribution graph
type ContributionGraph struct {
	Weeks               []ContributionWeek `json:"weeks"`
	TotalContributions  int                `json:"total_contributions"`
}

// Repository represents a GitHub repository
type Repository struct {
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	Stars           int       `json:"stars"`
	Forks           int       `json:"forks"`
	Language        string    `json:"language"`
	UpdatedAt       time.Time `json:"updated_at"`
	URL             string    `json:"url"`
	IsPrivate       bool      `json:"is_private"`
}

// GitHubConfig holds configuration for GitHub API integration
type GitHubConfig struct {
	Username    string
	Token       string
	CacheExpiry time.Duration
	Enabled     bool
}

// CacheEntry represents a cached GitHub data entry
type CacheEntry struct {
	Data      *GitHubStats `json:"data"`
	Timestamp time.Time    `json:"timestamp"`
	ExpiresAt time.Time    `json:"expires_at"`
}