package parsers

import (
	"strings"
	"testing"
)

func TestAllParsersConformance(t *testing.T) {
	r := NewRegistry()
	for _, f := range r.SupportedFrameworks() {
		t.Run(string(f), func(t *testing.T) {
			sample := mustReadFrameworkFixture(t, f)
			detected, got, err := r.Parse(f, strings.NewReader(string(sample)))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if detected != f {
				t.Fatalf("detected=%s want=%s", detected, f)
			}
			for _, tr := range got {
				if tr.ID.Canonical() == "" {
					t.Fatalf("empty canonical id")
				}
				if tr.Status != StatusPass && tr.Status != StatusFail && tr.Status != StatusSkip {
					t.Fatalf("invalid status %s", tr.Status)
				}
			}
		})
	}
}

func TestMalformedInputsFail(t *testing.T) {
	r := NewRegistry()
	for _, f := range r.SupportedFrameworks() {
		t.Run(string(f), func(t *testing.T) {
			_, _, err := r.Parse(f, strings.NewReader("garbage"))
			if err == nil {
				t.Fatalf("expected parse error")
			}
		})
	}
}
