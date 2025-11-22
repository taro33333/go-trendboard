package presenter

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourname/go-trendboard/internal/config"
	"github.com/yourname/go-trendboard/internal/domain"
)

func getTestTrends(t *testing.T) []*domain.Trend {
	t.Helper()
	repo1, _ := domain.NewRepository("owner/repo1", 1000)
	repo2, _ := domain.NewRepository("owner/repo2", 2500)
	return []*domain.Trend{
		domain.NewTrend(repo1, 50, domain.TrendWeekly),
		domain.NewTrend(repo2, 25, domain.TrendWeekly),
	}
}

func TestMarkdownPresenter_Render(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.NewFile(0, os.DevNull), nil))
	presenter := NewMarkdownPresenter(logger)
	trends := getTestTrends(t)

	var buf bytes.Buffer
	err := presenter.Render(&buf, trends)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "# Go OSS Trending (Weekly)")
	assert.Contains(t, output, "| 1 | [owner/repo1](https://github.com/owner/repo1) | 1000 | 50 ★ |")
	assert.Contains(t, output, "| 2 | [owner/repo2](https://github.com/owner/repo2) | 2500 | 25 ★ |")
}

func TestHTMLPresenter_Render(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.NewFile(0, os.DevNull), nil))
	trends := getTestTrends(t)

	// Create a temporary template file
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "test.tpl")
	templateContent := `
	<h1>{{ .Period }}</h1>
	{{ range .Trends }}
		<p>{{ .RepoName }}: {{ .Diff }}</p>
	{{ end }}
	`
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	presenter, err := NewHTMLPresenter(templatePath, logger)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = presenter.Render(&buf, trends)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<h1>Weekly</h1>")
	assert.Contains(t, output, "<p>owner/repo1: 50</p>")
	assert.Contains(t, output, "<p>owner/repo2: 25</p>")
}

func TestNewPresenter(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.NewFile(0, os.DevNull), nil))
	
	// Create a dummy template file for the HTML presenter case
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "dummy.tpl")
	err := os.WriteFile(templatePath, []byte("dummy"), 0644)
	require.NoError(t, err)


	testCases := []struct {
		name          string
		format        string
		expectedType  interface{}
		expectError   bool
	}{
		{
			name:         "Markdown format",
			format:       "md",
			expectedType: &MarkdownPresenter{},
			expectError:  false,
		},
		{
			name:         "HTML format",
			format:       "html",
			expectedType: &HTMLPresenter{},
			expectError:  false,
		},
		{
			name:         "Unknown format",
			format:       "xml",
			expectedType: nil,
			expectError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{DashboardFormat: tc.format}
			if tc.format == "html" {
				cfg.DashboardTemplatePath = templatePath
			}

			presenter, err := NewPresenter(cfg, logger)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, presenter)
			} else {
				require.NoError(t, err)
				assert.IsType(t, tc.expectedType, presenter)
			}
		})
	}
}
