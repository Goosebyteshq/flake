package domain

import (
	"fmt"
)

func DeriveState(history string, previous *DerivedState) (DerivedState, error) {
	return DeriveStateWithPolicy(history, previous, DefaultClassificationPolicy())
}

type ClassificationExplanation struct {
	BaseClass         Class    `json:"base_class"`
	RecoveringApplied bool     `json:"recovering_applied"`
	Reasons           []string `json:"reasons,omitempty"`
}

func DeriveStateWithPolicy(history string, previous *DerivedState, policy ClassificationPolicy) (DerivedState, error) {
	state, _, err := DeriveStateExplainedWithPolicy(history, previous, policy)
	return state, err
}

func DeriveStateExplainedWithPolicy(history string, previous *DerivedState, policy ClassificationPolicy) (DerivedState, ClassificationExplanation, error) {
	if err := policy.Validate(); err != nil {
		return DerivedState{}, ClassificationExplanation{}, err
	}
	failures := 0
	passes := 0
	for _, ch := range history {
		switch ch {
		case 'F':
			failures++
		case 'P':
			passes++
		default:
			return DerivedState{}, ClassificationExplanation{}, fmt.Errorf("history contains invalid marker %q", ch)
		}
	}
	n := len(history)
	rate := 0.0
	if n > 0 {
		rate = float64(failures) / float64(n)
	}

	baseClass := classifyBase(failures, passes, n, rate, policy)
	state := DerivedState{
		Class:       baseClass,
		FailureRate: rate,
		Failures:    failures,
		Passes:      passes,
		SampleSize:  n,
	}
	explain := ClassificationExplanation{
		BaseClass:         baseClass,
		RecoveringApplied: false,
		Reasons:           baseReasons(failures, passes, n, rate, policy),
	}

	if shouldRecover(previous, baseClass, rate, policy) {
		state.Class = ClassRecovering
		explain.RecoveringApplied = true
		explain.Reasons = append(explain.Reasons, recoveringReasons(previous, baseClass, rate, policy)...)
	}
	return state, explain, nil
}

func classifyBase(failures, passes, sampleSize int, rate float64, policy ClassificationPolicy) Class {
	if failures == 0 {
		return ClassHealthy
	}
	if sampleSize < policy.NewFailSampleMax && failures > 0 {
		return ClassNewFail
	}
	if failures >= policy.FlakyMinFailures && passes > 0 && rate >= policy.FlakyMinRate && rate <= policy.FlakyMaxRate {
		return ClassFlaky
	}
	if rate > policy.BrokenMinRate {
		return ClassBroken
	}
	return ClassHealthy
}

func shouldRecover(previous *DerivedState, currentBase Class, currentRate float64, policy ClassificationPolicy) bool {
	if !policy.EnableRecovering {
		return false
	}
	if previous == nil {
		return false
	}
	if !isTroubledClass(previous.Class) {
		return false
	}
	rateDrop := previous.FailureRate - currentRate
	if rateDrop >= policy.RecoveringMinDrop {
		return true
	}
	return ClassSeverity(currentBase) < ClassSeverity(previous.Class)
}

func isTroubledClass(class Class) bool {
	switch class {
	case ClassBroken, ClassFlaky, ClassNewFail:
		return true
	default:
		return false
	}
}

func ClassSeverity(class Class) int {
	switch class {
	case ClassBroken:
		return 5
	case ClassFlaky:
		return 4
	case ClassNewFail:
		return 3
	case ClassRecovering:
		return 2
	case ClassHealthy:
		return 1
	default:
		return 0
	}
}

func baseReasons(failures, passes, sampleSize int, rate float64, policy ClassificationPolicy) []string {
	if failures == 0 {
		return []string{"healthy_no_failures"}
	}
	if sampleSize < policy.NewFailSampleMax && failures > 0 {
		return []string{"newfail_small_sample"}
	}
	if failures >= policy.FlakyMinFailures && passes > 0 && rate >= policy.FlakyMinRate && rate <= policy.FlakyMaxRate {
		return []string{"flaky_mixed_history_rate_band"}
	}
	if rate > policy.BrokenMinRate {
		return []string{"broken_high_failure_rate"}
	}
	return []string{"healthy_default_gap_rule"}
}

func recoveringReasons(previous *DerivedState, currentBase Class, currentRate float64, policy ClassificationPolicy) []string {
	if previous == nil {
		return nil
	}
	out := make([]string, 0, 2)
	rateDrop := previous.FailureRate - currentRate
	if rateDrop >= policy.RecoveringMinDrop {
		out = append(out, "recovering_rate_drop")
	}
	if ClassSeverity(currentBase) < ClassSeverity(previous.Class) {
		out = append(out, "recovering_class_boundary_down")
	}
	return out
}
