package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	// GitHubToken is the token for authenticating with the GitHub API.
	GitHubToken string `mapstructure:"github_token"`

	// LogLevel is the logging level (e.g., debug, info, warn, error).
	LogLevel string `mapstructure:"log_level"`

	// ReposFilePath is the path to the JSON file containing the list of repositories to track.
	ReposFilePath string `mapstructure:"repos_file_path"`

	// DataDirPath is the path to the directory where daily trend data is stored.
	DataDirPath string `mapstructure:"data_dir_path"`

	// DashboardFilePath is the path where the generated dashboard file will be saved.
	DashboardFilePath string `mapstructure:"dashboard_file_path"`

	// DashboardFormat is the format of the generated dashboard (md or html).
	DashboardFormat string `mapstructure:"dashboard_format"`

	// DashboardTemplatePath is the path to the HTML template file.
	DashboardTemplatePath string `mapstructure:"dashboard_template_path"`
}

// Load loads the configuration from environment variables and sets defaults.
func Load() (*Config, error) {
	v := viper.New()

	// Configure viper to read from environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set default values
	v.SetDefault("log_level", "info")
	v.SetDefault("repos_file_path", "repos.json")
	v.SetDefault("data_dir_path", "data")
	v.SetDefault("dashboard_file_path", "dashboard.md")
	v.SetDefault("dashboard_format", "md")
	v.SetDefault("dashboard_template_path", "dashboard.tpl")

	// Bind environment variables
	// Note: GITHUB_TOKEN is not bound here to prevent accidental exposure via other means.
	// It's read directly.

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Read the GitHub token directly from the environment variable.
	// This is a more secure practice for sensitive credentials.
	cfg.GitHubToken = v.GetString("github_token")

	// Validate required configuration
	if cfg.GitHubToken == "" {
		return nil, fmt.Errorf("required configuration not set: GITHUB_TOKEN")
	}

	return &cfg, nil
}
