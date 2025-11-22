package presenter

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/yourname/go-trendboard/internal/config"
	"github.com/yourname/go-trendboard/internal/domain"
)

// Presenter defines the interface for rendering trend data into a dashboard.
type Presenter interface {
	Render(writer io.Writer, trends []*domain.Trend) error
}

// NewPresenter is a factory function that returns the appropriate presenter
// based on the configuration.
func NewPresenter(cfg *config.Config, logger *slog.Logger) (Presenter, error) {
	switch cfg.DashboardFormat {
	case "md", "markdown":
		return NewMarkdownPresenter(logger), nil
	case "html":
		return NewHTMLPresenter(cfg.DashboardTemplatePath, logger)
	default:
		return nil, fmt.Errorf("unknown dashboard format: %s", cfg.DashboardFormat)
	}
}
