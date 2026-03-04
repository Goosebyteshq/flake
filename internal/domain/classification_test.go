package domain

import "testing"

func TestDeriveStateBoundaries(t *testing.T) {
	policy := DefaultClassificationPolicy()
	cases := []struct {
		name    string
		history string
		want    Class
		rate    float64
	}{
		{name: "healthy no failures", history: "PPPP", want: ClassHealthy, rate: 0.0},
		{name: "newfail one fail", history: "F", want: ClassNewFail, rate: 1.0},
		{name: "newfail short window", history: "PF", want: ClassNewFail, rate: 0.5},
		{name: "flaky middle rate", history: "PPPPFFFF", want: ClassFlaky, rate: 0.5},
		{name: "broken above threshold", history: "FFFFFP", want: ClassBroken, rate: 5.0 / 6.0},
		{name: "healthy gap class", history: "FFFPP", want: ClassHealthy, rate: 3.0 / 5.0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DeriveStateWithPolicy(tc.history, nil, policy)
			if err != nil {
				t.Fatalf("DeriveState error: %v", err)
			}
			if got.Class != tc.want {
				t.Fatalf("class=%q, want %q", got.Class, tc.want)
			}
			if got.FailureRate != tc.rate {
				t.Fatalf("failureRate=%v, want %v", got.FailureRate, tc.rate)
			}
		})
	}
}

func TestDeriveStateRecoveringOverride(t *testing.T) {
	policy := DefaultClassificationPolicy()
	cases := []struct {
		name     string
		previous DerivedState
		history  string
		want     Class
	}{
		{
			name:     "rate drop >= 0.20",
			previous: DerivedState{Class: ClassBroken, FailureRate: 0.9},
			history:  "FFPPPPPPPP", // 0.2
			want:     ClassRecovering,
		},
		{
			name:     "crosses class boundary downward",
			previous: DerivedState{Class: ClassFlaky, FailureRate: 0.3},
			history:  "PPPP", // healthy
			want:     ClassRecovering,
		},
		{
			name:     "not troubled previous class",
			previous: DerivedState{Class: ClassHealthy, FailureRate: 0.0},
			history:  "PPPP",
			want:     ClassHealthy,
		},
		{
			name:     "troubled but not improved",
			previous: DerivedState{Class: ClassBroken, FailureRate: 0.95},
			history:  "FFFFFP", // still broken, minor drop
			want:     ClassBroken,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DeriveStateWithPolicy(tc.history, &tc.previous, policy)
			if err != nil {
				t.Fatalf("DeriveState error: %v", err)
			}
			if got.Class != tc.want {
				t.Fatalf("class=%q, want %q", got.Class, tc.want)
			}
		})
	}
}

func TestDeriveStateInvalidHistory(t *testing.T) {
	_, err := DeriveStateWithPolicy("PXF", nil, DefaultClassificationPolicy())
	if err == nil {
		t.Fatalf("expected invalid history error")
	}
}

func TestDeriveStatePolicyOverrides(t *testing.T) {
	p := DefaultClassificationPolicy()
	p.BrokenMinRate = 0.60
	state, err := DeriveStateWithPolicy("FFFPP", nil, p) // 0.60
	if err != nil {
		t.Fatalf("DeriveStateWithPolicy error: %v", err)
	}
	if state.Class != ClassHealthy {
		t.Fatalf("expected Healthy because broken is strictly > threshold, got %s", state.Class)
	}
	state, err = DeriveStateWithPolicy("FFFFPP", nil, p) // 0.66
	if err != nil {
		t.Fatalf("DeriveStateWithPolicy error: %v", err)
	}
	if state.Class != ClassBroken {
		t.Fatalf("expected Broken with lowered threshold, got %s", state.Class)
	}
}

func TestDeriveStateRecoveringToggle(t *testing.T) {
	p := DefaultClassificationPolicy()
	p.EnableRecovering = false
	prev := DerivedState{Class: ClassBroken, FailureRate: 0.95}
	state, err := DeriveStateWithPolicy("FFPPPPPPPP", &prev, p)
	if err != nil {
		t.Fatalf("DeriveStateWithPolicy error: %v", err)
	}
	if state.Class == ClassRecovering {
		t.Fatalf("recovering must be disabled by policy")
	}
}

func TestPolicyValidation(t *testing.T) {
	p := DefaultClassificationPolicy()
	p.FlakyMinRate = 0.8
	p.FlakyMaxRate = 0.2
	if _, err := DeriveStateWithPolicy("PPFF", nil, p); err == nil {
		t.Fatalf("expected policy validation error")
	}
}

func TestClassSeverity(t *testing.T) {
	if ClassSeverity(ClassBroken) <= ClassSeverity(ClassFlaky) {
		t.Fatalf("broken severity should be higher than flaky")
	}
	if ClassSeverity(ClassNewFail) <= ClassSeverity(ClassRecovering) {
		t.Fatalf("newfail severity should be higher than recovering")
	}
}

func TestDeriveStateExplainedReasons(t *testing.T) {
	p := DefaultClassificationPolicy()
	state, explain, err := DeriveStateExplainedWithPolicy("FFFFFP", nil, p)
	if err != nil {
		t.Fatalf("DeriveStateExplainedWithPolicy error: %v", err)
	}
	if state.Class != ClassBroken {
		t.Fatalf("class=%s, want %s", state.Class, ClassBroken)
	}
	if explain.BaseClass != ClassBroken {
		t.Fatalf("base_class=%s, want %s", explain.BaseClass, ClassBroken)
	}
	if len(explain.Reasons) == 0 || explain.Reasons[0] != "broken_high_failure_rate" {
		t.Fatalf("unexpected reasons: %+v", explain.Reasons)
	}
}

func TestDeriveStateExplainedRecoveringReason(t *testing.T) {
	p := DefaultClassificationPolicy()
	prev := DerivedState{Class: ClassBroken, FailureRate: 1.0}
	state, explain, err := DeriveStateExplainedWithPolicy("PPPPF", &prev, p)
	if err != nil {
		t.Fatalf("DeriveStateExplainedWithPolicy error: %v", err)
	}
	if state.Class != ClassRecovering {
		t.Fatalf("class=%s, want %s", state.Class, ClassRecovering)
	}
	if !explain.RecoveringApplied {
		t.Fatalf("expected recovering_applied")
	}
	found := false
	for _, reason := range explain.Reasons {
		if reason == "recovering_rate_drop" || reason == "recovering_class_boundary_down" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected recovering reason, got %+v", explain.Reasons)
	}
}
