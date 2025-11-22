package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourname/go-trendboard/internal/config"
	"github.com/yourname/go-trendboard/internal/infra/github"
	"github.com/yourname/go-trendboard/internal/infra/storage"
	"github.com/yourname/go-trendboard/internal/logger"
	"github.com/yourname/go-trendboard/internal/usecase"
)

var rootCmd = &cobra.Command{
	Use:   "go-trendboard",
	Short: "A CLI tool to track and generate trend dashboards for Go OSS.",
	Long: `go-trendboard is a tool that automatically collects star counts and trends
of specified Go open-source software from GitHub and generates a static dashboard
in Markdown or HTML format.`,
	SilenceUsage: true, // Prevents usage from being displayed on error
}

func init() {
	// init command
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a default repos.json file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			log := logger.NewLogger(cfg)
			storer := storage.NewFileStorer(cfg, log)
			uc := usecase.NewUsecase(cfg, log, nil, storer) // Fetcher is not needed for init

			return uc.Initialize(cmd.Context())
		},
	}

	// update command
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Fetch the latest star counts from GitHub",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			log := logger.NewLogger(cfg)
			fetcher := github.NewClient(cfg, log)
			storer := storage.NewFileStorer(cfg, log)
			uc := usecase.NewUsecase(cfg, log, fetcher, storer)

			return uc.Update(cmd.Context())
		},
	}

	// generate command
	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate the trend dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			log := logger.NewLogger(cfg)
			storer := storage.NewFileStorer(cfg, log)
			uc := usecase.NewUsecase(cfg, log, nil, storer) // Fetcher is not needed for generate

			return uc.Generate(cmd.Context())
		},
	}

	rootCmd.AddCommand(initCmd, updateCmd, generateCmd)
}

func main() {
	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// The error is already logged by cobra, so we just exit.
		os.Exit(1)
	}
}
