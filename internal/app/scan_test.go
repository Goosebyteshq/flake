package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leomorpho/flake/internal/domain"
	"github.com/leomorpho/flake/internal/parsers"
	"github.com/leomorpho/flake/internal/state"
)

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

func TestScanRunGoInput(t *testing.T) {
	d := t.TempDir()
	inputPath := filepath.Join(d, "input.txt")
	statePath := filepath.Join(d, "state.json")
	hintsPath := filepath.Join(d, "hints.json")
	input := `--- PASS: TestA (0.00s)
--- FAIL: TestB (0.00s)
--- SKIP: TestC (0.00s)
`
	if err := os.WriteFile(inputPath, []byte(input), 0o600); err != nil {
		t.Fatalf("WriteFile input error: %v", err)
	}

	now := time.Date(2026, 3, 3, 13, 0, 0, 0, time.UTC)
	buf := bytes.Buffer{}
	scanner := Scanner{Clock: fixedClock{t: now}}
	got, err := scanner.Run(ScanOptions{
		StatePath:       statePath,
		InputPath:       inputPath,
		Framework:       "go",
		Window:          5,
		ParserHintsPath: hintsPath,
		RepoKey:         d,
	}, &buf)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if got.Framework != "go" {
		t.Fatalf("framework=%q, want go", got.Framework)
	}
	if got.TestsParsed != 3 {
		t.Fatalf("tests=%d, want 3", got.TestsParsed)
	}
	if got.Policy.NewFailSampleMax != domain.DefaultClassificationPolicy().NewFailSampleMax {
		t.Fatalf("unexpected policy snapshot: %+v", got.Policy)
	}
	if len(got.Classifications) != 3 {
		t.Fatalf("classifications=%d, want 3", len(got.Classifications))
	}
	if got.Classifications[0].TestID != "TestA" || got.Classifications[1].TestID != "TestB" || got.Classifications[2].TestID != "TestC" {
		t.Fatalf("classifications not sorted by test id: %+v", got.Classifications)
	}
	if got.Classifications[1].Derived.Class != domain.ClassNewFail {
		t.Fatalf("expected TestB NewFail derived class, got %s", got.Classifications[1].Derived.Class)
	}
	if len(got.Classifications[1].Explanation.Reasons) == 0 {
		t.Fatalf("expected explanation reasons for TestB")
	}

	st, err := (state.Store{Path: statePath}).Load()
	if err != nil {
		t.Fatalf("Load state error: %v", err)
	}
	if st.Tests["TestA"].History != "P" {
		t.Fatalf("TestA history=%q, want P", st.Tests["TestA"].History)
	}
	if st.Tests["TestB"].History != "F" {
		t.Fatalf("TestB history=%q, want F", st.Tests["TestB"].History)
	}
	if st.Tests["TestC"].History != "" {
		t.Fatalf("TestC history=%q, want empty because skip is no-op", st.Tests["TestC"].History)
	}
	if st.LastRun.RunID == "" {
		t.Fatalf("expected run id")
	}
}

func TestDeterministicRunID(t *testing.T) {
	now := "2026-03-03T00:00:00Z"
	a := deterministicRunID(now, parsers.FrameworkGo, map[string]domain.TestStatus{
		"a": domain.Pass,
		"b": domain.Fail,
	})
	b := deterministicRunID(now, parsers.FrameworkGo, map[string]domain.TestStatus{
		"b": domain.Fail,
		"a": domain.Pass,
	})
	if a != b {
		t.Fatalf("run id mismatch: %q vs %q", a, b)
	}
}

func TestScanDebugIncludesCandidates(t *testing.T) {
	d := t.TempDir()
	inputPath := filepath.Join(d, "input.txt")
	statePath := filepath.Join(d, "state.json")
	hintsPath := filepath.Join(d, "hints.json")
	input := `--- PASS: TestA (0.00s)
`
	if err := os.WriteFile(inputPath, []byte(input), 0o600); err != nil {
		t.Fatalf("WriteFile input error: %v", err)
	}

	buf := bytes.Buffer{}
	scanner := Scanner{Clock: fixedClock{t: time.Date(2026, 3, 3, 13, 0, 0, 0, time.UTC)}}
	got, err := scanner.Run(ScanOptions{
		StatePath:       statePath,
		InputPath:       inputPath,
		Framework:       "auto",
		Window:          5,
		ParserHintsPath: hintsPath,
		RepoKey:         d,
		Debug:           true,
	}, &buf)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if got.Framework != "go" {
		t.Fatalf("framework=%q, want go", got.Framework)
	}
	if len(got.DebugMessages) == 0 {
		t.Fatalf("expected debug messages")
	}
	foundCandidates := false
	for _, m := range got.DebugMessages {
		if len(m) >= len("parser_candidates=") && m[:len("parser_candidates=")] == "parser_candidates=" {
			foundCandidates = true
			break
		}
	}
	if !foundCandidates {
		t.Fatalf("expected parser_candidates debug message, got %+v", got.DebugMessages)
	}
}

func TestScanPolicyOverrideFromConfig(t *testing.T) {
	d := t.TempDir()
	inputPath := filepath.Join(d, "input.txt")
	statePath := filepath.Join(d, "state.json")
	hintsPath := filepath.Join(d, "hints.json")
	configPath := filepath.Join(d, "config.yaml")
	input := `--- FAIL: TestA (0.00s)
--- FAIL: TestA (0.00s)
--- FAIL: TestA (0.00s)
--- PASS: TestA (0.00s)
--- PASS: TestA (0.00s)
`
	if err := os.WriteFile(inputPath, []byte(input), 0o600); err != nil {
		t.Fatalf("WriteFile input error: %v", err)
	}
	cfg := `policy:
  newfail_sample_max: 1
  broken_min_rate: 0.50
`
	if err := os.WriteFile(configPath, []byte(cfg), 0o600); err != nil {
		t.Fatalf("WriteFile config error: %v", err)
	}
	buf := bytes.Buffer{}
	scanner := Scanner{Clock: fixedClock{t: time.Date(2026, 3, 3, 13, 0, 0, 0, time.UTC)}}
	_, err := scanner.Run(ScanOptions{
		StatePath:       statePath,
		InputPath:       inputPath,
		Framework:       "go",
		Window:          50,
		ParserHintsPath: hintsPath,
		RepoKey:         d,
		ConfigPath:      configPath,
	}, &buf)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	st, err := (state.Store{Path: statePath}).Load()
	if err != nil {
		t.Fatalf("Load state error: %v", err)
	}
	if st.Tests["TestA"].LastState.Class != domain.ClassBroken {
		t.Fatalf("expected policy override to classify Broken, got %s", st.Tests["TestA"].LastState.Class)
	}
}
