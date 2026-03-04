package report

import (
	"sort"

	"github.com/leomorpho/flake/internal/domain"
	"github.com/leomorpho/flake/internal/state"
)

type View string

const (
	ViewDefault          View = "default"
	ViewUnstable         View = "unstable"
	ViewRecovered        View = "recovered"
	ViewLongFailing      View = "long-failing"
	DefaultMinFailStreak      = 3
)

type Options struct {
	View          View
	Limit         int
	MinFailStreak int
}

type Row struct {
	TestID      string       `json:"test_id"`
	Class       domain.Class `json:"class"`
	FailureRate float64      `json:"failure_rate"`
	History     string       `json:"history"`
	LastSeen    string       `json:"last_seen,omitempty"`
	FailStreak  int          `json:"fail_streak,omitempty"`
}

type Summary struct {
	TotalByClass map[domain.Class]int `json:"total_by_class"`
	Rows         []Row                `json:"rows"`
	View         string               `json:"view,omitempty"`
}

func Build(st state.FileState) Summary {
	return BuildWithOptions(st, Options{View: ViewDefault})
}

func BuildWithOptions(st state.FileState, opts Options) Summary {
	rows := make([]Row, 0, len(st.Tests))
	counts := map[domain.Class]int{}
	for testID, slot := range st.Tests {
		counts[slot.LastState.Class]++
		rows = append(rows, Row{
			TestID:      testID,
			Class:       slot.LastState.Class,
			FailureRate: slot.LastState.FailureRate,
			History:     slot.History,
			LastSeen:    slot.LastSeen,
			FailStreak:  trailingFailStreak(slot.History),
		})
	}

	view := opts.View
	if view == "" {
		view = ViewDefault
	}
	switch view {
	case ViewUnstable:
		rows = filter(rows, func(r Row) bool {
			return r.Class == domain.ClassBroken || r.Class == domain.ClassFlaky || r.Class == domain.ClassNewFail
		})
		sort.SliceStable(rows, func(i, j int) bool {
			si := domain.ClassSeverity(rows[i].Class)
			sj := domain.ClassSeverity(rows[j].Class)
			if si != sj {
				return si > sj
			}
			if rows[i].FailureRate != rows[j].FailureRate {
				return rows[i].FailureRate > rows[j].FailureRate
			}
			return rows[i].TestID < rows[j].TestID
		})
	case ViewRecovered:
		rows = filter(rows, func(r Row) bool { return r.Class == domain.ClassRecovering })
		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].FailureRate != rows[j].FailureRate {
				return rows[i].FailureRate > rows[j].FailureRate
			}
			if rows[i].LastSeen != rows[j].LastSeen {
				return rows[i].LastSeen > rows[j].LastSeen
			}
			return rows[i].TestID < rows[j].TestID
		})
	case ViewLongFailing:
		minFailStreak := opts.MinFailStreak
		if minFailStreak <= 0 {
			minFailStreak = DefaultMinFailStreak
		}
		rows = filter(rows, func(r Row) bool { return r.FailStreak >= minFailStreak })
		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].FailStreak != rows[j].FailStreak {
				return rows[i].FailStreak > rows[j].FailStreak
			}
			if rows[i].FailureRate != rows[j].FailureRate {
				return rows[i].FailureRate > rows[j].FailureRate
			}
			return rows[i].TestID < rows[j].TestID
		})
	default:
		sort.SliceStable(rows, func(i, j int) bool {
			si := domain.ClassSeverity(rows[i].Class)
			sj := domain.ClassSeverity(rows[j].Class)
			if si != sj {
				return si > sj
			}
			if rows[i].FailureRate != rows[j].FailureRate {
				return rows[i].FailureRate > rows[j].FailureRate
			}
			return rows[i].TestID < rows[j].TestID
		})
	}

	if opts.Limit > 0 && len(rows) > opts.Limit {
		rows = rows[:opts.Limit]
	}
	return Summary{TotalByClass: counts, Rows: rows, View: string(view)}
}

func filter(rows []Row, keep func(Row) bool) []Row {
	out := make([]Row, 0, len(rows))
	for _, r := range rows {
		if keep(r) {
			out = append(out, r)
		}
	}
	return out
}

func trailingFailStreak(history string) int {
	n := 0
	for i := len(history) - 1; i >= 0; i-- {
		if history[i] != 'F' {
			break
		}
		n++
	}
	return n
}
