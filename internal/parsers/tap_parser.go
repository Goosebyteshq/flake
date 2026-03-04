package parsers

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

var tapLineRE = regexp.MustCompile(`^(ok|not ok)\s+\d+\s*-?\s*(.+)$`)

type TAPParser struct{}

func init() {
	registerBuiltin(TAPParser{}, registerOptions{
		Priority: 70,
		Aliases:  []string{"node-tap", "prove", "bats", "perl", "node-test", "node-test-runner", "deno", "deno-test"},
	})
}

func (TAPParser) Name() Framework { return FrameworkTAP }

func (TAPParser) Detect(sample []byte) bool {
	s := string(sample)
	return strings.Contains(s, "TAP version") || tapLineRE.Match(sample)
}

func (TAPParser) DetectConfidence(sample []byte) int {
	s := string(sample)
	score := 0
	if strings.Contains(s, "TAP version") {
		score += 70
	}
	if tapLineRE.Match(sample) {
		score += 30
	}
	return score
}

func (TAPParser) Parse(r io.Reader) ([]TestResult, error) {
	results := make([]TestResult, 0)
	err := scanLines(r, func(line string) {
		line = strings.TrimSpace(line)
		m := tapLineRE.FindStringSubmatch(line)
		if len(m) != 3 {
			return
		}
		name := strings.TrimSpace(m[2])
		if name == "" {
			name = fmt.Sprintf("tap#%d", len(results)+1)
		}
		status := StatusPass
		if strings.EqualFold(m[1], "not ok") {
			status = StatusFail
		}
		if strings.Contains(strings.ToLower(name), "# skip") {
			status = StatusSkip
		}
		results = append(results, TestResult{ID: TestID{Name: name}, Status: status})
	})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no tap results detected")
	}
	return results, nil
}
