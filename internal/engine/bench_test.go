package engine

import (
	"fmt"
	"testing"

	"github.com/leomorpho/flake/internal/domain"
)

func BenchmarkApplyRun(b *testing.B) {
	for _, n := range []int{1000, 10000, 50000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			history := make(map[string]string, n)
			statuses := make(map[string]domain.TestStatus, n)
			for i := 0; i < n; i++ {
				id := fmt.Sprintf("Test%05d", i)
				history[id] = "PPPPPP"
				if i%10 == 0 {
					statuses[id] = domain.Fail
				} else {
					statuses[id] = domain.Pass
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				got, err := ApplyRun(history, statuses, 50)
				if err != nil {
					b.Fatalf("apply error: %v", err)
				}
				if len(got) != n {
					b.Fatalf("len=%d want=%d", len(got), n)
				}
			}
		})
	}
}
