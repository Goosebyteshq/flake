package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPerformance10k(t *testing.T) {
	d := t.TempDir()
	statePath := filepath.Join(d, "state.json")
	inputPath := filepath.Join(d, "in.txt")

	buf := bytes.Buffer{}
	for i := 0; i < 10000; i++ {
		status := "PASS"
		if i%10 == 0 {
			status = "FAIL"
		}
		fmt.Fprintf(&buf, "--- %s: Test%05d (0.00s)\n", status, i)
	}
	if err := os.WriteFile(inputPath, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
	start := time.Now()
	out := bytes.Buffer{}
	_, err := (Scanner{}).Run(ScanOptions{StatePath: statePath, InputPath: inputPath, Framework: "go", Window: 50, RepoKey: d, ParserHintsPath: filepath.Join(d, "hints.json")}, &out)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("scan too slow: %v", time.Since(start))
	}
}
