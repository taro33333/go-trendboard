package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortTrends(t *testing.T) {
	t.Parallel()

	// Helper to create repositories without error handling for tests
	mustNewRepo := func(fullName string, stars int) *Repository {
		repo, _ := NewRepository(fullName, stars)
		return repo
	}

	trends := []*Trend{
		NewTrend(mustNewRepo("owner/repo-c", 100), 10, TrendDaily),
		NewTrend(mustNewRepo("owner/repo-a", 300), 50, TrendDaily),
		NewTrend(mustNewRepo("owner/repo-b", 200), -5, TrendDaily),
		NewTrend(mustNewRepo("owner/repo-d", 500), 25, TrendDaily),
	}

	expectedOrder := []string{"owner/repo-a", "owner/repo-d", "owner/repo-c", "owner/repo-b"}

	SortTrends(trends)

	for i, trend := range trends {
		assert.Equal(t, expectedOrder[i], trend.Repository.FullName)
	}
}
