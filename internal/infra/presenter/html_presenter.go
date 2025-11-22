package presenter

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/yourname/go-trendboard/internal/domain"
)

// HTMLPresenter renders trend data as an HTML page.
type HTMLPresenter struct {
	templatePath string
	logger       *slog.Logger
}

// NewHTMLPresenter creates a new HTMLPresenter.
func NewHTMLPresenter(templatePath string, logger *slog.Logger) (*HTMLPresenter, error) {
	// A simple check to see if the template file is accessible.
	// The actual parsing happens in Render.
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		logger.Error("HTML template not found", "path", templatePath)
		return nil, fmt.Errorf("html template not found at '%s'", templatePath)
	}
	return &HTMLPresenter{
		templatePath: templatePath,
		logger:       logger.With("component", "html_presenter"),
	}, nil
}

// Render generates an HTML report from the trend data.
func (p *HTMLPresenter) Render(writer io.Writer, trends []*domain.Trend) error {
	p.logger.Debug("Rendering trends to HTML", "template", p.templatePath)

	tmpl, err := template.ParseFiles(p.templatePath)
	if err != nil {
		p.logger.Error("Failed to parse HTML template", "error", err)
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	if len(trends) == 0 {
		p.logger.Info("No trends to render, rendering empty state")
		// Fallback for empty trends, though the template itself handles this
		data := map[string]interface{}{
			"GeneratedAt": time.Now().Format(time.RFC1123),
			"Trends":      nil,
		}
		return tmpl.Execute(writer, data)
	}

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
		Period      string
		TrendIcon   string
		GeneratedAt string
		Trends      []TemplateTrend
	}{
		Period:      string(period),
		TrendIcon:   trendIcon,
		GeneratedAt: time.Now().Format(time.RFC1123),
		Trends:      make([]TemplateTrend, len(trends)),
	}

	for i, t := range trends {
		templateData.Trends[i] = TemplateTrend{
			Rank:     i + 1,
			RepoName: t.Repository.FullName,
			Stars:    t.Repository.Stars,
			Diff:     t.Diff,
		}
	}

	if err := tmpl.Execute(writer, templateData); err != nil {
		p.logger.Error("Failed to execute HTML template", "error", err)
		return fmt.Errorf("failed to render HTML: %w", err)
	}

	p.logger.Info("Successfully rendered HTML report")
	return nil
}
