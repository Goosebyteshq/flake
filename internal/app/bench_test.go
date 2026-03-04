package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type benchClock struct{}

func (benchClock) Now() time.Time { return time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC) }

func BenchmarkScanEndToEnd(b *testing.B) {
	for _, n := range []int{1000, 10000, 50000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			d := b.TempDir()
			statePath := filepath.Join(d, "state.json")
			hintsPath := filepath.Join(d, "hints.json")
			inputPath := filepath.Join(d, "in.txt")
			var buf bytes.Buffer
			for i := 0; i < n; i++ {
				status := "PASS"
				if i%10 == 0 {
					status = "FAIL"
				}
				fmt.Fprintf(&buf, "--- %s: Test%05d (0.00s)\n", status, i)
			}
			if err := os.WriteFile(inputPath, buf.Bytes(), 0o600); err != nil {
				b.Fatalf("write input: %v", err)
			}
			scanner := Scanner{Clock: benchClock{}}
			out := bytes.Buffer{}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := scanner.Run(ScanOptions{StatePath: statePath, InputPath: inputPath, Framework: "go", Window: 50, RepoKey: d, ParserHintsPath: hintsPath}, &out); err != nil {
					b.Fatalf("scan error: %v", err)
				}
				out.Reset()
			}
		})
	}
}
