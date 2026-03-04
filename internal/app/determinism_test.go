package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fixedClockDet struct{ t time.Time }

func (f fixedClockDet) Now() time.Time { return f.t }

func TestDeterministicStateAndOutput(t *testing.T) {
	input := "--- PASS: A (0.00s)\n--- FAIL: B (0.00s)\n"
	now := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)
	run := func() (string, string) {
		d := t.TempDir()
		statePath := filepath.Join(d, "state.json")
		inputPath := filepath.Join(d, "in.txt")
		if err := os.WriteFile(inputPath, []byte(input), 0o600); err != nil {
			t.Fatalf("write input: %v", err)
		}
		out := bytes.Buffer{}
		_, err := (Scanner{Clock: fixedClockDet{t: now}}).Run(ScanOptions{StatePath: statePath, InputPath: inputPath, Framework: "go", Window: 50, RepoKey: d, ParserHintsPath: filepath.Join(d, "hints.json")}, &out)
		if err != nil {
			t.Fatalf("scan: %v", err)
		}
		b, err := os.ReadFile(statePath)
		if err != nil {
			t.Fatalf("read state: %v", err)
		}
		return out.String(), string(b)
	}
	o1, s1 := run()
	o2, s2 := run()
	if o1 != o2 {
		t.Fatalf("stdout mismatch\n%s\n!=\n%s", o1, o2)
	}
	if s1 != s2 {
		t.Fatalf("state mismatch\n%s\n!=\n%s", s1, s2)
	}
}
