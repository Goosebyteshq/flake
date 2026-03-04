package app

import (
	"fmt"
	"io"

	"github.com/Goosebyteshq/flake/internal/report"
	"github.com/Goosebyteshq/flake/internal/state"
)

type ReportOptions struct {
	StatePath     string
	View          string
	Limit         int
	MinFailStreak int
}

type Reporter struct{}

func (Reporter) Run(opts ReportOptions, stdout io.Writer) (report.Summary, error) {
	store := state.Store{Path: opts.StatePath}
	st, err := store.Load()
	if err != nil {
		return report.Summary{}, err
	}
	summary := report.BuildWithOptions(st, report.Options{
		View:          report.View(opts.View),
		Limit:         opts.Limit,
		MinFailStreak: opts.MinFailStreak,
	})
	for _, row := range summary.Rows {
		fmt.Fprintf(stdout, "%s %0.2f %s\n", row.Class, row.FailureRate, row.TestID)
	}
	return summary, nil
}
