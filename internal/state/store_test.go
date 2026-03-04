package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/leomorpho/flake/internal/domain"
)

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

func TestStoreLoadMissingReturnsNew(t *testing.T) {
	d := t.TempDir()
	s := Store{Path: filepath.Join(d, "missing.json")}
	st, err := s.Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if st.SchemaVersion != CurrentSchemaVersion {
		t.Fatalf("schema_version=%d, want %d", st.SchemaVersion, CurrentSchemaVersion)
	}
	if st.Window != DefaultWindow {
		t.Fatalf("window=%d, want %d", st.Window, DefaultWindow)
	}
}

func TestStoreSaveLoadRoundTrip(t *testing.T) {
	d := t.TempDir()
	path := filepath.Join(d, "state.json")
	now := time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC)

	s := Store{Path: path, Clock: fixedClock{t: now}}
	in := New()
	in.Window = 75
	in.Tests["pkg/TestA"] = TestSlot{
		History: "PPFP",
		LastState: SlotState{
			Class:       domain.ClassFlaky,
			FailureRate: 0.25,
		},
	}
	in.LastRun = LastRun{
		RunID:     "run-1",
		Timestamp: now.Format(time.RFC3339),
		Framework: "go",
		RunMeta: map[string]string{
			"repo": "org/repo",
		},
		Transitions: []domain.Transition{
			{TestID: "pkg/TestA", From: domain.ClassHealthy, To: domain.ClassFlaky, FailureRate: 0.25, Severity: 4},
		},
	}

	if err := s.Save(in); err != nil {
		t.Fatalf("Save error: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if got.UpdatedAt != now.Format(time.RFC3339) {
		t.Fatalf("updated_at=%q, want %q", got.UpdatedAt, now.Format(time.RFC3339))
	}
	if got.Tests["pkg/TestA"].History != "PPFP" {
		t.Fatalf("history=%q, want PPFP", got.Tests["pkg/TestA"].History)
	}
	if got.PolicyVersion != DefaultPolicyVersion {
		t.Fatalf("policy_version=%d, want %d", got.PolicyVersion, DefaultPolicyVersion)
	}
}

func TestStoreLoadInvalidJSON(t *testing.T) {
	d := t.TempDir()
	path := filepath.Join(d, "state.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	_, err := (Store{Path: path}).Load()
	if err == nil || !strings.Contains(err.Error(), "invalid state json") {
		t.Fatalf("expected invalid state json error, got: %v", err)
	}
}

func TestStoreLoadUnsupportedSchema(t *testing.T) {
	d := t.TempDir()
	path := filepath.Join(d, "state.json")
	invalid := map[string]any{
		"schema_version": 999,
		"policy_version": 1,
		"window":         50,
		"tests":          map[string]any{},
		"last_run": map[string]any{
			"run_meta":    map[string]string{},
			"transitions": []any{},
		},
	}
	b, _ := json.Marshal(invalid)
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
	_, err := (Store{Path: path}).Load()
	if err == nil || !strings.Contains(err.Error(), "unsupported schema_version") {
		t.Fatalf("expected unsupported schema error, got: %v", err)
	}
}

func TestStoreSaveAtomicOverwrite(t *testing.T) {
	d := t.TempDir()
	path := filepath.Join(d, "state.json")
	s := Store{Path: path, Clock: fixedClock{t: time.Date(2026, 3, 3, 1, 2, 3, 0, time.UTC)}}

	first := New()
	first.Tests["t"] = TestSlot{History: "P", LastState: SlotState{Class: domain.ClassHealthy, FailureRate: 0.0}}
	if err := s.Save(first); err != nil {
		t.Fatalf("Save(first) error: %v", err)
	}

	second := New()
	second.Tests["t"] = TestSlot{History: "F", LastState: SlotState{Class: domain.ClassNewFail, FailureRate: 1.0}}
	if err := s.Save(second); err != nil {
		t.Fatalf("Save(second) error: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if strings.Contains(string(b), ".tmp") {
		t.Fatalf("state file should not contain temp marker")
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if got.Tests["t"].History != "F" {
		t.Fatalf("history=%q, want F", got.Tests["t"].History)
	}
}

func TestStableSortTransitions(t *testing.T) {
	items := []domain.Transition{
		{TestID: "c", Severity: 3, FailureRate: 0.3},
		{TestID: "a", Severity: 5, FailureRate: 0.4},
		{TestID: "b", Severity: 5, FailureRate: 0.9},
	}
	StableSortTransitions(items)
	if items[0].TestID != "b" || items[1].TestID != "a" || items[2].TestID != "c" {
		t.Fatalf("unexpected order: %+v", items)
	}
}
