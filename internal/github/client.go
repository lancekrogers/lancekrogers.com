package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	githubAPIURL    = "https://api.github.com"
	githubGraphQLURL = "https://api.github.com/graphql"
	userAgent       = "BlockheadConsulting-Website/1.0"
	requestTimeout  = 30 * time.Second
)

// Client handles GitHub API interactions
type Client struct {
	httpClient *http.Client
	token      string
	username   string
}

// NewClient creates a new GitHub API client
func NewClient(token, username string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
		token:    token,
		username: username,
	}
}

// GraphQL query for contribution data
const contributionQuery = `
query($username: String!) {
  user(login: $username) {
    contributionsCollection {
      contributionCalendar {
        totalContributions
        weeks {
          contributionDays {
            contributionCount
            date
            contributionLevel
          }
        }
      }
    }
    repositories(first: 100, ownerAffiliations: OWNER, orderBy: {field: UPDATED_AT, direction: DESC}) {
      totalCount
      nodes {
        name
        description
        stargazerCount
        forkCount
        primaryLanguage {
          name
        }
        updatedAt
        url
        isPrivate
      }
    }
    starredRepositories {
      totalCount
    }
    pullRequests(first: 100, states: [OPEN, CLOSED, MERGED]) {
      totalCount
    }
    issues(first: 100, states: [OPEN, CLOSED]) {
      totalCount
    }
    followers {
      totalCount
    }
    following {
      totalCount
    }
  }
}
`

// GraphQL request structure
type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// GraphQL response structures
type graphQLResponse struct {
	Data   graphQLData `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

type graphQLData struct {
	User struct {
		ContributionsCollection struct {
			ContributionCalendar struct {
				TotalContributions int `json:"totalContributions"`
				Weeks              []struct {
					ContributionDays []struct {
						ContributionCount int    `json:"contributionCount"`
						Date              string `json:"date"`
						ContributionLevel string `json:"contributionLevel"`
					} `json:"contributionDays"`
				} `json:"weeks"`
			} `json:"contributionCalendar"`
		} `json:"contributionsCollection"`
		Repositories struct {
			TotalCount int `json:"totalCount"`
			Nodes      []struct {
				Name         string `json:"name"`
				Description  string `json:"description"`
				StarCount    int    `json:"stargazerCount"`
				ForkCount    int    `json:"forkCount"`
				Language     *struct {
					Name string `json:"name"`
				} `json:"primaryLanguage"`
				UpdatedAt string `json:"updatedAt"`
				URL       string `json:"url"`
				IsPrivate bool   `json:"isPrivate"`
			} `json:"nodes"`
		} `json:"repositories"`
		StarredRepositories struct {
			TotalCount int `json:"totalCount"`
		} `json:"starredRepositories"`
		PullRequests struct {
			TotalCount int `json:"totalCount"`
		} `json:"pullRequests"`
		Issues struct {
			TotalCount int `json:"totalCount"`
		} `json:"issues"`
		Followers struct {
			TotalCount int `json:"totalCount"`
		} `json:"followers"`
		Following struct {
			TotalCount int `json:"totalCount"`
		} `json:"following"`
	} `json:"user"`
}

// FetchUserData fetches comprehensive user data from GitHub GraphQL API
func (c *Client) FetchUserData(ctx context.Context) (*GitHubStats, error) {
	// Prepare GraphQL request
	reqBody := graphQLRequest{
		Query:     contributionQuery,
		Variables: map[string]interface{}{
			"username": c.username,
		},
	}
	
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", githubGraphQLURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	
	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GraphQL request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	var graphqlResp graphQLResponse
	if err := json.Unmarshal(body, &graphqlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GraphQL response: %w", err)
	}
	
	// Check for GraphQL errors
	if len(graphqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", graphqlResp.Errors)
	}
	
	// Convert to GitHubStats
	return c.convertToGitHubStats(graphqlResp.Data), nil
}

// convertToGitHubStats converts GraphQL response to GitHubStats
func (c *Client) convertToGitHubStats(data graphQLData) *GitHubStats {
	stats := &GitHubStats{
		PublicRepos:     0, // Will count non-private repos
		TotalStars:      data.User.StarredRepositories.TotalCount, // Count of starred repositories
		StarsReceived:   0, // Will sum up stars on user's own repos
		TotalForks:      0,
		Followers:       data.User.Followers.TotalCount,
		Following:       data.User.Following.TotalCount,
		Contributions:   data.User.ContributionsCollection.ContributionCalendar.TotalContributions,
		PullRequests:    data.User.PullRequests.TotalCount,
		Issues:          data.User.Issues.TotalCount,
		RecentActivity:  []ActivityItem{}, // TODO: Implement activity parsing
		ContributionMap: make(map[string]int),
		Languages:       make(map[string]int),
		LastUpdated:     time.Now(),
	}
	
	// Process repositories
	for _, repo := range data.User.Repositories.Nodes {
		if !repo.IsPrivate {
			stats.PublicRepos++
		}
		stats.StarsReceived += repo.StarCount // Sum stars received on user's repos
		stats.TotalForks += repo.ForkCount
		
		// Count languages
		if repo.Language != nil && repo.Language.Name != "" {
			stats.Languages[repo.Language.Name]++
		}
	}
	
	// Process contribution calendar
	for _, week := range data.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			if day.ContributionCount > 0 {
				stats.ContributionMap[day.Date] = day.ContributionCount
			}
		}
	}
	
	return stats
}

// FetchRecentEvents fetches recent public events for the user
func (c *Client) FetchRecentEvents(ctx context.Context, limit int) ([]ActivityItem, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	
	url := fmt.Sprintf("%s/users/%s/events/public?per_page=%d", githubAPIURL, c.username, limit)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create events request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read events response: %w", err)
	}
	
	var events []map[string]interface{}
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}
	
	// Convert to ActivityItems
	activities := make([]ActivityItem, 0, len(events))
	for _, event := range events {
		activity := c.convertEventToActivity(event)
		if activity != nil {
			activities = append(activities, *activity)
		}
	}
	
	return activities, nil
}

// convertEventToActivity converts a GitHub event to an ActivityItem
func (c *Client) convertEventToActivity(event map[string]interface{}) *ActivityItem {
	eventType, ok := event["type"].(string)
	if !ok {
		return nil
	}
	
	repo, ok := event["repo"].(map[string]interface{})
	if !ok {
		return nil
	}
	
	repoName, ok := repo["name"].(string)
	if !ok {
		return nil
	}
	
	createdAt, ok := event["created_at"].(string)
	if !ok {
		return nil
	}
	
	timestamp, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil
	}
	
	activity := &ActivityItem{
		Repository: repoName,
		Timestamp:  timestamp,
	}
	
	// Map GitHub event types to our activity types
	payload, _ := event["payload"].(map[string]interface{})
	
	switch eventType {
	case "PushEvent":
		activity.Type = "push"
		// Only show pushes to main/master branches
		if ref, ok := payload["ref"].(string); ok {
			if strings.HasSuffix(ref, "/main") || strings.HasSuffix(ref, "/master") {
				activity.Description = "Pushed to main branch"
			} else {
				// Skip non-main branch pushes
				return nil
			}
		}
	case "PullRequestEvent":
		if action, ok := payload["action"].(string); ok {
			switch action {
			case "opened":
				activity.Type = "pr"
				activity.Description = "Opened pull request"
			case "closed":
				if pr, ok := payload["pull_request"].(map[string]interface{}); ok {
					if merged, ok := pr["merged"].(bool); ok && merged {
						activity.Type = "merge"
						activity.Description = "Merged pull request"
					} else {
						// Skip closed but not merged PRs
						return nil
					}
				}
			default:
				// Skip other PR actions (edited, labeled, etc.)
				return nil
			}
		}
	case "IssuesEvent":
		if action, ok := payload["action"].(string); ok {
			if action == "opened" || action == "closed" {
				activity.Type = "issue"
				activity.Description = fmt.Sprintf("Issue %s", action)
			} else {
				// Skip other issue actions
				return nil
			}
		}
	case "WatchEvent":
		activity.Type = "star"
		activity.Description = "Starred repository"
	case "ForkEvent":
		activity.Type = "fork"
		activity.Description = "Forked repository"
	case "CreateEvent":
		// Only show repository creation, not branch creation
		if refType, ok := payload["ref_type"].(string); ok {
			if refType == "repository" {
				activity.Type = "create"
				activity.Description = "Created repository"
			} else {
				// Skip branch/tag creation
				return nil
			}
		}
	case "ReleaseEvent":
		activity.Type = "release"
		activity.Description = "Published release"
	case "DeleteEvent":
		// Skip branch/tag deletions as they're usually part of PR workflow
		return nil
	case "IssueCommentEvent", "PullRequestReviewCommentEvent":
		// Skip comment events to reduce noise
		return nil
	default:
		// Skip unknown event types
		return nil
	}
	
	return activity
}