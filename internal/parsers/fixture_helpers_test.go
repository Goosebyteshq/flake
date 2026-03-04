package parsers

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func mustReadFixture(t *testing.T, rel string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", rel)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %q: %v", path, err)
	}
	return b
}

func mustReadFrameworkFixture(t *testing.T, framework Framework) []byte {
	t.Helper()
	dir := filepath.Join("..", "..", "testdata", string(framework))
	matches, err := filepath.Glob(filepath.Join(dir, "sample.*"))
	if err != nil {
		t.Fatalf("glob fixture for framework %q: %v", framework, err)
	}
	if len(matches) == 0 {
		t.Fatalf("missing fixture for framework %q; expected %s", framework, filepath.Join("testdata", string(framework), "sample.*"))
	}
	sort.Strings(matches)
	b, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read fixture %q: %v", matches[0], err)
	}
	return b
}
