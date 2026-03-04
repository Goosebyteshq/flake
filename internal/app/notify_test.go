package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/leomorpho/flake/internal/domain"
	"github.com/leomorpho/flake/internal/state"
)

type fakeNotifier struct {
	msgs []string
	err  error
}

func (f *fakeNotifier) Send(_ context.Context, message string) error {
	f.msgs = append(f.msgs, message)
	return f.err
}

func TestNotifyRun(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	cfgPath := filepath.Join(d, "cfg.yaml")
	cfg := `notify:
  on_transition: true
  min_failure_rate: 0.05
  include_classes: ["Broken", "Flaky", "NewFail", "Recovering"]
  suppress_repeats_for_runs: 5
  min_transition_age_runs: 0
  oscillation_window_runs: 3
  batch: true
  max_items_per_message: 50
slack:
  timeout_seconds: 1
  webhook: ""
`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	st := state.New()
	st.LastRun.RunID = "r1"
	st.LastRunIndex = 1
	st.Tests["T"] = state.TestSlot{LastState: state.SlotState{Class: domain.ClassNewFail, FailureRate: 1}}
	st.LastRun.Transitions = []domain.Transition{{TestID: "T", From: domain.ClassHealthy, To: domain.ClassNewFail, FailureRate: 1, Severity: 3}}
	if err := (state.Store{Path: statePath}).Save(st); err != nil {
		t.Fatalf("save state: %v", err)
	}

	buf := bytes.Buffer{}
	n := &fakeNotifier{}
	res, err := (NotifierRunner{Notifier: n}).Run(NotifyOptions{StatePath: statePath, ConfigPath: cfgPath}, &buf)
	if err != nil {
		t.Fatalf("run notify: %v", err)
	}
	if res.Sent != 1 || len(n.msgs) != 1 {
		t.Fatalf("unexpected notify result: %+v msgs=%d", res, len(n.msgs))
	}
}
