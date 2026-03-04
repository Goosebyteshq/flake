package engine

import (
	"reflect"
	"testing"

	"github.com/Goosebyteshq/flake/internal/domain"
)

func TestAppendStatus(t *testing.T) {
	cases := []struct {
		name    string
		history string
		status  domain.TestStatus
		window  int
		want    string
		wantErr bool
	}{
		{name: "append pass", history: "PF", status: domain.Pass, window: 10, want: "PFP"},
		{name: "append fail", history: "PP", status: domain.Fail, window: 10, want: "PPF"},
		{name: "skip no-op", history: "PF", status: domain.Skip, window: 10, want: "PF"},
		{name: "truncate", history: "PPFF", status: domain.Pass, window: 4, want: "PFFP"},
		{name: "invalid window", history: "P", status: domain.Pass, window: 0, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AppendStatus(tc.history, tc.status, tc.window)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("AppendStatus error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestApplyRunMissingTestsNoOp(t *testing.T) {
	start := map[string]string{
		"a": "PP",
		"b": "FF",
	}
	statuses := map[string]domain.TestStatus{
		"a": domain.Pass,
	}
	got, err := ApplyRun(start, statuses, 5)
	if err != nil {
		t.Fatalf("ApplyRun error: %v", err)
	}
	want := map[string]string{
		"a": "PPP",
		"b": "FF",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestApplyRunNewTest(t *testing.T) {
	got, err := ApplyRun(map[string]string{}, map[string]domain.TestStatus{"new": domain.Fail}, 3)
	if err != nil {
		t.Fatalf("ApplyRun error: %v", err)
	}
	if got["new"] != "F" {
		t.Fatalf("got %q, want F", got["new"])
	}
}
