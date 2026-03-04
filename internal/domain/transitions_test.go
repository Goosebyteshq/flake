package domain

import "testing"

func TestDetectTransitionsStableOrder(t *testing.T) {
	previous := map[string]DerivedState{
		"a": {Class: ClassHealthy, FailureRate: 0.0},
		"b": {Class: ClassHealthy, FailureRate: 0.0},
		"c": {Class: ClassFlaky, FailureRate: 0.3},
		"d": {Class: ClassBroken, FailureRate: 0.9},
	}
	current := map[string]DerivedState{
		"a": {Class: ClassBroken, FailureRate: 0.95},
		"b": {Class: ClassBroken, FailureRate: 0.85},
		"c": {Class: ClassRecovering, FailureRate: 0.1},
		"d": {Class: ClassBroken, FailureRate: 0.92}, // unchanged class, ignored
	}

	got := DetectTransitions(current, previous)
	if len(got) != 3 {
		t.Fatalf("len(got)=%d, want 3", len(got))
	}
	if got[0].TestID != "a" || got[1].TestID != "b" {
		t.Fatalf("expected broken transitions sorted by rate desc, got %+v", got)
	}
	if got[2].TestID != "c" {
		t.Fatalf("expected recovering transition last, got %+v", got)
	}
}

func TestDetectTransitionsNoPreviousNoTransition(t *testing.T) {
	got := DetectTransitions(
		map[string]DerivedState{"x": {Class: ClassNewFail, FailureRate: 1.0}},
		map[string]DerivedState{},
	)
	if len(got) != 0 {
		t.Fatalf("expected no transitions, got %d", len(got))
	}
}
