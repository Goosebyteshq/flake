package app

import (
	"fmt"
	"io"
	"time"

	"github.com/leomorpho/flake/internal/state"
)

type CompactOptions struct {
	StatePath         string
	MaxTests          int
	DropUntouchedDays int
	DryRun            bool
	Debug             bool
}

type CompactResult struct {
	Before        int      `json:"before"`
	After         int      `json:"after"`
	Removed       int      `json:"removed"`
	RemovedByAge  int      `json:"removed_by_age"`
	RemovedByCap  int      `json:"removed_by_cap"`
	DebugMessages []string `json:"debug_messages,omitempty"`
}

type Compactor struct {
	Clock state.Clock
}

func (c Compactor) Run(opts CompactOptions, stdout io.Writer) (CompactResult, error) {
	store := state.Store{Path: opts.StatePath, Clock: c.Clock}
	st, err := store.Load()
	if err != nil {
		return CompactResult{}, err
	}
	res := state.Compact(&st, state.CompactOptions{
		MaxTests:          opts.MaxTests,
		DropUntouchedDays: opts.DropUntouchedDays,
		Now:               c.now(),
	})
	if !opts.DryRun {
		if err := store.Save(st); err != nil {
			return CompactResult{}, err
		}
	}
	mode := "applied"
	if opts.DryRun {
		mode = "dry-run"
	}
	debug := make([]string, 0)
	if opts.Debug {
		debug = append(debug, fmt.Sprintf("settings:max_tests=%d drop_days=%d dry_run=%t", opts.MaxTests, opts.DropUntouchedDays, opts.DryRun))
		debug = append(debug, fmt.Sprintf("removed_age_ids=%v", res.AgeIDs))
		debug = append(debug, fmt.Sprintf("removed_cap_ids=%v", res.CapIDs))
		fmt.Fprintf(stdout, "debug: %s | %s | %s\n", debug[0], debug[1], debug[2])
	}
	fmt.Fprintf(stdout, "compact: mode=%s before=%d after=%d removed=%d\n", mode, res.Before, res.After, res.Removed)
	return CompactResult{
		Before:        res.Before,
		After:         res.After,
		Removed:       res.Removed,
		RemovedByAge:  res.RemovedByAge,
		RemovedByCap:  res.RemovedByCap,
		DebugMessages: debug,
	}, nil
}

func (c Compactor) now() time.Time {
	if c.Clock == nil {
		return time.Now().UTC()
	}
	return c.Clock.Now().UTC()
}
