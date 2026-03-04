package state

import (
	"fmt"
	"testing"
	"time"

	"github.com/Goosebyteshq/flake/internal/domain"
)

func TestCompactByAgeAndCap(t *testing.T) {
	st := New()
	now := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		seen := now.Add(-time.Duration(i) * 24 * time.Hour).Format(time.RFC3339)
		st.Tests[fmt.Sprintf("T%d", i)] = TestSlot{LastSeen: seen, LastState: SlotState{Class: domain.ClassHealthy, FailureRate: 0}}
	}
	res := Compact(&st, CompactOptions{DropUntouchedDays: 2, MaxTests: 2, Now: now})
	if res.Before != 5 || res.After != 2 {
		t.Fatalf("unexpected compact result: %+v", res)
	}
	if _, ok := st.Tests["T0"]; !ok {
		t.Fatalf("expected newest test kept")
	}
	if _, ok := st.Tests["T1"]; !ok {
		t.Fatalf("expected second newest test kept")
	}
}

func TestCompactNoop(t *testing.T) {
	st := New()
	st.Tests["A"] = TestSlot{LastSeen: "2026-03-03T00:00:00Z", LastState: SlotState{Class: domain.ClassHealthy, FailureRate: 0}}
	res := Compact(&st, CompactOptions{})
	if res.Removed != 0 || len(st.Tests) != 1 {
		t.Fatalf("expected noop compact")
	}
}
