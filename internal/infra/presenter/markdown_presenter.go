package presenter

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"text/template"

	"github.com/yourname/go-trendboard/internal/domain"
)

const markdownTemplate = `
# Go OSS Trending ({{ .Period }})

| Rank | Repository | Stars | Trend ({{ .TrendIcon }}) |
|:----:|:-----------|:------|:-----------|
{{- range .Trends }}
| {{ .Rank }} | [{{ .RepoName }}](https://github.com/{{ .RepoName }}) | {{ .Stars }} | {{ .Diff }} ★ |
{{- end }}
`

// MarkdownPresenter renders trend data as a Markdown table.
type MarkdownPresenter struct {
	logger *slog.Logger
}

// NewMarkdownPresenter creates a new MarkdownPresenter.
func NewMarkdownPresenter(logger *slog.Logger) *MarkdownPresenter {
	return &MarkdownPresenter{
		logger: logger.With("component", "markdown_presenter"),
	}
}

// Render generates a Markdown report from the trend data.
func (p *MarkdownPresenter) Render(writer io.Writer, trends []*domain.Trend) error {
	p.logger.Debug("Rendering trends to Markdown")

	if len(trends) == 0 {
		p.logger.Info("No trends to render, writing empty message")
		_, err := writer.Write([]byte("# Go OSS Trending\n\nNo trending data available.\n"))
		return err
	}

	// Prepare data for the template
	period := trends[0].Period
	trendIcon := "7d"
	switch period {
	case domain.TrendDaily:
		trendIcon = "24h"
	case domain.TrendWeekly:
		trendIcon = "7d"
	case domain.TrendMonthly:
		trendIcon = "30d"
	}

	type TemplateTrend struct {
		Rank     int
		RepoName string
		Stars    int
		Diff     int
	}
	
	templateData := struct {
		Period    string
		TrendIcon string
		Trends    []TemplateTrend
	}{
		Period:    string(period),
		TrendIcon: trendIcon,
		Trends:    make([]TemplateTrend, len(trends)),
	}

	for i, t := range trends {
		templateData.Trends[i] = TemplateTrend{
			Rank:     i + 1,
			RepoName: t.Repository.FullName,
			Stars:    t.Repository.Stars,
			Diff:     t.Diff,
		}
	}

	// Use text/template for simple replacements
	tmpl, err := template.New("markdown").Parse(strings.TrimSpace(markdownTemplate))
	if err != nil {
		p.logger.Error("Failed to parse markdown template", "error", err)
		return fmt.Errorf("failed to parse markdown template: %w", err)
	}

	if err := tmpl.Execute(writer, templateData); err != nil {
		p.logger.Error("Failed to execute markdown template", "error", err)
		return fmt.Errorf("failed to render markdown: %w", err)
	}

	p.logger.Info("Successfully rendered Markdown report")
	return nil
}
