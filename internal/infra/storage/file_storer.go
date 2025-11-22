package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/yourname/go-trendboard/internal/config"
	"github.com/yourname/go-trendboard/internal/domain"
)

// FileStorer implements the Storer interface using the local file system.
type FileStorer struct {
	cfg    *config.Config
	logger *slog.Logger
}

// NewFileStorer creates a new FileStorer.
func NewFileStorer(cfg *config.Config, logger *slog.Logger) *FileStorer {
	return &FileStorer{
		cfg:    cfg,
		logger: logger.With("component", "file_storer"),
	}
}

// getDailyDataPath returns the path to the data file for a given date.
func (fs *FileStorer) getDailyDataPath(date time.Time) string {
	fileName := fmt.Sprintf("%s.json", date.Format("2006-01-02"))
	return filepath.Join(fs.cfg.DataDirPath, fileName)
}

// Save saves repository data to a JSON file for a specific date.
func (fs *FileStorer) Save(date time.Time, repos []*domain.Repository) error {
	path := fs.getDailyDataPath(date)
	fs.logger.Debug("Saving data", "path", path)

	// Ensure the data directory exists.
	if err := os.MkdirAll(fs.cfg.DataDirPath, 0755); err != nil {
		fs.logger.Error("Failed to create data directory", "path", fs.cfg.DataDirPath, "error", err)
		return fmt.Errorf("could not create data directory '%s': %w", fs.cfg.DataDirPath, err)
	}

	file, err := os.Create(path)
	if err != nil {
		fs.logger.Error("Failed to create data file", "path", path, "error", err)
		return fmt.Errorf("could not create data file '%s': %w", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // for human-readability
	if err := encoder.Encode(repos); err != nil {
		fs.logger.Error("Failed to encode data to JSON", "path", path, "error", err)
		return fmt.Errorf("could not encode data to '%s': %w", path, err)
	}

	fs.logger.Info("Successfully saved data", "path", path)
	return nil
}

// Load loads repository data from a JSON file for a specific date.
func (fs *FileStorer) Load(date time.Time) ([]*domain.Repository, error) {
	path := fs.getDailyDataPath(date)
	fs.logger.Debug("Loading data", "path", path)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			fs.logger.Warn("Data file not found", "path", path)
			return nil, ErrDataNotFound
		}
		fs.logger.Error("Failed to open data file", "path", path, "error", err)
		return nil, fmt.Errorf("could not open data file '%s': %w", path, err)
	}
	defer file.Close()

	var repos []*domain.Repository
	if err := json.NewDecoder(file).Decode(&repos); err != nil {
		fs.logger.Error("Failed to decode JSON data", "path", path, "error", err)
		return nil, fmt.Errorf("could not decode JSON data from '%s': %w", path, err)
	}

	fs.logger.Debug("Successfully loaded data", "path", path)
	return repos, nil
}

// LoadTargetRepos loads the list of target repositories from repos.json.
func (fs *FileStorer) LoadTargetRepos() ([]string, error) {
	path := fs.cfg.ReposFilePath
	fs.logger.Debug("Loading target repos", "path", path)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			fs.logger.Warn("Repos config file not found", "path", path)
			return nil, ErrReposConfigNotFound
		}
		fs.logger.Error("Failed to open repos config file", "path", path, "error", err)
		return nil, fmt.Errorf("could not open repos config file '%s': %w", path, err)
	}
	defer file.Close()

	var repos []string
	if err := json.NewDecoder(file).Decode(&repos); err != nil {
		fs.logger.Error("Failed to decode repos config file", "path", path, "error", err)
		return nil, fmt.Errorf("could not decode repos config file '%s': %w", path, err)
	}

	fs.logger.Info("Successfully loaded target repos", "path", path, "count", len(repos))
	return repos, nil
}

// SaveTargetRepos saves the list of target repositories to repos.json.
func (fs *FileStorer) SaveTargetRepos(repos []string) error {
	path := fs.cfg.ReposFilePath
	fs.logger.Debug("Saving target repos", "path", path)

	file, err := os.Create(path)
	if err != nil {
		fs.logger.Error("Failed to create repos config file", "path", path, "error", err)
		return fmt.Errorf("could not create repos config file '%s': %w", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(repos); err != nil {
		fs.logger.Error("Failed to encode repos config to JSON", "path", path, "error", err)
		return fmt.Errorf("could not encode repos config to '%s': %w", path, err)
	}

	fs.logger.Info("Successfully saved target repos", "path", path, "count", len(repos))
	return nil
}
