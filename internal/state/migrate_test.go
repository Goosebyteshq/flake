package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leomorpho/flake/internal/domain"
)

func TestMigrateLegacyV0(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "state.json")
	legacy := map[string]any{
		"policy_version": 1,
		"window":         50,
		"tests": map[string]any{
			"A": map[string]any{
				"history": "PF",
				"last_state": map[string]any{
					"class":        "Flaky",
					"failure_rate": 0.5,
				},
			},
		},
		"last_run": map[string]any{"run_meta": map[string]any{}, "transitions": []any{}},
	}
	b, _ := json.Marshal(legacy)
	if err := os.WriteFile(p, b, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	from, to, err := MigrateInPlace(p, fixedClock{t: time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if from != 0 || to != 1 {
		t.Fatalf("unexpected versions %d -> %d", from, to)
	}
	st, err := (Store{Path: p}).Load()
	if err != nil {
		t.Fatalf("load migrated: %v", err)
	}
	if st.SchemaVersion != 1 || st.Tests["A"].LastState.Class != domain.ClassFlaky {
		t.Fatalf("unexpected migrated state: %+v", st)
	}
}

func TestMigrateUnsupportedSchema(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "state.json")
	if err := os.WriteFile(p, []byte(`{"schema_version":99}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, _, err := MigrateInPlace(p, nil); err == nil {
		t.Fatalf("expected unsupported schema error")
	}
}
