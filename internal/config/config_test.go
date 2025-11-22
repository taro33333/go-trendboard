package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Success(t *testing.T) {

	// Set environment variables for the test
	t.Setenv("GITHUB_TOKEN", "test_token_123")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("REPOS_FILE_PATH", "my_repos.json")
	t.Setenv("DATA_DIR_PATH", "my_data")
	t.Setenv("DASHBOARD_FILE_PATH", "my_dashboard.html")
	t.Setenv("DASHBOARD_FORMAT", "html")
	t.Setenv("DASHBOARD_TEMPLATE_PATH", "my_template.tpl")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "test_token_123", cfg.GitHubToken)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "my_repos.json", cfg.ReposFilePath)
	assert.Equal(t, "my_data", cfg.DataDirPath)
	assert.Equal(t, "my_dashboard.html", cfg.DashboardFilePath)
	assert.Equal(t, "html", cfg.DashboardFormat)
	assert.Equal(t, "my_template.tpl", cfg.DashboardTemplatePath)
}

func TestLoad_DefaultValues(t *testing.T) {

	// Ensure environment variables are unset
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("REPOS_FILE_PATH")
	os.Unsetenv("DATA_DIR_PATH")
	os.Unsetenv("DASHBOARD_FILE_PATH")
	os.Unsetenv("DASHBOARD_FORMAT")
	os.Unsetenv("DASHBOARD_TEMPLATE_PATH")
	
	// Set only the required environment variable
	t.Setenv("GITHUB_TOKEN", "test_token_456")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "test_token_456", cfg.GitHubToken)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "repos.json", cfg.ReposFilePath)
	assert.Equal(t, "data", cfg.DataDirPath)
	assert.Equal(t, "dashboard.md", cfg.DashboardFilePath)
	assert.Equal(t, "md", cfg.DashboardFormat)
	assert.Equal(t, "dashboard.tpl", cfg.DashboardTemplatePath)
}

func TestLoad_MissingGitHubToken_Error(t *testing.T) {
	
	// Ensure the required GITHUB_TOKEN is not set
	os.Unsetenv("GITHUB_TOKEN")

	cfg, err := Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "required configuration not set: GITHUB_TOKEN")
}
