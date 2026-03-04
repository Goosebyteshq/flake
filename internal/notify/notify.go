package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/leomorpho/flake/internal/config"
	"github.com/leomorpho/flake/internal/domain"
	"github.com/leomorpho/flake/internal/state"
)

type Notifier interface {
	Send(ctx context.Context, message string) error
}

type SlackWebhookNotifier struct {
	Webhook string
	Client  *http.Client
}

func (n SlackWebhookNotifier) Send(ctx context.Context, message string) error {
	if strings.TrimSpace(n.Webhook) == "" {
		return fmt.Errorf("slack webhook is empty")
	}
	client := n.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	body, _ := json.Marshal(map[string]string{"text": message})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.Webhook, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}
	return nil
}

type Item struct {
	Transition domain.Transition
	Suppressed bool
}

func Filter(st state.FileState, cfg config.Config) []Item {
	allowed := map[string]bool{}
	for _, c := range cfg.Notify.IncludeClasses {
		allowed[c] = true
	}
	items := make([]Item, 0, len(st.LastRun.Transitions))
	for _, tr := range st.LastRun.Transitions {
		if !cfg.Notify.OnTransition {
			continue
		}
		if tr.FailureRate < cfg.Notify.MinFailureRate {
			continue
		}
		if len(allowed) > 0 && !allowed[string(tr.To)] {
			continue
		}
		items = append(items, Item{Transition: tr})
	}
	sort.SliceStable(items, func(i, j int) bool {
		a := items[i].Transition
		b := items[j].Transition
		if a.Severity != b.Severity {
			return a.Severity > b.Severity
		}
		if a.FailureRate != b.FailureRate {
			return a.FailureRate > b.FailureRate
		}
		return a.TestID < b.TestID
	})
	return items
}

func ApplySuppression(st *state.FileState, items []Item, suppressRuns int) []Item {
	return ApplySuppressionWithRules(st, items, suppressRuns, 0, 0)
}

func ApplySuppressionWithRules(st *state.FileState, items []Item, suppressRuns, minAgeRuns, oscillationWindowRuns int) []Item {
	out := make([]Item, 0, len(items))
	runIdx := st.LastRunIndex
	for _, it := range items {
		slot := st.Tests[it.Transition.TestID]
		lastClass := slot.Meta["last_notified_class"]
		lastRunIndex := atoiDefault(slot.Meta["last_notified_run_index"], -999999)
		firstSeenRunIndex := atoiDefault(slot.Meta["first_seen_run_index"], runIdx)
		lastChangeRunIndex := atoiDefault(slot.Meta["last_state_change_run_index"], runIdx)
		prevClass := slot.Meta["prev_class"]

		// Minimum age gate: suppress very fresh tests/transitions.
		if minAgeRuns > 0 && runIdx-firstSeenRunIndex < minAgeRuns {
			it.Suppressed = true
		}
		// Oscillation gate: suppress rapid class ping-pong back to a recent class.
		if !it.Suppressed && oscillationWindowRuns > 0 && prevClass == string(it.Transition.To) && runIdx-lastChangeRunIndex <= oscillationWindowRuns {
			it.Suppressed = true
		}
		if suppressRuns > 0 && lastClass == string(it.Transition.To) && runIdx-lastRunIndex < suppressRuns {
			it.Suppressed = true
		}
		out = append(out, it)
	}
	return out
}

func Unsuppressed(items []Item) []Item {
	out := make([]Item, 0, len(items))
	for _, it := range items {
		if !it.Suppressed {
			out = append(out, it)
		}
	}
	return out
}

func Chunk(items []Item, size int) [][]Item {
	if size <= 0 {
		size = 50
	}
	var out [][]Item
	for i := 0; i < len(items); i += size {
		j := i + size
		if j > len(items) {
			j = len(items)
		}
		out = append(out, items[i:j])
	}
	return out
}

func BuildMessage(runID string, chunk []Item) string {
	lines := []string{fmt.Sprintf("flake transitions for run %s", runID)}
	for _, it := range chunk {
		tr := it.Transition
		lines = append(lines, fmt.Sprintf("- %s: %s -> %s (rate=%.2f)", tr.TestID, tr.From, tr.To, tr.FailureRate))
	}
	return strings.Join(lines, "\n")
}

func MarkNotified(st *state.FileState, notified []Item) {
	for _, it := range notified {
		slot := st.Tests[it.Transition.TestID]
		if slot.Meta == nil {
			slot.Meta = map[string]string{}
		}
		slot.LastNotifiedRunID = st.LastRun.RunID
		slot.Meta["last_notified_class"] = string(it.Transition.To)
		slot.Meta["last_notified_run_index"] = fmt.Sprintf("%d", st.LastRunIndex)
		if _, ok := slot.Meta["first_seen_run_index"]; !ok {
			slot.Meta["first_seen_run_index"] = fmt.Sprintf("%d", st.LastRunIndex)
		}
		slot.Meta["prev_class"] = string(it.Transition.From)
		slot.Meta["last_state_change_run_index"] = fmt.Sprintf("%d", st.LastRunIndex)
		st.Tests[it.Transition.TestID] = slot
	}
}

func atoiDefault(v string, fallback int) int {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return fallback
		}
		n = (n * 10) + int(ch-'0')
	}
	return n
}
