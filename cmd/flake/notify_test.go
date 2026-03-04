package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/leomorpho/flake/internal/app"
	"github.com/leomorpho/flake/internal/domain"
	"github.com/leomorpho/flake/internal/state"
)

func TestNotifyNeverFailsCI(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	cfgPath := filepath.Join(d, "cfg.yaml")
	cfg := "slack:\n  webhook: \"\"\nnotify:\n  on_transition: true\n"
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	st := state.New()
	st.LastRun.RunID = "r1"
	st.LastRun.Transitions = []domain.Transition{{TestID: "A", From: domain.ClassHealthy, To: domain.ClassNewFail, FailureRate: 1, Severity: 3}}
	st.Tests["A"] = state.TestSlot{LastState: state.SlotState{Class: domain.ClassNewFail, FailureRate: 1}}
	if err := (state.Store{Path: statePath}).Save(st); err != nil {
		t.Fatalf("save state: %v", err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"notify", "--state", statePath, "--config", cfgPath}, &out, &errOut, app.Scanner{})
	if code != 0 {
		t.Fatalf("notify returned non-zero: %d", code)
	}
}
