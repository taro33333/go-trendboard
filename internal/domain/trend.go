package domain

import "sort"

// TrendPeriod represents the period for which a trend is calculated.
type TrendPeriod string

const (
	TrendDaily   TrendPeriod = "Daily"
	TrendWeekly  TrendPeriod = "Weekly"
	TrendMonthly TrendPeriod = "Monthly"
)

// Trend represents the star count trend for a repository over a specific period.
type Trend struct {
	// Repository is the repository for which the trend is calculated.
	Repository *Repository
	// Diff is the difference in star count over the period.
	Diff int
	// Period is the time period of the trend calculation.
	Period TrendPeriod
}

// NewTrend creates a new Trend object.
func NewTrend(repo *Repository, diff int, period TrendPeriod) *Trend {
	return &Trend{
		Repository: repo,
		Diff:       diff,
		Period:     period,
	}
}

// ByStarsDiff implements sort.Interface for []Trend based on the Diff field (descending).
type ByStarsDiff []*Trend

func (a ByStarsDiff) Len() int           { return len(a) }
func (a ByStarsDiff) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByStarsDiff) Less(i, j int) bool { return a[i].Diff > a[j].Diff }

// SortTrends sorts a slice of trends by star difference in descending order.
func SortTrends(trends []*Trend) {
	sort.Sort(ByStarsDiff(trends))
}
