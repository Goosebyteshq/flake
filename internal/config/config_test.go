package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Window != 50 || cfg.PolicyVersion != 1 {
		t.Fatalf("unexpected defaults")
	}
	if cfg.Notify.MaxItemsPerMessage != 50 {
		t.Fatalf("unexpected notify defaults: %+v", cfg.Notify)
	}
}

func TestLoadOverrideNestedYAML(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "c.yaml")
	content := `
window: 100
policy_version: 2
policy:
  broken_min_rate: 0.75
  recovering:
    enabled: false
    min_drop: 0.30
notify:
  max_items_per_message: 10
  min_transition_age_runs: 2
  oscillation_window_runs: 4
  include_classes:
    - Broken
    - Flaky
  suppress_repeats_for_runs: 7
run_meta:
  prefer_env: false
  allow_file_override: true
`
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Window != 100 || cfg.PolicyVersion != 2 {
		t.Fatalf("unexpected top-level cfg: %+v", cfg)
	}
	if cfg.Policy.BrokenMinRate != 0.75 || cfg.Policy.Recovering.Enabled != false || cfg.Policy.Recovering.MinDrop != 0.30 {
		t.Fatalf("unexpected policy cfg: %+v", cfg.Policy)
	}
	if cfg.Notify.MaxItemsPerMessage != 10 || len(cfg.Notify.IncludeClasses) != 2 || cfg.Notify.SuppressRepeatsForRuns != 7 {
		t.Fatalf("unexpected notify cfg: %+v", cfg.Notify)
	}
	if cfg.Notify.MinTransitionAgeRuns != 2 || cfg.Notify.OscillationWindowRuns != 4 {
		t.Fatalf("unexpected notify age/oscillation cfg: %+v", cfg.Notify)
	}
	if cfg.RunMeta.PreferEnv != false || cfg.RunMeta.AllowFileOverride != true {
		t.Fatalf("unexpected runmeta cfg: %+v", cfg.RunMeta)
	}
}

func TestLoadAutoDefaultFile(t *testing.T) {
	d := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(orig) }()
	if err := os.Chdir(d); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	content := `window: 77
policy_version: 3
`
	if err := os.WriteFile(".flake-config.yaml", []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("load auto: %v", err)
	}
	if cfg.Window != 77 || cfg.PolicyVersion != 3 {
		t.Fatalf("unexpected auto-loaded cfg: %+v", cfg)
	}
}

func TestLoadMissingPathErrors(t *testing.T) {
	if _, err := Load("/definitely/missing/file.yaml"); err == nil {
		t.Fatalf("expected missing file error")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "bad.yaml")
	if err := os.WriteFile(p, []byte("window: ["), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := Load(p); err == nil {
		t.Fatalf("expected YAML error")
	}
}

func TestLoadValidation(t *testing.T) {
	d := t.TempDir()
	cases := []struct {
		name string
		yml  string
	}{
		{name: "window must be positive", yml: "window: 0\n"},
		{name: "policy version must be positive", yml: "policy_version: 0\n"},
		{name: "invalid policy bounds", yml: "policy:\n  flaky_min_rate: 0.9\n  flaky_max_rate: 0.1\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := filepath.Join(d, tc.name+".yaml")
			if err := os.WriteFile(p, []byte(tc.yml), 0o600); err != nil {
				t.Fatalf("write: %v", err)
			}
			if _, err := Load(p); err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}
}

func TestLoadAppliesFallbacks(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "fallbacks.yaml")
	content := `
notify:
  max_items_per_message: 0
slack:
  timeout_seconds: 0
`
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Notify.MaxItemsPerMessage != 50 || cfg.Slack.TimeoutSeconds != 5 {
		t.Fatalf("fallbacks not applied: %+v", cfg)
	}
}
