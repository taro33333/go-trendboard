package github

import (
	"context"

	"github.com/yourname/go-trendboard/internal/domain"
)

// Fetcher defines the interface for fetching repository data from a source like GitHub.
type Fetcher interface {
	// FetchStars fetches the star count for a given repository.
	// The repoName is expected to be in "owner/name" format.
	FetchStars(ctx context.Context, repoName string) (*domain.Repository, error)
}
