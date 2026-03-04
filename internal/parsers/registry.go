package parsers

import (
	"bytes"
	"fmt"
	"io"
	"sort"
)

type Registry struct {
	ordered []Parser
	byName  map[Framework]Parser
}

type parseSuccess struct {
	framework Framework
	source    string
	score     int
	results   []TestResult
}

// NewRegistry returns builtin parsers in deterministic detect order.
func NewRegistry() *Registry {
	entries := snapshotCatalog()
	ordered := make([]Parser, 0, len(entries))
	for _, e := range entries {
		ordered = append(ordered, e.Parser)
	}
	byName := make(map[Framework]Parser, len(ordered))
	for _, p := range ordered {
		byName[p.Name()] = p
	}
	return &Registry{ordered: ordered, byName: byName}
}

func (r *Registry) SupportedFrameworks() []Framework {
	out := make([]Framework, 0, len(r.ordered))
	for _, p := range r.ordered {
		out = append(out, p.Name())
	}
	return out
}

func (r *Registry) Resolve(framework Framework, sample []byte) (Parser, error) {
	if framework == FrameworkAuto {
		for _, p := range r.ordered {
			if p.Detect(sample) {
				return p, nil
			}
		}
		return nil, fmt.Errorf("could not auto-detect framework")
	}
	p, ok := r.byName[framework]
	if !ok {
		return nil, fmt.Errorf("framework %q is not registered", framework)
	}
	return p, nil
}

func (r *Registry) Parse(framework Framework, input io.Reader) (Framework, []TestResult, error) {
	return r.ParseWithHint(framework, "", input)
}

// ParseWithHint prefers a previously successful framework first.
// If that parser fails, it falls back deterministically.
func (r *Registry) ParseWithHint(framework Framework, hinted Framework, input io.Reader) (Framework, []TestResult, error) {
	raw, rr, err := readAllAndRestore(input)
	if err != nil {
		return "", nil, err
	}
	_ = rr

	candidates, err := r.candidates(framework, hinted, raw)
	if err != nil {
		return "", nil, err
	}

	var parseErrs []error
	successes := make([]parseSuccess, 0, len(candidates))

	for _, c := range candidates {
		p, ok := r.byName[c.Framework]
		if !ok {
			continue
		}
		results, err := p.Parse(bytes.NewReader(raw))
		if err != nil {
			parseErrs = append(parseErrs, fmt.Errorf("%s: %w", p.Name(), err))
			continue
		}
		for i := range results {
			if err := results[i].Validate(); err != nil {
				parseErrs = append(parseErrs, fmt.Errorf("%s invalid result %d: %w", p.Name(), i, err))
				results = nil
				break
			}
		}
		if results == nil {
			continue
		}
		successes = append(successes, parseSuccess{
			framework: p.Name(),
			source:    c.Source,
			score:     c.Score,
			results:   results,
		})
		if framework != FrameworkAuto {
			StableSortResults(results)
			return p.Name(), results, nil
		}
	}

	if framework == FrameworkAuto && len(successes) > 0 {
		best := successes[0]
		for i := 1; i < len(successes); i++ {
			if len(successes[i].results) > len(best.results) {
				best = successes[i]
			}
		}
		merged := mergeSuccessfulResults(successes)
		StableSortResults(merged)
		return best.framework, merged, nil
	}
	if len(parseErrs) > 0 {
		msg := "all candidate parsers failed:"
		for _, pe := range parseErrs {
			msg += " " + pe.Error() + ";"
		}
		return "", nil, fmt.Errorf("%s", msg)
	}
	return "", nil, fmt.Errorf("no parser candidates available")
}

// ExplainCandidates returns the ranked candidate order used for parser attempts.
func (r *Registry) ExplainCandidates(framework Framework, hinted Framework, sample []byte) ([]CandidateInfo, error) {
	return r.candidates(framework, hinted, sample)
}

func (r *Registry) Detect(framework Framework, input io.Reader) (Framework, error) {
	if framework != FrameworkAuto {
		if _, ok := r.byName[framework]; ok {
			return framework, nil
		}
		return "", fmt.Errorf("framework %q is not registered", framework)
	}
	raw, err := io.ReadAll(input)
	if err != nil {
		return "", err
	}
	candidates, err := r.candidates(FrameworkAuto, "", raw)
	if err != nil {
		return "", err
	}
	for _, c := range candidates {
		if c.Source == "detect" || c.Source == "hint" {
			return c.Framework, nil
		}
	}
	_ = bytes.TrimSpace(raw)
	return "", fmt.Errorf("could not auto-detect framework")
}

func (r *Registry) candidates(framework Framework, hinted Framework, sample []byte) ([]CandidateInfo, error) {
	if framework != FrameworkAuto {
		if _, ok := r.byName[framework]; !ok {
			return nil, fmt.Errorf("framework %q is not registered", framework)
		}
		return []CandidateInfo{{Framework: framework, Score: 100, Source: "explicit"}}, nil
	}

	out := make([]CandidateInfo, 0, len(r.ordered))
	seen := map[Framework]bool{}

	if hinted != "" && hinted != FrameworkAuto {
		if _, ok := r.byName[hinted]; ok {
			out = append(out, CandidateInfo{Framework: hinted, Score: 101, Source: "hint"})
			seen[hinted] = true
		}
	}

	detected := make([]CandidateInfo, 0, len(r.ordered))
	index := map[Framework]int{}
	for i, p := range r.ordered {
		index[p.Name()] = i
	}
	for _, p := range r.ordered {
		if seen[p.Name()] {
			continue
		}
		score := detectConfidence(p, sample)
		if score > 0 {
			detected = append(detected, CandidateInfo{Framework: p.Name(), Score: score, Source: "detect"})
			seen[p.Name()] = true
		}
	}
	sort.SliceStable(detected, func(i, j int) bool {
		if detected[i].Score != detected[j].Score {
			return detected[i].Score > detected[j].Score
		}
		return index[detected[i].Framework] < index[detected[j].Framework]
	})
	out = append(out, detected...)

	for _, p := range r.ordered {
		if seen[p.Name()] {
			continue
		}
		out = append(out, CandidateInfo{Framework: p.Name(), Score: 0, Source: "fallback"})
	}
	return out, nil
}

func detectConfidence(p Parser, sample []byte) int {
	if cp, ok := p.(ConfidenceParser); ok {
		score := cp.DetectConfidence(sample)
		if score < 0 {
			return 0
		}
		if score > 100 {
			return 100
		}
		return score
	}
	if p.Detect(sample) {
		return 60
	}
	return 0
}

func mergeSuccessfulResults(successes []parseSuccess) []TestResult {
	if len(successes) == 0 {
		return nil
	}
	// Prefer detected/hinted/explicit parses; only use fallback parses if they are the only successes.
	useFallback := true
	for _, s := range successes {
		if s.source != "fallback" {
			useFallback = false
			break
		}
	}
	merged := make([]TestResult, 0)
	seen := map[string]bool{}
	for _, s := range successes {
		if !useFallback && s.source == "fallback" {
			continue
		}
		for _, r := range s.results {
			key := r.ID.Canonical() + "|" + string(r.Status)
			if seen[key] {
				continue
			}
			seen[key] = true
			merged = append(merged, r)
		}
	}
	return merged
}
