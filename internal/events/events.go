package events

import (
	"sort"

	"github.com/leomorpho/flake/internal/domain"
	"github.com/leomorpho/flake/internal/state"
)

type Payload struct {
	SchemaVersion int               `json:"schema_version"`
	PolicyVersion int               `json:"policy_version"`
	Run           RunPayload        `json:"run"`
	Tests         []TestPayload     `json:"tests"`
	Transitions   []TransitionEvent `json:"transitions"`
}

type RunPayload struct {
	RunID     string            `json:"run_id"`
	Timestamp string            `json:"timestamp"`
	Framework string            `json:"framework"`
	Meta      map[string]string `json:"meta"`
}

type TestPayload struct {
	TestID     string       `json:"test_id"`
	Derived    DerivedEvent `json:"derived"`
	Timestamps TimeEvent    `json:"timestamps"`
}

type DerivedEvent struct {
	Class       domain.Class `json:"class"`
	FailureRate float64      `json:"failure_rate"`
	SampleSize  int          `json:"sample_size"`
	Failures    int          `json:"failures"`
	Passes      int          `json:"passes"`
}

type TimeEvent struct {
	LastSeen     string `json:"last_seen,omitempty"`
	LastFailedAt string `json:"last_failed_at,omitempty"`
	LastPassedAt string `json:"last_passed_at,omitempty"`
}

type TransitionEvent struct {
	TestID      string       `json:"test_id"`
	From        domain.Class `json:"from"`
	To          domain.Class `json:"to"`
	FailureRate float64      `json:"failure_rate"`
	Severity    int          `json:"severity"`
}

func Build(st state.FileState) Payload {
	tests := make([]TestPayload, 0, len(st.Tests))
	for testID, slot := range st.Tests {
		failures := 0
		passes := 0
		for _, ch := range slot.History {
			if ch == 'F' {
				failures++
			}
			if ch == 'P' {
				passes++
			}
		}
		tests = append(tests, TestPayload{
			TestID: testID,
			Derived: DerivedEvent{
				Class:       slot.LastState.Class,
				FailureRate: slot.LastState.FailureRate,
				SampleSize:  len(slot.History),
				Failures:    failures,
				Passes:      passes,
			},
			Timestamps: TimeEvent{
				LastSeen:     slot.LastSeen,
				LastFailedAt: slot.LastFailedAt,
				LastPassedAt: slot.LastPassedAt,
			},
		})
	}
	sort.SliceStable(tests, func(i, j int) bool {
		si := domain.ClassSeverity(tests[i].Derived.Class)
		sj := domain.ClassSeverity(tests[j].Derived.Class)
		if si != sj {
			return si > sj
		}
		if tests[i].Derived.FailureRate != tests[j].Derived.FailureRate {
			return tests[i].Derived.FailureRate > tests[j].Derived.FailureRate
		}
		return tests[i].TestID < tests[j].TestID
	})

	trs := make([]TransitionEvent, 0, len(st.LastRun.Transitions))
	for _, tr := range st.LastRun.Transitions {
		trs = append(trs, TransitionEvent{
			TestID:      tr.TestID,
			From:        tr.From,
			To:          tr.To,
			FailureRate: tr.FailureRate,
			Severity:    tr.Severity,
		})
	}
	sort.SliceStable(trs, func(i, j int) bool {
		if trs[i].Severity != trs[j].Severity {
			return trs[i].Severity > trs[j].Severity
		}
		if trs[i].FailureRate != trs[j].FailureRate {
			return trs[i].FailureRate > trs[j].FailureRate
		}
		return trs[i].TestID < trs[j].TestID
	})

	meta := map[string]string{}
	for k, v := range st.LastRun.RunMeta {
		meta[k] = v
	}
	return Payload{
		SchemaVersion: st.SchemaVersion,
		PolicyVersion: st.PolicyVersion,
		Run: RunPayload{
			RunID:     st.LastRun.RunID,
			Timestamp: st.LastRun.Timestamp,
			Framework: st.LastRun.Framework,
			Meta:      meta,
		},
		Tests:       tests,
		Transitions: trs,
	}
}
