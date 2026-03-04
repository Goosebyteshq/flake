package parsers

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkGoParserParse(b *testing.B) {
	for _, n := range []int{1000, 10000, 50000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			input := buildGoInput(n)
			p := GoParser{}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results, err := p.Parse(strings.NewReader(input))
				if err != nil {
					b.Fatalf("parse error: %v", err)
				}
				if len(results) != n {
					b.Fatalf("len(results)=%d want=%d", len(results), n)
				}
			}
		})
	}
}

func BenchmarkRegistryAutoParseGo(b *testing.B) {
	for _, n := range []int{1000, 10000, 50000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			input := buildGoInput(n)
			r := NewRegistry()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f, results, err := r.Parse(FrameworkAuto, strings.NewReader(input))
				if err != nil {
					b.Fatalf("parse error: %v", err)
				}
				if f != FrameworkGo {
					b.Fatalf("framework=%s", f)
				}
				if len(results) != n {
					b.Fatalf("len(results)=%d want=%d", len(results), n)
				}
			}
		})
	}
}

func buildGoInput(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		status := "PASS"
		if i%10 == 0 {
			status = "FAIL"
		}
		fmt.Fprintf(&sb, "--- %s: Test%05d (0.00s)\n", status, i)
	}
	return sb.String()
}
