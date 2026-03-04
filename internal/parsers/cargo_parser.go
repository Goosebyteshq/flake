package parsers

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

var cargoLineRE = regexp.MustCompile(`^test\s+(.+?)\s+\.\.\.\s+(ok|FAILED|ignored)$`)

type CargoParser struct{}

func init() {
	registerBuiltin(CargoParser{}, registerOptions{
		Priority: 40,
		Aliases:  []string{"rust", "nextest"},
	})
}

func (CargoParser) Name() Framework { return FrameworkCargo }

func (CargoParser) Detect(sample []byte) bool {
	s := string(sample)
	return strings.Contains(s, "cargo test") || strings.Contains(s, "test result:") && strings.Contains(s, "FAILED")
}

func (CargoParser) DetectConfidence(sample []byte) int {
	s := string(sample)
	score := 0
	if strings.Contains(s, "cargo test") {
		score += 60
	}
	if strings.Contains(s, "test result:") {
		score += 20
	}
	if strings.Contains(s, "... FAILED") || strings.Contains(s, "... ok") {
		score += 20
	}
	return score
}

func (CargoParser) Parse(r io.Reader) ([]TestResult, error) {
	results := make([]TestResult, 0)
	err := scanLines(r, func(line string) {
		m := cargoLineRE.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) != 3 {
			return
		}
		status, ok := mapStatusToken(m[2])
		if !ok {
			return
		}
		results = append(results, TestResult{ID: TestID{Name: m[1]}, Status: status})
	})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no cargo test results detected")
	}
	return results, nil
}
