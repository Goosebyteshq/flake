package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leomorpho/flake/internal/app"
	"github.com/leomorpho/flake/internal/domain"
	"github.com/leomorpho/flake/internal/state"
)

func TestCLIGoldenPublishAndCompactJSON(t *testing.T) {
	d := t.TempDir()
	eventsPath := filepath.Join(d, "events.json")
	statePath := filepath.Join(d, "state.json")

	payload := `{"schema_version":1,"policy_version":1,"run":{"run_id":"r","timestamp":"t","framework":"go","meta":{}},"tests":[],"transitions":[]}`
	if err := os.WriteFile(eventsPath, []byte(payload), 0o600); err != nil {
		t.Fatalf("write events: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"publish", "--events", eventsPath, "--url", srv.URL, "--retries", "0", "--json"}, &out, &errOut, app.Scanner{})
	if code != 0 {
		t.Fatalf("publish exit=%d stderr=%s", code, errOut.String())
	}
	assertGolden(t, "publish_json.golden", out.String())

	st := state.New()
	now := time.Now().UTC()
	st.Tests["A"] = state.TestSlot{LastSeen: now.Add(-20 * 24 * time.Hour).Format(time.RFC3339), LastState: state.SlotState{Class: domain.ClassHealthy, FailureRate: 0}}
	st.Tests["B"] = state.TestSlot{LastSeen: now.Add(-10 * 24 * time.Hour).Format(time.RFC3339), LastState: state.SlotState{Class: domain.ClassHealthy, FailureRate: 0}}
	st.Tests["C"] = state.TestSlot{LastSeen: now.Format(time.RFC3339), LastState: state.SlotState{Class: domain.ClassHealthy, FailureRate: 0}}
	if err := (state.Store{Path: statePath, Clock: fixedClock{t: now}}).Save(st); err != nil {
		t.Fatalf("save state: %v", err)
	}

	out.Reset()
	errOut.Reset()
	code = run([]string{"compact", "--state", statePath, "--drop-untouched-days", "2", "--max-tests", "1", "--json"}, &out, &errOut, app.Scanner{})
	if code != 0 {
		t.Fatalf("compact exit=%d stderr=%s", code, errOut.String())
	}
	assertGolden(t, "compact_json.golden", out.String())
}
