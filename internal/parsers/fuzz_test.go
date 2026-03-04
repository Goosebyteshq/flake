package parsers

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func FuzzRegistryParseAutoDeterministic(f *testing.F) {
	r := NewRegistry()

	for _, rel := range []string{
		"go/sample.txt",
		"pytest/sample.txt",
		"jest/sample.txt",
		"junitxml/sample.xml",
		"tap/sample.txt",
		"cargo/sample.txt",
		"trx/sample.xml",
		"surefire/sample.txt",
		"gradle/sample.txt",
		"nunitxml/sample.xml",
		"mocha/sample.txt",
		"mixed/go_pytest.txt",
	} {
		addFixtureSeed(f, rel)
	}

	f.Add([]byte("ok 1 - smoke\n1..1\n"))
	f.Add([]byte("=== RUN   TestSmoke\n--- PASS: TestSmoke (0.00s)\nPASS\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 2<<20 {
			t.Skip()
		}

		f1, got1, err1 := r.Parse(FrameworkAuto, bytes.NewReader(data))
		f2, got2, err2 := r.Parse(FrameworkAuto, bytes.NewReader(data))

		if (err1 == nil) != (err2 == nil) {
			t.Fatalf("nondeterministic error outcome: err1=%v err2=%v", err1, err2)
		}
		if f1 != f2 {
			t.Fatalf("nondeterministic framework: %q vs %q", f1, f2)
		}
		if !reflect.DeepEqual(got1, got2) {
			t.Fatalf("nondeterministic results")
		}
		if err1 != nil {
			return
		}
		if strings.TrimSpace(string(f1)) == "" {
			t.Fatalf("empty detected framework on success")
		}

		lastKey := ""
		for i, tr := range got1 {
			if err := tr.Validate(); err != nil {
				t.Fatalf("invalid result[%d]: %v", i, err)
			}
			if strings.TrimSpace(tr.ID.Canonical()) == "" {
				t.Fatalf("empty canonical id at index %d", i)
			}
			key := tr.ID.Canonical() + "|" + string(tr.Status)
			if key < lastKey {
				t.Fatalf("results not stable-sorted: %q before %q", key, lastKey)
			}
			lastKey = key
		}
	})
}

func FuzzRegistryDetectAutoDeterministic(f *testing.F) {
	r := NewRegistry()
	supported := make(map[Framework]bool)
	for _, framework := range r.SupportedFrameworks() {
		supported[framework] = true
	}

	for _, rel := range []string{
		"go/sample.txt",
		"pytest/sample.txt",
		"jest/sample.txt",
		"junitxml/sample.xml",
		"tap/sample.txt",
		"trx/sample.xml",
		"mixed/go_pytest.txt",
	} {
		addFixtureSeed(f, rel)
	}

	f.Add([]byte("garbage"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 2<<20 {
			t.Skip()
		}
		f1, err1 := r.Detect(FrameworkAuto, bytes.NewReader(data))
		f2, err2 := r.Detect(FrameworkAuto, bytes.NewReader(data))
		if (err1 == nil) != (err2 == nil) {
			t.Fatalf("nondeterministic detect error outcome: err1=%v err2=%v", err1, err2)
		}
		if f1 != f2 {
			t.Fatalf("nondeterministic detect framework: %q vs %q", f1, f2)
		}
		if err1 != nil {
			return
		}
		if !supported[f1] {
			t.Fatalf("detect returned unsupported framework %q", f1)
		}
	})
}

func addFixtureSeed(f *testing.F, rel string) {
	path := filepath.Join("..", "..", "testdata", rel)
	b, err := os.ReadFile(path)
	if err == nil && len(b) > 0 {
		f.Add(b)
	}
}
