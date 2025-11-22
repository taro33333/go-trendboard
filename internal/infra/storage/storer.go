package storage

import (
	"errors"
	"time"

	"github.com/yourname/go-trendboard/internal/domain"
)

var (
	// ErrDataNotFound is returned when trend data for a specific date is not found.
	ErrDataNotFound = errors.New("trend data not found for the specified date")
	// ErrReposConfigNotFound is returned when the repos.json file is not found.
	ErrReposConfigNotFound = errors.New("repositories config file not found")
)

// Storer defines the interface for persisting and retrieving trend data.
type Storer interface {
	// Save saves the list of repositories for a specific date.
	Save(date time.Time, repos []*domain.Repository) error

	// Load loads the list of repositories for a specific date.
	Load(date time.Time) ([]*domain.Repository, error)

	// LoadTargetRepos loads the list of target repository names from the configuration.
	LoadTargetRepos() ([]string, error)

	// SaveTargetRepos saves the list of target repository names to the configuration.
	// This is used for the 'init' command.
	SaveTargetRepos(repos []string) error
}
