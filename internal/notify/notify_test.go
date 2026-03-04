package notify

import (
	"testing"

	"github.com/Goosebyteshq/flake/internal/config"
	"github.com/Goosebyteshq/flake/internal/domain"
	"github.com/Goosebyteshq/flake/internal/state"
)

func TestFilterAndSuppression(t *testing.T) {
	st := state.New()
	st.LastRun.RunID = "run-2"
	st.LastRunIndex = 2
	st.LastRun.Transitions = []domain.Transition{{TestID: "a", From: domain.ClassHealthy, To: domain.ClassFlaky, FailureRate: 0.3, Severity: 4}}
	st.Tests["a"] = state.TestSlot{Meta: map[string]string{"last_notified_class": "Flaky", "last_notified_run_index": "1"}}
	cfg := config.Default()
	items := Filter(st, cfg)
	if len(items) != 1 {
		t.Fatalf("len items")
	}
	items = ApplySuppression(&st, items, 5)
	if !items[0].Suppressed {
		t.Fatalf("expected suppressed")
	}
}

func TestChunkAndMessage(t *testing.T) {
	items := []Item{{Transition: domain.Transition{TestID: "a", From: domain.ClassHealthy, To: domain.ClassNewFail, FailureRate: 1, Severity: 3}}}
	chunks := Chunk(items, 1)
	if len(chunks) != 1 {
		t.Fatalf("chunks")
	}
	msg := BuildMessage("run-x", chunks[0])
	if msg == "" {
		t.Fatalf("message empty")
	}
}

func TestApplySuppressionMinAge(t *testing.T) {
	st := state.New()
	st.LastRunIndex = 3
	st.Tests["a"] = state.TestSlot{
		Meta: map[string]string{
			"first_seen_run_index": "2",
		},
	}
	items := []Item{{Transition: domain.Transition{TestID: "a", From: domain.ClassHealthy, To: domain.ClassNewFail, FailureRate: 1.0, Severity: 3}}}
	out := ApplySuppressionWithRules(&st, items, 0, 3, 0)
	if !out[0].Suppressed {
		t.Fatalf("expected min-age suppression")
	}
}

func TestApplySuppressionOscillation(t *testing.T) {
	st := state.New()
	st.LastRunIndex = 10
	st.Tests["a"] = state.TestSlot{
		Meta: map[string]string{
			"prev_class":                  "Flaky",
			"last_state_change_run_index": "9",
			"first_seen_run_index":        "1",
		},
	}
	items := []Item{{Transition: domain.Transition{TestID: "a", From: domain.ClassBroken, To: domain.ClassFlaky, FailureRate: 0.4, Severity: 4}}}
	out := ApplySuppressionWithRules(&st, items, 0, 0, 3)
	if !out[0].Suppressed {
		t.Fatalf("expected oscillation suppression")
	}
}
