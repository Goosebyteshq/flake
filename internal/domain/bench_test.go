package domain

import (
	"fmt"
	"testing"
)

func BenchmarkDeriveStateBatch(b *testing.B) {
	for _, n := range []int{1000, 10000, 50000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			histories := make([]string, n)
			for i := 0; i < n; i++ {
				h := "PPPPPPPPPP"
				if i%10 == 0 {
					h = "FFFFFPPPPP"
				}
				histories[i] = h
			}
			prev := &DerivedState{Class: ClassFlaky, FailureRate: 0.5}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, h := range histories {
					if _, err := DeriveState(h, prev); err != nil {
						b.Fatalf("derive error: %v", err)
					}
				}
			}
		})
	}
}
