package state

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Goosebyteshq/flake/internal/domain"
)

func BenchmarkStateSaveLoad(b *testing.B) {
	for _, n := range []int{1000, 10000, 50000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			d := b.TempDir()
			path := filepath.Join(d, "state.json")
			s := Store{Path: path}
			st := New()
			for i := 0; i < n; i++ {
				id := fmt.Sprintf("Test%05d", i)
				st.Tests[id] = TestSlot{History: "PPPPPF", LastState: SlotState{Class: domain.ClassFlaky, FailureRate: 1.0 / 6.0}}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := s.Save(st); err != nil {
					b.Fatalf("save error: %v", err)
				}
				got, err := s.Load()
				if err != nil {
					b.Fatalf("load error: %v", err)
				}
				if len(got.Tests) != n {
					b.Fatalf("len=%d want=%d", len(got.Tests), n)
				}
			}
		})
	}
}
