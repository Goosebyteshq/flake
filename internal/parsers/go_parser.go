package parsers

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

var goLineRE = regexp.MustCompile(`^--- (PASS|FAIL|SKIP):\s+(.+?)(?:\s+\(|$)`)

type GoParser struct{}

func init() {
	registerBuiltin(GoParser{}, registerOptions{
		Priority: 30,
		Aliases:  []string{"gotest"},
	})
}

func (GoParser) Name() Framework { return FrameworkGo }

func (GoParser) Detect(sample []byte) bool {
	s := string(sample)
	return strings.Contains(s, "go test") || strings.Contains(s, "--- PASS:") || strings.Contains(s, "--- FAIL:")
}

func (GoParser) DetectConfidence(sample []byte) int {
	s := string(sample)
	score := 0
	if strings.Contains(s, "go test") {
		score += 50
	}
	if strings.Contains(s, "--- PASS:") || strings.Contains(s, "--- FAIL:") || strings.Contains(s, "--- SKIP:") {
		score += 50
	}
	return score
}

func (GoParser) Parse(r io.Reader) ([]TestResult, error) {
	results := make([]TestResult, 0)
	err := scanLines(r, func(line string) {
		m := goLineRE.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) != 3 {
			return
		}
		status, ok := mapStatusToken(m[1])
		if !ok {
			return
		}
		results = append(results, TestResult{ID: TestID{Name: m[2]}, Status: status})
	})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no go test results detected")
	}
	return results, nil
}
