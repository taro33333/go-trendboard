package github

import "fmt"

// RateLimitError is returned when the GitHub API rate limit is exceeded.
type RateLimitError struct {
	repoName string
	cause    error
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded while fetching repository '%s'", e.repoName)
}

func (e *RateLimitError) Unwrap() error {
	return e.cause
}

// FetchError is returned for general errors during GitHub API fetches.
type FetchError struct {
	repoName string
	cause    error
}

func (e *FetchError) Error() string {
	return fmt.Sprintf("failed to fetch repository '%s'", e.repoName)
}

func (e *FetchError) Unwrap() error {
	return e.cause
}
