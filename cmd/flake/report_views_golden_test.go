package main

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/Goosebyteshq/flake/internal/app"
	"github.com/Goosebyteshq/flake/internal/domain"
	"github.com/Goosebyteshq/flake/internal/state"
)

func TestCLIReportViewsGolden(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	st := state.New()
	st.Tests["broken"] = state.TestSlot{History: "PPFFF", LastState: state.SlotState{Class: domain.ClassBroken, FailureRate: 0.9}, LastSeen: "2026-03-03T12:00:00Z"}
	st.Tests["flaky"] = state.TestSlot{History: "PFPFP", LastState: state.SlotState{Class: domain.ClassFlaky, FailureRate: 0.4}, LastSeen: "2026-03-03T12:00:00Z"}
	st.Tests["recover"] = state.TestSlot{History: "PPF", LastState: state.SlotState{Class: domain.ClassRecovering, FailureRate: 0.33}, LastSeen: "2026-03-03T12:00:00Z"}
	st.Tests["healthyFailTail"] = state.TestSlot{History: "PPFF", LastState: state.SlotState{Class: domain.ClassHealthy, FailureRate: 0.5}, LastSeen: "2026-03-03T12:00:00Z"}
	if err := (state.Store{Path: statePath}).Save(st); err != nil {
		t.Fatalf("save state: %v", err)
	}

	cases := []struct {
		name   string
		args   []string
		golden string
	}{
		{name: "unstable", args: []string{"report", "--state", statePath, "--view", "unstable"}, golden: "report_unstable.golden"},
		{name: "recovered", args: []string{"report", "--state", statePath, "--view", "recovered"}, golden: "report_recovered.golden"},
		{name: "long-failing", args: []string{"report", "--state", statePath, "--view", "long-failing"}, golden: "report_long_failing.golden"},
		{name: "long-failing-json", args: []string{"report", "--state", statePath, "--view", "long-failing", "--limit", "2", "--json"}, golden: "report_long_failing_json.golden"},
		{name: "long-failing-custom-threshold", args: []string{"report", "--state", statePath, "--view", "long-failing", "--min-fail-streak", "1"}, golden: "report_long_failing_min1.golden"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			var errOut bytes.Buffer
			code := run(tc.args, &out, &errOut, app.Scanner{})
			if code != 0 {
				t.Fatalf("exit=%d stderr=%s", code, errOut.String())
			}
			assertGolden(t, tc.golden, out.String())
		})
	}
}
