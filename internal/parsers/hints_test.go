package parsers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileHintStoreRoundTrip(t *testing.T) {
	d := t.TempDir()
	store := FileHintStore{Path: filepath.Join(d, "hints.json")}

	if err := store.Put("my/repo", FrameworkGo); err != nil {
		t.Fatalf("Put error: %v", err)
	}

	got, ok, err := store.Get("my/repo")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !ok {
		t.Fatalf("expected stored hint")
	}
	if got != FrameworkGo {
		t.Fatalf("got %q, want %q", got, FrameworkGo)
	}
}

func TestFileHintStoreDeterministicFile(t *testing.T) {
	d := t.TempDir()
	path := filepath.Join(d, "hints.json")
	store := FileHintStore{Path: path}

	if err := store.Put("z-repo", FrameworkJest); err != nil {
		t.Fatalf("Put z-repo error: %v", err)
	}
	if err := store.Put("a-repo", FrameworkGo); err != nil {
		t.Fatalf("Put a-repo error: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	var hf hintFile
	if err := json.Unmarshal(b, &hf); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if len(hf.Repos) != 2 {
		t.Fatalf("len(hf.Repos)=%d, want 2", len(hf.Repos))
	}
	if hf.Repos[0].RepoKey > hf.Repos[1].RepoKey {
		t.Fatalf("expected repos to be sorted by key")
	}
}

func TestParseWithHintFallback(t *testing.T) {
	r := NewRegistry()
	input := `--- PASS: TestA (0.00s)
--- FAIL: TestB (0.00s)
`

	framework, results, err := r.ParseWithHint(FrameworkAuto, FrameworkPytest, strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseWithHint error: %v", err)
	}
	if framework != FrameworkGo {
		t.Fatalf("framework=%q, want %q", framework, FrameworkGo)
	}
	if len(results) != 2 {
		t.Fatalf("len(results)=%d, want 2", len(results))
	}
}
