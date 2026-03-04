package app

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Goosebyteshq/flake/internal/domain"
	"github.com/Goosebyteshq/flake/internal/state"
)

type compactClock struct{ t time.Time }

func (c compactClock) Now() time.Time { return c.t }

func TestCompactorDryRun(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	st := state.New()
	now := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 4; i++ {
		st.Tests[fmt.Sprintf("T%d", i)] = state.TestSlot{LastSeen: now.Add(-time.Duration(i) * 24 * time.Hour).Format(time.RFC3339), LastState: state.SlotState{Class: domain.ClassHealthy, FailureRate: 0}}
	}
	if err := (state.Store{Path: statePath}).Save(st); err != nil {
		t.Fatalf("save: %v", err)
	}
	buf := bytes.Buffer{}
	res, err := (Compactor{Clock: compactClock{t: now}}).Run(CompactOptions{StatePath: statePath, MaxTests: 2, DryRun: true, Debug: true}, &buf)
	if err != nil {
		t.Fatalf("compact: %v", err)
	}
	if res.After != 2 || res.Removed != 2 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if len(res.DebugMessages) == 0 {
		t.Fatalf("expected debug messages")
	}
	loaded, err := (state.Store{Path: statePath}).Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Tests) != 4 {
		t.Fatalf("dry-run should not persist changes")
	}
}
