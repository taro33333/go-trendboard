package domain

import (
	"fmt"
	"strings"
)

// Repository represents a single GitHub repository being tracked.
type Repository struct {
	// FullName is the full name of the repository in "owner/name" format.
	FullName string
	// Stars is the current number of stars.
	Stars int
}

// NewRepository creates a new Repository object.
// It validates that the fullName is in the correct "owner/name" format.
func NewRepository(fullName string, stars int) (*Repository, error) {
	// A simple validation, can be improved with regex if needed.
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid repository full name format: %s", fullName)
	}
	if stars < 0 {
		return nil, fmt.Errorf("stars count cannot be negative: %d", stars)
	}
	return &Repository{
		FullName: fullName,
		Stars:    stars,
	}, nil
}
