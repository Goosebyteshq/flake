package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Goosebyteshq/flake/internal/domain"
	"github.com/Goosebyteshq/flake/internal/state"
)

func TestBuildStableOrder(t *testing.T) {
	st := state.New()
	st.LastRun = state.LastRun{RunID: "r", Timestamp: "t", Framework: "go", RunMeta: map[string]string{"repo": "o/r"}}
	st.Tests["b"] = state.TestSlot{History: "FFP", LastState: state.SlotState{Class: domain.ClassBroken, FailureRate: 0.9}}
	st.Tests["a"] = state.TestSlot{History: "FP", LastState: state.SlotState{Class: domain.ClassFlaky, FailureRate: 0.5}}
	st.LastRun.Transitions = []domain.Transition{{TestID: "a", Severity: 4}, {TestID: "b", Severity: 5}}
	p := Build(st)
	if p.Tests[0].TestID != "b" || p.Transitions[0].TestID != "b" {
		t.Fatalf("unexpected order")
	}
}

func TestGoldenPayload(t *testing.T) {
	st := state.New()
	st.LastRun = state.LastRun{RunID: "run-1", Timestamp: "2026-03-03T00:00:00Z", Framework: "go", RunMeta: map[string]string{"repo": "org/repo"}}
	st.Tests["TestA"] = state.TestSlot{History: "PPF", LastSeen: "2026-03-03T00:00:00Z", LastFailedAt: "2026-03-03T00:00:00Z", LastState: state.SlotState{Class: domain.ClassFlaky, FailureRate: 0.33}}
	p := Build(st)
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b) + "\n"
	path := filepath.Join("testdata", "events_golden.json")
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Fatalf("events golden mismatch\n--- got ---\n%s--- want ---\n%s", got, string(want))
	}
}
