package github

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"

	"github.com/yourname/go-trendboard/internal/config"
	"github.com/yourname/go-trendboard/internal/domain"
)

// Client is a GitHub API client that implements the Fetcher interface.
type Client struct {
	client *github.Client
	logger *slog.Logger
}

// NewClient creates a new instance of the GitHub API client.
// It requires a configuration object for the API token and a logger.
func NewClient(cfg *config.Config, logger *slog.Logger) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Client{
		client: client,
		logger: logger.With("component", "github_client"),
	}
}

// FetchStars fetches the star count for a given repository from the GitHub API.
func (c *Client) FetchStars(ctx context.Context, repoName string) (*domain.Repository, error) {
	c.logger.Debug("Fetching stars", "repo", repoName)

	parts := strings.Split(repoName, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return domain.NewRepository(repoName, 0)
	}
	owner, repo := parts[0], parts[1]

	ghRepo, resp, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		// Handle rate limit error specifically
		if _, ok := err.(*github.RateLimitError); ok {
			c.logger.Warn("GitHub API rate limit exceeded", "repo", repoName, "response", resp)
			return nil, &RateLimitError{repoName: repoName, cause: err}
		}
		c.logger.Error("Failed to fetch repository from GitHub API", "repo", repoName, "error", err)
		return nil, &FetchError{repoName: repoName, cause: err}
	}

	stars := ghRepo.GetStargazersCount()
	c.logger.Debug("Successfully fetched stars", "repo", repoName, "stars", stars)

	return domain.NewRepository(repoName, stars)
}
