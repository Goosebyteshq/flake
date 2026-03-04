package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Goosebyteshq/flake/internal/app"
)

func TestPublishCommandRequiresURL(t *testing.T) {
	d := t.TempDir()
	events := filepath.Join(d, "events.json")
	payload := `{"schema_version":1,"policy_version":1,"run":{"run_id":"r","timestamp":"t","framework":"go","meta":{}},"tests":[],"transitions":[]}`
	if err := os.WriteFile(events, []byte(payload), 0o600); err != nil {
		t.Fatalf("write events: %v", err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"publish", "--events", events}, &out, &errOut, app.Scanner{})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestPublishCommandSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := t.TempDir()
	events := filepath.Join(d, "events.json")
	payload := `{"schema_version":1,"policy_version":1,"run":{"run_id":"r","timestamp":"t","framework":"go","meta":{}},"tests":[],"transitions":[]}`
	if err := os.WriteFile(events, []byte(payload), 0o600); err != nil {
		t.Fatalf("write events: %v", err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"publish", "--events", events, "--url", srv.URL, "--retries", "0"}, &out, &errOut, app.Scanner{})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, errOut.String())
	}
}

func TestPublishCommandDebugRedactsQuerySecrets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	d := t.TempDir()
	events := filepath.Join(d, "events.json")
	payload := `{"schema_version":1,"policy_version":1,"run":{"run_id":"r","timestamp":"t","framework":"go","meta":{}},"tests":[],"transitions":[]}`
	if err := os.WriteFile(events, []byte(payload), 0o600); err != nil {
		t.Fatalf("write events: %v", err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"publish", "--events", events, "--url", srv.URL + "?token=secret", "--token", "super-secret", "--debug"}, &out, &errOut, app.Scanner{})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, errOut.String())
	}
	got := out.String()
	if strings.Contains(got, "super-secret") || strings.Contains(got, "token=secret") {
		t.Fatalf("debug leaked secret: %s", got)
	}
}
