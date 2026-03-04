package state

import (
	"sort"
	"time"
)

type CompactOptions struct {
	MaxTests          int
	DropUntouchedDays int
	Now               time.Time
}

type CompactResult struct {
	Before       int      `json:"before"`
	After        int      `json:"after"`
	Removed      int      `json:"removed"`
	RemovedByAge int      `json:"removed_by_age"`
	RemovedByCap int      `json:"removed_by_cap"`
	AgeIDs       []string `json:"age_ids,omitempty"`
	CapIDs       []string `json:"cap_ids,omitempty"`
}

func Compact(st *FileState, opts CompactOptions) CompactResult {
	st.Normalize()
	before := len(st.Tests)
	if before == 0 {
		return CompactResult{Before: 0, After: 0, Removed: 0}
	}

	now := opts.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	type row struct {
		id       string
		slot     TestSlot
		lastSeen time.Time
	}
	rows := make([]row, 0, len(st.Tests))
	for id, slot := range st.Tests {
		rows = append(rows, row{id: id, slot: slot, lastSeen: parseRFC3339(slot.LastSeen)})
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].id < rows[j].id })

	kept := make([]row, 0, len(rows))
	removedByAge := 0
	ageIDs := make([]string, 0)
	if opts.DropUntouchedDays > 0 {
		cutoff := now.Add(-time.Duration(opts.DropUntouchedDays) * 24 * time.Hour)
		for _, r := range rows {
			if r.lastSeen.IsZero() || r.lastSeen.Before(cutoff) {
				removedByAge++
				ageIDs = append(ageIDs, r.id)
				continue
			}
			kept = append(kept, r)
		}
	} else {
		kept = append(kept, rows...)
	}

	sort.SliceStable(kept, func(i, j int) bool {
		if !kept[i].lastSeen.Equal(kept[j].lastSeen) {
			return kept[i].lastSeen.After(kept[j].lastSeen)
		}
		return kept[i].id < kept[j].id
	})

	removedByCap := 0
	capIDs := make([]string, 0)
	if opts.MaxTests > 0 && len(kept) > opts.MaxTests {
		removedByCap = len(kept) - opts.MaxTests
		for _, r := range kept[opts.MaxTests:] {
			capIDs = append(capIDs, r.id)
		}
		kept = kept[:opts.MaxTests]
	}

	next := make(map[string]TestSlot, len(kept))
	for _, r := range kept {
		next[r.id] = r.slot
	}
	st.Tests = next
	after := len(st.Tests)
	return CompactResult{
		Before:       before,
		After:        after,
		Removed:      before - after,
		RemovedByAge: removedByAge,
		RemovedByCap: removedByCap,
		AgeIDs:       ageIDs,
		CapIDs:       capIDs,
	}
}

func parseRFC3339(v string) time.Time {
	if v == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return time.Time{}
	}
	return t
}
