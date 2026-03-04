package parsers

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

var surefireLineRE = regexp.MustCompile(`^(?:\[[A-Z]+\]\s+)?([A-Za-z0-9_.$-]+(?:\.[A-Za-z0-9_.$-]+)+)\s+.*<<<\s+(FAILURE|ERROR|SKIPPED|SUCCESS)\b`)
var testngLineRE = regexp.MustCompile(`^(PASSED|FAILED|SKIPPED):\s+(.+)$`)

type SurefireParser struct{}

func init() {
	registerBuiltin(SurefireParser{}, registerOptions{Priority: 35, Aliases: []string{"maven-surefire", "testng", "failsafe", "maven-failsafe"}})
}

func (SurefireParser) Name() Framework { return FrameworkSurefire }

func (SurefireParser) Detect(sample []byte) bool {
	s := string(sample)
	return strings.Contains(s, "<<< FAILURE!") ||
		strings.Contains(s, "maven-surefire-plugin") ||
		strings.Contains(s, "[INFO] --- surefire:") ||
		strings.Contains(s, "PASSED: ") ||
		strings.Contains(s, "FAILED: ")
}

func (SurefireParser) DetectConfidence(sample []byte) int {
	s := string(sample)
	score := 0
	if strings.Contains(s, "maven-surefire-plugin") {
		score += 60
	}
	if strings.Contains(s, "[INFO] --- surefire:") {
		score += 20
	}
	if strings.Contains(s, "<<< FAILURE!") || strings.Contains(s, "<<< SUCCESS!") || strings.Contains(s, "<<< SKIPPED!") {
		score += 40
	}
	if strings.Contains(s, "PASSED: ") || strings.Contains(s, "FAILED: ") {
		score += 30
	}
	return score
}

func (SurefireParser) Parse(r io.Reader) ([]TestResult, error) {
	b, _, err := readAllAndRestore(r)
	if err != nil {
		return nil, err
	}
	results := make([]TestResult, 0)
	err = scanLines(strings.NewReader(string(b)), func(line string) {
		line = strings.TrimSpace(strings.TrimSuffix(line, "!"))
		m := surefireLineRE.FindStringSubmatch(line)
		if len(m) == 3 {
			status, ok := mapStatusToken(m[2])
			if !ok {
				return
			}
			results = append(results, TestResult{ID: TestID{Name: m[1]}, Status: status})
			return
		}
		m = testngLineRE.FindStringSubmatch(line)
		if len(m) != 3 {
			return
		}
		status, ok := mapStatusToken(m[1])
		if !ok {
			return
		}
		results = append(results, TestResult{ID: TestID{Name: strings.TrimSpace(m[2])}, Status: status})
	})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no surefire results detected")
	}
	return results, nil
}
