package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Goosebyteshq/flake/internal/app"
	"github.com/Goosebyteshq/flake/internal/parsers"
)

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

type memoryHintStore struct {
	v parsers.Framework
}

func (m *memoryHintStore) Get(_ string) (parsers.Framework, bool, error) {
	if m.v == "" {
		return "", false, nil
	}
	return m.v, true, nil
}

func (m *memoryHintStore) Put(_ string, framework parsers.Framework) error {
	m.v = framework
	return nil
}

func TestCLIGoldenScanSummaryAndReport(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	inputPath := filepath.Join("..", "..", "testdata", "go", "sample.txt")

	scanner := app.Scanner{
		Clock:     fixedClock{t: time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC)},
		HintStore: &memoryHintStore{},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"scan", "--state", statePath, "--framework", "go", "--input", inputPath}, &stdout, &stderr, scanner)
	if code != 0 {
		t.Fatalf("scan exit=%d stderr=%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
	assertGolden(t, "scan_summary.golden", stdout.String())

	stdout.Reset()
	stderr.Reset()
	code = run([]string{"report", "--state", statePath}, &stdout, &stderr, app.Scanner{})
	if code != 0 {
		t.Fatalf("report exit=%d stderr=%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
	assertGolden(t, "report_output.golden", stdout.String())
}

func TestCLIGoldenJSON(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	inputPath := filepath.Join("..", "..", "testdata", "go", "sample.txt")
	scanner := app.Scanner{
		Clock:     fixedClock{t: time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC)},
		HintStore: &memoryHintStore{},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"scan", "--state", statePath, "--framework", "go", "--input", inputPath, "--json"}, &stdout, &stderr, scanner)
	if code != 0 {
		t.Fatalf("scan exit=%d stderr=%s", code, stderr.String())
	}
	assertGolden(t, "scan_json.golden", stdout.String())

	stdout.Reset()
	stderr.Reset()
	code = run([]string{"report", "--state", statePath, "--json"}, &stdout, &stderr, app.Scanner{})
	if code != 0 {
		t.Fatalf("report exit=%d stderr=%s", code, stderr.String())
	}
	assertGolden(t, "report_json.golden", stdout.String())
}

func TestCLIStdinFlow(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	inputPath := filepath.Join("..", "..", "testdata", "go", "sample.txt")
	f, err := os.Open(inputPath)
	if err != nil {
		t.Fatalf("open input: %v", err)
	}
	defer f.Close()
	origStdin := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = origStdin }()

	scanner := app.Scanner{
		Clock:     fixedClock{t: time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC)},
		HintStore: &memoryHintStore{},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"scan", "--state", statePath, "--framework", "go"}, &stdout, &stderr, scanner)
	if code != 0 {
		t.Fatalf("scan via stdin exit=%d stderr=%s", code, stderr.String())
	}
	if stdout.Len() == 0 {
		t.Fatalf("expected stdout summary")
	}
}

func TestFailOn(t *testing.T) {
	counts := map[string]int{"NewFail": 1, "Flaky": 0, "Broken": 0}
	if !shouldFailOn("newfail", counts) {
		t.Fatalf("expected fail-on newfail to trigger")
	}
	if shouldFailOn("broken", counts) {
		t.Fatalf("did not expect fail-on broken to trigger")
	}
	if !shouldFailOn("broken,newfail", counts) {
		t.Fatalf("expected multi fail-on to trigger")
	}
}

func TestScanFailOnExitCode(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	inputPath := filepath.Join("..", "..", "testdata", "go", "sample.txt")
	scanner := app.Scanner{
		Clock:     fixedClock{t: time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC)},
		HintStore: &memoryHintStore{},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"scan", "--state", statePath, "--framework", "go", "--input", inputPath, "--fail-on", "newfail"}, &stdout, &stderr, scanner)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), `fail-on triggered: spec="newfail"`) {
		t.Fatalf("expected fail-on diagnostics, stderr=%q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "NewFail 1.00 TestBeta reasons=newfail_small_sample") {
		t.Fatalf("expected detailed fail-on row, stderr=%q", stderr.String())
	}
}

func TestMigrateCommand(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	if err := os.WriteFile(statePath, []byte(`{"policy_version":1,"window":50,"tests":{},"last_run":{"run_meta":{},"transitions":[]}}`), 0o600); err != nil {
		t.Fatalf("write legacy state: %v", err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"migrate", "--state", statePath}, &out, &errOut, app.Scanner{})
	if code != 0 {
		t.Fatalf("migrate exit=%d stderr=%s", code, errOut.String())
	}
	if out.String() == "" {
		t.Fatalf("expected migrate output")
	}
}

func assertGolden(t *testing.T, fileName, got string) {
	t.Helper()
	path := filepath.Join("testdata", "golden", fileName)
	wantBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", path, err)
	}
	if got != string(wantBytes) {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s--- want ---\n%s", fileName, got, string(wantBytes))
	}
}
