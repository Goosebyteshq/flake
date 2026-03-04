package domain

import "sort"

func DetectTransitions(current map[string]DerivedState, previous map[string]DerivedState) []Transition {
	out := make([]Transition, 0)
	for testID, curr := range current {
		prev, ok := previous[testID]
		if !ok {
			continue
		}
		if prev.Class == curr.Class {
			continue
		}
		out = append(out, Transition{
			TestID:      testID,
			From:        prev.Class,
			To:          curr.Class,
			FailureRate: curr.FailureRate,
			Severity:    ClassSeverity(curr.Class),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Severity != out[j].Severity {
			return out[i].Severity > out[j].Severity
		}
		if out[i].FailureRate != out[j].FailureRate {
			return out[i].FailureRate > out[j].FailureRate
		}
		return out[i].TestID < out[j].TestID
	})
	return out
}
