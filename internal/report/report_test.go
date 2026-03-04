package report

import (
	"testing"

	"github.com/Goosebyteshq/flake/internal/domain"
	"github.com/Goosebyteshq/flake/internal/state"
)

func TestBuildSortOrder(t *testing.T) {
	st := state.New()
	st.Tests["c"] = state.TestSlot{History: "PPF", LastState: state.SlotState{Class: domain.ClassFlaky, FailureRate: 0.33}}
	st.Tests["a"] = state.TestSlot{History: "FFFFP", LastState: state.SlotState{Class: domain.ClassBroken, FailureRate: 0.8}}
	st.Tests["b"] = state.TestSlot{History: "FFFFFP", LastState: state.SlotState{Class: domain.ClassBroken, FailureRate: 0.9}}

	s := Build(st)
	if len(s.Rows) != 3 {
		t.Fatalf("len(rows)=%d, want 3", len(s.Rows))
	}
	if s.Rows[0].TestID != "b" || s.Rows[1].TestID != "a" || s.Rows[2].TestID != "c" {
		t.Fatalf("unexpected order: %+v", s.Rows)
	}
}

func TestBuildWithOptionsViews(t *testing.T) {
	st := state.New()
	st.Tests["broken"] = state.TestSlot{History: "PPFFF", LastState: state.SlotState{Class: domain.ClassBroken, FailureRate: 0.9}}
	st.Tests["flaky"] = state.TestSlot{History: "PFPFP", LastState: state.SlotState{Class: domain.ClassFlaky, FailureRate: 0.4}}
	st.Tests["recover"] = state.TestSlot{History: "PPF", LastState: state.SlotState{Class: domain.ClassRecovering, FailureRate: 0.33}}
	st.Tests["healthyFailTail"] = state.TestSlot{History: "PPFF", LastState: state.SlotState{Class: domain.ClassHealthy, FailureRate: 0.5}}

	unstable := BuildWithOptions(st, Options{View: ViewUnstable})
	if len(unstable.Rows) != 2 {
		t.Fatalf("unstable rows=%d, want 2", len(unstable.Rows))
	}
	recovered := BuildWithOptions(st, Options{View: ViewRecovered})
	if len(recovered.Rows) != 1 || recovered.Rows[0].TestID != "recover" {
		t.Fatalf("unexpected recovered rows: %+v", recovered.Rows)
	}
	longFailing := BuildWithOptions(st, Options{View: ViewLongFailing})
	if len(longFailing.Rows) != 1 || longFailing.Rows[0].TestID != "broken" {
		t.Fatalf("expected only broken to meet default long-failing threshold, got %+v", longFailing.Rows)
	}

	longFailingCustom := BuildWithOptions(st, Options{View: ViewLongFailing, MinFailStreak: 2})
	if len(longFailingCustom.Rows) != 2 {
		t.Fatalf("expected 2 rows with custom long-failing threshold, got %d", len(longFailingCustom.Rows))
	}

	limited := BuildWithOptions(st, Options{View: ViewDefault, Limit: 1})
	if len(limited.Rows) != 1 {
		t.Fatalf("limit not applied")
	}
}
