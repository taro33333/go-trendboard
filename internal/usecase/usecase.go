package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/yourname/go-trendboard/internal/config"
	"github.com/yourname/go-trendboard/internal/domain"
	"github.com/yourname/go-trendboard/internal/infra/github"
	"github.com/yourname/go-trendboard/internal/infra/presenter"
	"github.com/yourname/go-trendboard/internal/infra/storage"
)

// Usecase handles the main business logic of the application.
type Usecase struct {
	cfg       *config.Config
	logger    *slog.Logger
	fetcher   github.Fetcher
	storer    storage.Storer
}

// NewUsecase creates a new Usecase.
func NewUsecase(cfg *config.Config, logger *slog.Logger, fetcher github.Fetcher, storer storage.Storer) *Usecase {
	return &Usecase{
		cfg:       cfg,
		logger:    logger.With("component", "usecase"),
		fetcher:   fetcher,
		storer:    storer,
	}
}

// Initialize creates a default repos.json file if it doesn't exist.
func (u *Usecase) Initialize(ctx context.Context) error {
	u.logger.Info("Initializing...")

	// Check if the file already exists.
	if _, err := os.Stat(u.cfg.ReposFilePath); err == nil {
		u.logger.Info("Repositories config file already exists, skipping creation.", "path", u.cfg.ReposFilePath)
		return nil
	}

	defaultRepos := []string{
		"gin-gonic/gin",
		"go-chi/chi",
		"gorm/gorm",
		"spf13/cobra",
		"stretchr/testify",
		"uber-go/zap",
		"golang/mock",
		"google/go-github",
	}

	if err := u.storer.SaveTargetRepos(defaultRepos); err != nil {
		u.logger.Error("Failed to save default repositories config", "error", err)
		return fmt.Errorf("failed to initialize repositories config: %w", err)
	}

	u.logger.Info("Successfully created default repositories config file.", "path", u.cfg.ReposFilePath)
	return nil
}

// Update fetches the latest star counts for all target repositories and saves the data.
func (u *Usecase) Update(ctx context.Context) error {
	u.logger.Info("Updating repository data...")

	targetRepos, err := u.storer.LoadTargetRepos()
	if err != nil {
		u.logger.Error("Failed to load target repositories", "error", err)
		return fmt.Errorf("failed to load target repositories: %w", err)
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(8) // Limit concurrency to avoid hitting rate limits too quickly.

	results := make(chan *domain.Repository, len(targetRepos))

	for _, repoName := range targetRepos {
		repoName := repoName // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			repo, err := u.fetcher.FetchStars(gCtx, repoName)
			if err != nil {
				// Log the error but don't fail the entire update.
				// This allows the process to continue even if one repo is unavailable.
				u.logger.Warn("Failed to fetch stars for repository", "repo", repoName, "error", err)
				return nil
			}
			results <- repo
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		// This should ideally not happen since we are returning nil in goroutines.
		u.logger.Error("Error during concurrent fetching", "error", err)
		// We can choose to continue or fail here. Let's continue.
	}
	close(results)

	updatedRepos := make([]*domain.Repository, 0, len(targetRepos))
	for repo := range results {
		updatedRepos = append(updatedRepos, repo)
	}

	if len(updatedRepos) == 0 {
		u.logger.Warn("No repository data was successfully updated.")
		return nil
	}

	today := time.Now().UTC()
	if err := u.storer.Save(today, updatedRepos); err != nil {
		u.logger.Error("Failed to save updated repository data", "error", err)
		return fmt.Errorf("failed to save updated data: %w", err)
	}

	u.logger.Info("Successfully updated repository data.", "count", len(updatedRepos))
	return nil
}

// Generate creates a dashboard file based on historical data.
func (u *Usecase) Generate(ctx context.Context) error {
	u.logger.Info("Generating trend dashboard...")

	// For simplicity, we generate for the weekly period.
	// This could be extended to support daily, weekly, monthly via CLI flags.
	period := domain.TrendWeekly
	today := time.Now().UTC()
	pastDate := today.AddDate(0, 0, -7)

	todayData, err := u.storer.Load(today)
	if err != nil {
		u.logger.Error("Failed to load today's data. Please run 'update' first.", "date", today.Format("2006-01-02"), "error", err)
		return fmt.Errorf("failed to load today's data: %w", err)
	}

	pastData, err := u.storer.Load(pastDate)
	if err != nil {
		u.logger.Warn("Failed to load past data. Trend will be calculated against 0.", "date", pastDate.Format("2006-01-02"), "error", err)
		// We can continue without past data, the trend will be the full star count.
	}

	pastDataMap := make(map[string]int, len(pastData))
	for _, repo := range pastData {
		pastDataMap[repo.FullName] = repo.Stars
	}

	trends := make([]*domain.Trend, 0, len(todayData))
	for _, repo := range todayData {
		pastStars := pastDataMap[repo.FullName] // Defaults to 0 if not found
		diff := repo.Stars - pastStars
		trends = append(trends, domain.NewTrend(repo, diff, period))
	}

	domain.SortTrends(trends)

	// Open the output file
	file, err := os.Create(u.cfg.DashboardFilePath)
	if err != nil {
		u.logger.Error("Failed to create dashboard file", "path", u.cfg.DashboardFilePath, "error", err)
		return fmt.Errorf("failed to create dashboard file: %w", err)
	}
	defer file.Close()

	p, err := presenter.NewPresenter(u.cfg, u.logger)
	if err != nil {
		return err // Already logged in presenter factory
	}

	if err := p.Render(file, trends); err != nil {
		// Already logged in presenter
		return fmt.Errorf("failed to render dashboard: %w", err)
	}

	u.logger.Info("Successfully generated trend dashboard.", "path", u.cfg.DashboardFilePath)
	return nil
}
