package engine

import (
	"fmt"
	"sort"

	"github.com/Goosebyteshq/flake/internal/domain"
)

func AppendStatus(history string, status domain.TestStatus, window int) (string, error) {
	if window <= 0 {
		return "", fmt.Errorf("window must be > 0")
	}
	if err := status.Validate(); err != nil {
		return "", err
	}

	out := history
	switch status {
	case domain.Pass:
		out += "P"
	case domain.Fail:
		out += "F"
	case domain.Skip:
		return out, nil
	}

	if len(out) > window {
		out = out[len(out)-window:]
	}
	return out, nil
}

func ApplyRun(historyByTest map[string]string, statuses map[string]domain.TestStatus, window int) (map[string]string, error) {
	if window <= 0 {
		return nil, fmt.Errorf("window must be > 0")
	}
	out := make(map[string]string, len(historyByTest)+len(statuses))
	for k, v := range historyByTest {
		out[k] = v
	}

	keys := make([]string, 0, len(statuses))
	for k := range statuses {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, testID := range keys {
		next, err := AppendStatus(out[testID], statuses[testID], window)
		if err != nil {
			return nil, fmt.Errorf("test %q: %w", testID, err)
		}
		out[testID] = next
	}
	return out, nil
}
