package usecase

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yourname/go-trendboard/internal/config"
	"github.com/yourname/go-trendboard/internal/domain"
)

// --- Mocks ---

type MockFetcher struct {
	mock.Mock
}

func (m *MockFetcher) FetchStars(ctx context.Context, repoName string) (*domain.Repository, error) {
	args := m.Called(ctx, repoName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Repository), args.Error(1)
}

type MockStorer struct {
	mock.Mock
}

func (m *MockStorer) Save(date time.Time, repos []*domain.Repository) error {
	args := m.Called(date, repos)
	return args.Error(0)
}

func (m *MockStorer) Load(date time.Time) ([]*domain.Repository, error) {
	args := m.Called(date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Repository), args.Error(1)
}

func (m *MockStorer) LoadTargetRepos() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorer) SaveTargetRepos(repos []string) error {
	args := m.Called(repos)
	return args.Error(0)
}

// --- Tests ---

func setupTestUsecase(t *testing.T) (*Usecase, *MockFetcher, *MockStorer, *config.Config) {
	t.Helper()
	fetcher := new(MockFetcher)
	storer := new(MockStorer)
	logger := slog.New(slog.NewJSONHandler(os.NewFile(0, os.DevNull), nil))
	
	tempDir := t.TempDir()
	cfg := &config.Config{
		ReposFilePath:       filepath.Join(tempDir, "repos.json"),
		DashboardFilePath:   filepath.Join(tempDir, "dashboard.md"),
		DashboardTemplatePath: filepath.Join(tempDir, "dashboard.tpl"),
		DashboardFormat:     "md",
	}

	uc := NewUsecase(cfg, logger, fetcher, storer)
	return uc, fetcher, storer, cfg
}

func TestUsecase_Initialize(t *testing.T) {
	t.Run("File does not exist", func(t *testing.T) {
		uc, _, storer, _ := setupTestUsecase(t)
		storer.On("SaveTargetRepos", mock.AnythingOfType("[]string")).Return(nil).Once()
		
		err := uc.Initialize(context.Background())
		require.NoError(t, err)
		storer.AssertExpectations(t)
	})

	t.Run("File exists", func(t *testing.T) {
		uc, _, storer, cfg := setupTestUsecase(t)
		
		// Create the file to simulate it already existing
		_, err := os.Create(cfg.ReposFilePath)
		require.NoError(t, err)

		err = uc.Initialize(context.Background())
		require.NoError(t, err)
		storer.AssertNotCalled(t, "SaveTargetRepos", mock.Anything)
	})
}

func TestUsecase_Update(t *testing.T) {
	uc, fetcher, storer, _ := setupTestUsecase(t)

	targetRepos := []string{"owner/repo1", "owner/repo2"}
	repo1, _ := domain.NewRepository("owner/repo1", 100)
	repo2, _ := domain.NewRepository("owner/repo2", 200)

	storer.On("LoadTargetRepos").Return(targetRepos, nil).Once()
	fetcher.On("FetchStars", mock.Anything, "owner/repo1").Return(repo1, nil).Once()
	fetcher.On("FetchStars", mock.Anything, "owner/repo2").Return(repo2, nil).Once()
	storer.On("Save", mock.AnythingOfType("time.Time"), mock.AnythingOfType("[]*domain.Repository")).Return(nil).Once()

	err := uc.Update(context.Background())
	require.NoError(t, err)

	fetcher.AssertExpectations(t)
	storer.AssertExpectations(t)
}

func TestUsecase_Update_FetchFailure(t *testing.T) {
	uc, fetcher, storer, _ := setupTestUsecase(t)

	targetRepos := []string{"owner/repo1", "owner/repo2"}
	repo2, _ := domain.NewRepository("owner/repo2", 200)

	storer.On("LoadTargetRepos").Return(targetRepos, nil).Once()
	fetcher.On("FetchStars", mock.Anything, "owner/repo1").Return(nil, errors.New("fetch failed")).Once()
	fetcher.On("FetchStars", mock.Anything, "owner/repo2").Return(repo2, nil).Once()
	
	// Check that save is still called with the successfully fetched repo
	storer.On("Save", mock.AnythingOfType("time.Time"), mock.MatchedBy(func(repos []*domain.Repository) bool {
		return len(repos) == 1 && repos[0].FullName == "owner/repo2"
	})).Return(nil).Once()
	
	err := uc.Update(context.Background())
	require.NoError(t, err)

	fetcher.AssertExpectations(t)
	storer.AssertExpectations(t)
}


func TestUsecase_Generate(t *testing.T) {
	uc, _, storer, cfg := setupTestUsecase(t)

	today := time.Now().UTC()
	pastDate := today.AddDate(0, 0, -7)

	repo1, _ := domain.NewRepository("owner/repo1", 100)
	repo1Past, _ := domain.NewRepository("owner/repo1", 80)
	todayData := []*domain.Repository{repo1}
	pastData := []*domain.Repository{repo1Past}

	storer.On("Load", mock.MatchedBy(func(t time.Time) bool { return isSameDate(t, today) })).Return(todayData, nil).Once()
	storer.On("Load", mock.MatchedBy(func(t time.Time) bool { return isSameDate(t, pastDate) })).Return(pastData, nil).Once()

	err := uc.Generate(context.Background())
	require.NoError(t, err)

	// Check if the output file was created and contains content
	content, err := os.ReadFile(cfg.DashboardFilePath)
	require.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Contains(t, string(content), "owner/repo1")
	assert.Contains(t, string(content), "20 ★") // 100 - 80

	storer.AssertExpectations(t)
}

// isSameDate checks if two time.Time objects represent the same date (ignoring time).
func isSameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
