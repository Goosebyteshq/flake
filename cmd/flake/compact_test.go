package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/leomorpho/flake/internal/app"
	"github.com/leomorpho/flake/internal/domain"
	"github.com/leomorpho/flake/internal/state"
)

func TestCompactCommandDryRun(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	st := state.New()
	now := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		st.Tests[fmt.Sprintf("T%d", i)] = state.TestSlot{LastSeen: now.Add(-time.Duration(i) * 24 * time.Hour).Format(time.RFC3339), LastState: state.SlotState{Class: domain.ClassHealthy, FailureRate: 0}}
	}
	if err := (state.Store{Path: statePath}).Save(st); err != nil {
		t.Fatalf("save: %v", err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"compact", "--state", statePath, "--max-tests", "1", "--dry-run"}, &out, &errOut, app.Scanner{})
	if code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, errOut.String())
	}
	loaded, err := (state.Store{Path: statePath}).Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Tests) != 3 {
		t.Fatalf("dry-run should preserve state")
	}
}

func TestCompactCommandApply(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	st := state.New()
	now := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		st.Tests[fmt.Sprintf("T%d", i)] = state.TestSlot{LastSeen: now.Add(-time.Duration(i) * 24 * time.Hour).Format(time.RFC3339), LastState: state.SlotState{Class: domain.ClassHealthy, FailureRate: 0}}
	}
	if err := (state.Store{Path: statePath}).Save(st); err != nil {
		t.Fatalf("save: %v", err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"compact", "--state", statePath, "--max-tests", "1"}, &out, &errOut, app.Scanner{})
	if code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, errOut.String())
	}
	loaded, err := (state.Store{Path: statePath}).Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Tests) != 1 {
		t.Fatalf("expected compacted state to keep one test")
	}
}

func TestCompactCommandDebugOutput(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	st := state.New()
	now := time.Now().UTC()
	st.Tests["A"] = state.TestSlot{LastSeen: now.Add(-20 * 24 * time.Hour).Format(time.RFC3339), LastState: state.SlotState{Class: domain.ClassHealthy, FailureRate: 0}}
	st.Tests["B"] = state.TestSlot{LastSeen: now.Format(time.RFC3339), LastState: state.SlotState{Class: domain.ClassHealthy, FailureRate: 0}}
	if err := (state.Store{Path: statePath, Clock: fixedClock{t: now}}).Save(st); err != nil {
		t.Fatalf("save: %v", err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"compact", "--state", statePath, "--drop-untouched-days", "2", "--debug"}, &out, &errOut, app.Scanner{})
	if code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "debug: settings:") {
		t.Fatalf("expected debug output, got: %s", out.String())
	}
}
