package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leomorpho/flake/internal/app"
)

func TestCLIE2EGolden(t *testing.T) {
	// This test validates stable golden fixtures for --events output.
	// CI providers inject run metadata env vars, which would make run.meta differ
	// between local and CI (and even between CI runs), so we clear provider env here.
	githubEnv := []string{
		"GITHUB_ACTIONS",
		"GITHUB_REPOSITORY",
		"GITHUB_RUN_ID",
		"GITHUB_REF_NAME",
		"GITHUB_SHA",
		"GITHUB_REF",
	}
	gitlabEnv := []string{
		"GITLAB_CI",
		"CI_PROJECT_PATH",
		"CI_COMMIT_REF_NAME",
		"CI_COMMIT_SHA",
		"CI_PIPELINE_ID",
		"CI_PIPELINE_URL",
		"CI_MERGE_REQUEST_IID",
	}
	for _, k := range append(githubEnv, gitlabEnv...) {
		t.Setenv(k, "")
	}

	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	eventsPath := filepath.Join(d, "events.json")
	cfgPath := filepath.Join(d, "config.yaml")
	failInput := filepath.Join(d, "fail.txt")
	passInput := filepath.Join(d, "pass.txt")

	if err := os.WriteFile(cfgPath, []byte("notify:\n  on_transition: true\n  min_failure_rate: 0.05\n  include_classes: [\"Recovering\"]\nslack:\n  webhook: \"\"\n"), 0o600); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	if err := os.WriteFile(failInput, []byte("--- FAIL: TestA (0.00s)\n"), 0o600); err != nil {
		t.Fatalf("write fail input: %v", err)
	}
	if err := os.WriteFile(passInput, []byte("--- PASS: TestA (0.00s)\n"), 0o600); err != nil {
		t.Fatalf("write pass input: %v", err)
	}

	hints := &memoryHintStore{}

	var out bytes.Buffer
	var errOut bytes.Buffer

	code := run([]string{"scan", "--state", statePath, "--framework", "go", "--input", failInput}, &out, &errOut, app.Scanner{Clock: fixedClock{t: time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC)}, HintStore: hints})
	if code != 0 {
		t.Fatalf("scan1 exit=%d stderr=%s", code, errOut.String())
	}
	assertGoldenPath(t, "testdata/e2e/scan1.golden", out.String())
	out.Reset()
	errOut.Reset()

	code = run([]string{"scan", "--state", statePath, "--framework", "go", "--input", passInput}, &out, &errOut, app.Scanner{Clock: fixedClock{t: time.Date(2026, 3, 3, 12, 1, 0, 0, time.UTC)}, HintStore: hints})
	if code != 0 {
		t.Fatalf("scan2 exit=%d stderr=%s", code, errOut.String())
	}
	assertGoldenPath(t, "testdata/e2e/scan2.golden", out.String())
	out.Reset()
	errOut.Reset()

	code = run([]string{"scan", "--state", statePath, "--framework", "go", "--input", passInput, "--events", eventsPath}, &out, &errOut, app.Scanner{Clock: fixedClock{t: time.Date(2026, 3, 3, 12, 2, 0, 0, time.UTC)}, HintStore: hints})
	if code != 0 {
		t.Fatalf("scan3 exit=%d stderr=%s", code, errOut.String())
	}
	assertGoldenPath(t, "testdata/e2e/scan3.golden", out.String())
	b, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	assertGoldenPath(t, "testdata/e2e/events.golden", string(b))
	out.Reset()
	errOut.Reset()

	code = run([]string{"notify", "--state", statePath, "--config", cfgPath}, &out, &errOut, app.Scanner{})
	if code != 0 {
		t.Fatalf("notify exit=%d", code)
	}
	assertGoldenPath(t, "testdata/e2e/notify.golden", out.String())
	out.Reset()
	errOut.Reset()

	code = run([]string{"report", "--state", statePath}, &out, &errOut, app.Scanner{})
	if code != 0 {
		t.Fatalf("report exit=%d stderr=%s", code, errOut.String())
	}
	assertGoldenPath(t, "testdata/e2e/report.golden", out.String())
}

func assertGoldenPath(t *testing.T, goldenPath, got string) {
	t.Helper()
	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v", goldenPath, err)
	}
	if got != string(wantBytes) {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s--- want ---\n%s", goldenPath, got, string(wantBytes))
	}
}
