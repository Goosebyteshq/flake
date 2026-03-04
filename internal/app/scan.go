package app

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/leomorpho/flake/internal/config"
	"github.com/leomorpho/flake/internal/domain"
	"github.com/leomorpho/flake/internal/engine"
	"github.com/leomorpho/flake/internal/events"
	"github.com/leomorpho/flake/internal/parsers"
	"github.com/leomorpho/flake/internal/runmeta"
	"github.com/leomorpho/flake/internal/state"
)

type ScanOptions struct {
	StatePath       string
	InputPath       string
	Framework       string
	Window          int
	EventsPath      string
	ParserHintsPath string
	RepoKey         string
	RunMetaPath     string
	ConfigPath      string
	Debug           bool
}

type ScanResult struct {
	RunID           string                      `json:"run_id"`
	Timestamp       string                      `json:"timestamp"`
	Framework       string                      `json:"framework"`
	TestsParsed     int                         `json:"tests_parsed"`
	Transitions     []domain.Transition         `json:"transitions"`
	ClassCounts     map[string]int              `json:"class_counts"`
	Policy          domain.ClassificationPolicy `json:"policy"`
	Classifications []ClassificationDetail      `json:"classifications"`
	DebugMessages   []string                    `json:"debug_messages,omitempty"`
}

type Scanner struct {
	Clock     state.Clock
	Registry  *parsers.Registry
	HintStore parsers.HintStore
}

type ClassificationDetail struct {
	TestID              string                           `json:"test_id"`
	History             string                           `json:"history"`
	Derived             domain.DerivedState              `json:"derived"`
	Explanation         domain.ClassificationExplanation `json:"explanation"`
	PreviousClass       domain.Class                     `json:"previous_class,omitempty"`
	PreviousFailureRate float64                          `json:"previous_failure_rate,omitempty"`
}

func (s Scanner) Run(opts ScanOptions, stdout io.Writer) (ScanResult, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return ScanResult{}, err
	}
	framework, err := parsers.ParseFramework(opts.Framework)
	if err != nil {
		return ScanResult{}, err
	}

	input, err := readInput(opts.InputPath)
	if err != nil {
		return ScanResult{}, err
	}
	if strings.TrimSpace(string(input)) == "" {
		return ScanResult{}, fmt.Errorf("empty input")
	}

	reg := s.Registry
	if reg == nil {
		reg = parsers.NewRegistry()
	}
	hintStore := s.HintStore
	if hintStore == nil {
		hintStore = parsers.FileHintStore{Path: opts.ParserHintsPath}
	}
	hinted := parsers.Framework("")
	if h, ok, err := hintStore.Get(opts.RepoKey); err == nil && ok {
		hinted = h
	}
	ranked, _ := reg.ExplainCandidates(framework, hinted, input)

	detected, parsed, err := reg.ParseWithHint(framework, hinted, strings.NewReader(string(input)))
	if err != nil {
		return ScanResult{}, err
	}
	_ = hintStore.Put(opts.RepoKey, detected)

	store := state.Store{Path: opts.StatePath, Clock: s.Clock}
	st, err := store.Load()
	if err != nil {
		return ScanResult{}, err
	}
	if opts.Window > 0 {
		st.Window = opts.Window
	} else if cfg.Window > 0 {
		st.Window = cfg.Window
	}
	st.PolicyVersion = cfg.PolicyVersion
	classificationPolicy := cfg.ClassificationPolicy()

	statuses := collapseStatuses(parsed)
	histories := make(map[string]string, len(st.Tests))
	for testID, slot := range st.Tests {
		histories[testID] = slot.History
	}
	nextHistories, err := engine.ApplyRun(histories, statuses, st.Window)
	if err != nil {
		return ScanResult{}, err
	}

	now := s.now().UTC().Format(time.RFC3339)
	previous := map[string]domain.DerivedState{}
	current := map[string]domain.DerivedState{}
	classCounts := map[string]int{}
	classifications := make([]ClassificationDetail, 0, len(statuses))
	for testID, status := range statuses {
		slot := st.Tests[testID]
		prev := slotToDerived(slot)
		if slot.LastState.Class != "" {
			previous[testID] = prev
		}
		next, explain, err := domain.DeriveStateExplainedWithPolicy(nextHistories[testID], maybePrev(slot), classificationPolicy)
		if err != nil {
			return ScanResult{}, fmt.Errorf("derive state for %q: %w", testID, err)
		}
		current[testID] = next
		classCounts[string(next.Class)]++
		detail := ClassificationDetail{
			TestID:      testID,
			History:     nextHistories[testID],
			Derived:     next,
			Explanation: explain,
		}
		if slot.LastState.Class != "" {
			detail.PreviousClass = slot.LastState.Class
			detail.PreviousFailureRate = slot.LastState.FailureRate
		}
		classifications = append(classifications, detail)

		slot.History = nextHistories[testID]
		if slot.FirstSeen == "" {
			slot.FirstSeen = now
		}
		slot.LastSeen = now
		slot.LastState = state.SlotState{Class: next.Class, FailureRate: next.FailureRate}
		switch status {
		case domain.Fail:
			slot.LastFailedAt = now
		case domain.Pass:
			slot.LastPassedAt = now
		}
		st.Tests[testID] = slot
	}
	sort.SliceStable(classifications, func(i, j int) bool {
		return classifications[i].TestID < classifications[j].TestID
	})

	transitions := domain.DetectTransitions(current, previous)
	runID := deterministicRunID(now, detected, statuses)
	st.LastRunIndex++
	rm := runmeta.FromEnv()
	if cfg.RunMeta.AllowFileOverride && strings.TrimSpace(opts.RunMetaPath) != "" {
		fileMeta, err := runmeta.FromFile(opts.RunMetaPath)
		if err != nil {
			return ScanResult{}, err
		}
		rm = runmeta.Merge(!cfg.RunMeta.PreferEnv, rm, fileMeta)
	}
	st.LastRun = state.LastRun{
		RunID:       runID,
		Timestamp:   now,
		Framework:   string(detected),
		RunMeta:     rm.ToMap(),
		Transitions: transitions,
	}
	if err := store.Save(st); err != nil {
		return ScanResult{}, err
	}

	if strings.TrimSpace(opts.EventsPath) != "" {
		payload := events.Build(st)
		if err := writeEvents(opts.EventsPath, payload); err != nil {
			return ScanResult{}, err
		}
	}

	debug := []string{}
	if opts.Debug {
		debug = append(debug, fmt.Sprintf("hinted_parser=%s", hinted))
		debug = append(debug, fmt.Sprintf("detected_parser=%s", detected))
		if len(ranked) > 0 {
			parts := make([]string, 0, len(ranked))
			for _, c := range ranked {
				parts = append(parts, fmt.Sprintf("%s:%d:%s", c.Framework, c.Score, c.Source))
			}
			debug = append(debug, "parser_candidates="+strings.Join(parts, ","))
		}
		fmt.Fprintf(stdout, "debug: hinted=%s detected=%s\n", hinted, detected)
		if len(ranked) > 0 {
			fmt.Fprintf(stdout, "debug: candidates=%s\n", strings.TrimPrefix(debug[len(debug)-1], "parser_candidates="))
		}
	}
	fmt.Fprintf(stdout, "scan: framework=%s tests=%d transitions=%d\n", detected, len(statuses), len(transitions))
	return ScanResult{
		RunID:           runID,
		Timestamp:       now,
		Framework:       string(detected),
		TestsParsed:     len(statuses),
		Transitions:     transitions,
		ClassCounts:     classCounts,
		Policy:          classificationPolicy,
		Classifications: classifications,
		DebugMessages:   debug,
	}, nil
}

func writeEvents(path string, payload events.Payload) error {
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if path == "-" {
		_, err = os.Stdout.Write(b)
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func maybePrev(slot state.TestSlot) *domain.DerivedState {
	if slot.LastState.Class == "" {
		return nil
	}
	p := slotToDerived(slot)
	return &p
}

func slotToDerived(slot state.TestSlot) domain.DerivedState {
	failures := strings.Count(slot.History, "F")
	passes := strings.Count(slot.History, "P")
	n := len(slot.History)
	return domain.DerivedState{
		Class:       slot.LastState.Class,
		FailureRate: slot.LastState.FailureRate,
		Failures:    failures,
		Passes:      passes,
		SampleSize:  n,
	}
}

func deterministicRunID(timestamp string, framework parsers.Framework, statuses map[string]domain.TestStatus) string {
	keys := make([]string, 0, len(statuses))
	for k := range statuses {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha1.New()
	_, _ = h.Write([]byte(timestamp))
	_, _ = h.Write([]byte("|" + string(framework) + "|"))
	for _, k := range keys {
		_, _ = h.Write([]byte(k + ":" + string(statuses[k]) + ";"))
	}
	sum := hex.EncodeToString(h.Sum(nil))
	return "run-" + sum[:16]
}

func collapseStatuses(results []parsers.TestResult) map[string]domain.TestStatus {
	out := map[string]domain.TestStatus{}
	for _, r := range results {
		id := r.ID.Canonical()
		cand := toDomainStatus(r.Status)
		if existing, ok := out[id]; ok {
			out[id] = worseStatus(existing, cand)
			continue
		}
		out[id] = cand
	}
	return out
}

func worseStatus(a, b domain.TestStatus) domain.TestStatus {
	rank := func(v domain.TestStatus) int {
		switch v {
		case domain.Fail:
			return 3
		case domain.Pass:
			return 2
		case domain.Skip:
			return 1
		default:
			return 0
		}
	}
	if rank(a) >= rank(b) {
		return a
	}
	return b
}

func toDomainStatus(s parsers.TestStatus) domain.TestStatus {
	switch s {
	case parsers.StatusFail:
		return domain.Fail
	case parsers.StatusSkip:
		return domain.Skip
	default:
		return domain.Pass
	}
}

func readInput(path string) ([]byte, error) {
	if strings.TrimSpace(path) != "" {
		return os.ReadFile(path)
	}
	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if (info.Mode() & os.ModeCharDevice) != 0 {
		return nil, fmt.Errorf("no stdin input provided; use --input")
	}
	return io.ReadAll(os.Stdin)
}

func (s Scanner) now() time.Time {
	if s.Clock == nil {
		return time.Now()
	}
	return s.Clock.Now()
}
