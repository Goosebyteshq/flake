package state

import (
	"fmt"
	"sort"
	"time"

	"github.com/leomorpho/flake/internal/domain"
)

const (
	CurrentSchemaVersion = 1
	DefaultPolicyVersion = 1
	DefaultWindow        = 50
)

type FileState struct {
	SchemaVersion int                 `json:"schema_version"`
	PolicyVersion int                 `json:"policy_version"`
	Window        int                 `json:"window"`
	LastRunIndex  int                 `json:"last_run_index"`
	UpdatedAt     string              `json:"updated_at"`
	Tests         map[string]TestSlot `json:"tests"`
	LastRun       LastRun             `json:"last_run"`
}

type TestSlot struct {
	History           string            `json:"history"`
	FirstSeen         string            `json:"first_seen,omitempty"`
	LastSeen          string            `json:"last_seen,omitempty"`
	LastFailedAt      string            `json:"last_failed_at,omitempty"`
	LastPassedAt      string            `json:"last_passed_at,omitempty"`
	LastState         SlotState         `json:"last_state"`
	LastNotifiedRunID string            `json:"last_notified_run_id,omitempty"`
	Meta              map[string]string `json:"meta,omitempty"`
}

type SlotState struct {
	Class       domain.Class `json:"class"`
	FailureRate float64      `json:"failure_rate"`
}

type LastRun struct {
	RunID       string              `json:"run_id"`
	Timestamp   string              `json:"timestamp"`
	Framework   string              `json:"framework"`
	RunMeta     map[string]string   `json:"run_meta"`
	Transitions []domain.Transition `json:"transitions"`
}

func New() FileState {
	return FileState{
		SchemaVersion: CurrentSchemaVersion,
		PolicyVersion: DefaultPolicyVersion,
		Window:        DefaultWindow,
		Tests:         map[string]TestSlot{},
		LastRun: LastRun{
			RunMeta:     map[string]string{},
			Transitions: []domain.Transition{},
		},
	}
}

func (s *FileState) Normalize() {
	if s.Tests == nil {
		s.Tests = map[string]TestSlot{}
	}
	if s.LastRun.RunMeta == nil {
		s.LastRun.RunMeta = map[string]string{}
	}
	if s.LastRun.Transitions == nil {
		s.LastRun.Transitions = []domain.Transition{}
	}
}

func (s FileState) Validate() error {
	if s.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("unsupported schema_version=%d (expected %d)", s.SchemaVersion, CurrentSchemaVersion)
	}
	if s.PolicyVersion <= 0 {
		return fmt.Errorf("policy_version must be > 0")
	}
	if s.Window <= 0 {
		return fmt.Errorf("window must be > 0")
	}
	if s.UpdatedAt != "" {
		if _, err := time.Parse(time.RFC3339, s.UpdatedAt); err != nil {
			return fmt.Errorf("updated_at must be RFC3339: %w", err)
		}
	}
	for testID, slot := range s.Tests {
		for _, ch := range slot.History {
			if ch != 'P' && ch != 'F' {
				return fmt.Errorf("test %q has invalid history marker %q", testID, ch)
			}
		}
		if slot.LastState.Class == "" {
			return fmt.Errorf("test %q missing last_state.class", testID)
		}
	}
	for i := range s.LastRun.Transitions {
		tr := s.LastRun.Transitions[i]
		if tr.TestID == "" {
			return fmt.Errorf("last_run.transitions[%d].test_id is required", i)
		}
	}
	return nil
}

func StableSortTransitions(items []domain.Transition) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Severity != items[j].Severity {
			return items[i].Severity > items[j].Severity
		}
		if items[i].FailureRate != items[j].FailureRate {
			return items[i].FailureRate > items[j].FailureRate
		}
		return items[i].TestID < items[j].TestID
	})
}
