package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakePublisher struct {
	called bool
	last   []byte
	err    error
}

func (f *fakePublisher) Publish(_ context.Context, payload []byte) error {
	f.called = true
	f.last = append([]byte(nil), payload...)
	return f.err
}

func TestPublishRun(t *testing.T) {
	d := t.TempDir()
	eventsPath := filepath.Join(d, "events.json")
	payload := `{"schema_version":1,"policy_version":1,"run":{"run_id":"r","timestamp":"t","framework":"go","meta":{}},"tests":[],"transitions":[]}`
	if err := os.WriteFile(eventsPath, []byte(payload), 0o600); err != nil {
		t.Fatalf("write events: %v", err)
	}
	buf := bytes.Buffer{}
	fp := &fakePublisher{}
	res, err := (PublisherRunner{Publisher: fp}).Run(PublishOptions{EventsPath: eventsPath}, &buf)
	if err != nil {
		t.Fatalf("run publish: %v", err)
	}
	if !res.Published || !fp.called {
		t.Fatalf("expected published true")
	}
}

func TestPublishRequiresEvents(t *testing.T) {
	_, err := (PublisherRunner{}).Run(PublishOptions{}, &bytes.Buffer{})
	if err == nil {
		t.Fatalf("expected --events error")
	}
}

func TestPublishDebugMessagesNoSecrets(t *testing.T) {
	d := t.TempDir()
	eventsPath := filepath.Join(d, "events.json")
	payload := `{"schema_version":1,"policy_version":1,"run":{"run_id":"run-7","timestamp":"t","framework":"go","meta":{}},"tests":[],"transitions":[]}`
	if err := os.WriteFile(eventsPath, []byte(payload), 0o600); err != nil {
		t.Fatalf("write events: %v", err)
	}
	buf := bytes.Buffer{}
	res, err := (PublisherRunner{Publisher: &fakePublisher{}}).Run(PublishOptions{
		EventsPath: eventsPath,
		URL:        "https://example.com/api?token=secret",
		Token:      "super-secret",
		MaxRetries: 0,
		Debug:      true,
	}, &buf)
	if err != nil {
		t.Fatalf("publish run error: %v", err)
	}
	if len(res.DebugMessages) == 0 {
		t.Fatalf("expected debug messages")
	}
	for _, m := range res.DebugMessages {
		if containsSecret(m) {
			t.Fatalf("debug leaked secret: %s", m)
		}
	}
}

func containsSecret(s string) bool {
	return strings.Contains(s, "super-secret") || strings.Contains(s, "token=") || strings.Contains(strings.ToLower(s), "authorization")
}
