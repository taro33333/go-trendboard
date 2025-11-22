package storage

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourname/go-trendboard/internal/config"
	"github.com/yourname/go-trendboard/internal/domain"
)

// setupTestStorer creates a FileStorer instance for testing, using a temporary directory.
func setupTestStorer(t *testing.T) (*FileStorer, *config.Config) {
	t.Helper()
	tempDir := t.TempDir()

	cfg := &config.Config{
		DataDirPath:   filepath.Join(tempDir, "data"),
		ReposFilePath: filepath.Join(tempDir, "repos.json"),
	}
	logger := slog.New(slog.NewJSONHandler(os.NewFile(0, os.DevNull), nil)) // Discard logs

	return NewFileStorer(cfg, logger), cfg
}

func TestFileStorer_SaveAndLoad(t *testing.T) {
	storer, _ := setupTestStorer(t)
	date := time.Date(2025, 11, 22, 0, 0, 0, 0, time.UTC)

	// Data to save
	repo1, _ := domain.NewRepository("owner/repo1", 100)
	repo2, _ := domain.NewRepository("owner/repo2", 200)
	reposToSave := []*domain.Repository{repo1, repo2}

	// 1. Test Save
	err := storer.Save(date, reposToSave)
	require.NoError(t, err)

	// 2. Test Load
	loadedRepos, err := storer.Load(date)
	require.NoError(t, err)
	require.Len(t, loadedRepos, 2)

	assert.Equal(t, "owner/repo1", loadedRepos[0].FullName)
	assert.Equal(t, 100, loadedRepos[0].Stars)
	assert.Equal(t, "owner/repo2", loadedRepos[1].FullName)
	assert.Equal(t, 200, loadedRepos[1].Stars)
}

func TestFileStorer_Load_NotFound(t *testing.T) {
	storer, _ := setupTestStorer(t)
	date := time.Date(2025, 11, 23, 0, 0, 0, 0, time.UTC)

	_, err := storer.Load(date)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDataNotFound)
}

func TestFileStorer_SaveAndLoadTargetRepos(t *testing.T) {
	storer, _ := setupTestStorer(t)
	targetRepos := []string{"gin-gonic/gin", "go-chi/chi"}

	// 1. Test Save
	err := storer.SaveTargetRepos(targetRepos)
	require.NoError(t, err)

	// 2. Test Load
	loadedTargetRepos, err := storer.LoadTargetRepos()
	require.NoError(t, err)
	assert.Equal(t, targetRepos, loadedTargetRepos)
}

func TestFileStorer_LoadTargetRepos_NotFound(t *testing.T) {
	storer, _ := setupTestStorer(t)

	_, err := storer.LoadTargetRepos()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrReposConfigNotFound)
}

func TestFileStorer_Load_InvalidJSON(t *testing.T) {
	storer, cfg := setupTestStorer(t)
	date := time.Date(2025, 11, 24, 0, 0, 0, 0, time.UTC)
	dataPath := storer.getDailyDataPath(date)

	// Create an invalid JSON file
	require.NoError(t, os.MkdirAll(cfg.DataDirPath, 0755))
	err := os.WriteFile(dataPath, []byte(`[{"name": "bad-json",]`), 0644)
	require.NoError(t, err)

	_, err = storer.Load(date)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not decode JSON data")
}

func TestFileStorer_LoadTargetRepos_InvalidJSON(t *testing.T) {
	storer, cfg := setupTestStorer(t)

	// Create an invalid JSON file
	err := os.WriteFile(cfg.ReposFilePath, []byte(`["repo1", "repo2",]`), 0644)
	require.NoError(t, err)

	_, err = storer.LoadTargetRepos()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not decode repos config file")
}
