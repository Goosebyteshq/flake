package parsers

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

var gradleLineRE = regexp.MustCompile(`^(.+?)\s*>\s*(.+?)\s+(PASSED|FAILED|SKIPPED)$`)

type GradleParser struct{}

func init() {
	registerBuiltin(GradleParser{}, registerOptions{Priority: 45, Aliases: []string{"gradle-test", "kotest", "sbt"}})
}

func (GradleParser) Name() Framework { return FrameworkGradle }

func (GradleParser) Detect(sample []byte) bool {
	s := string(sample)
	return strings.Contains(s, " > ") && (strings.Contains(s, " FAILED") || strings.Contains(s, " PASSED"))
}

func (GradleParser) DetectConfidence(sample []byte) int {
	s := string(sample)
	if !strings.Contains(s, " > ") {
		return 0
	}
	score := 50
	if strings.Contains(s, " FAILED") || strings.Contains(s, " PASSED") || strings.Contains(s, " SKIPPED") {
		score += 50
	}
	return score
}

func (GradleParser) Parse(r io.Reader) ([]TestResult, error) {
	results := make([]TestResult, 0)
	err := scanLines(r, func(line string) {
		m := gradleLineRE.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) != 4 {
			return
		}
		status, ok := mapStatusToken(m[3])
		if !ok {
			return
		}
		name := strings.TrimSpace(m[1] + "::" + m[2])
		results = append(results, TestResult{ID: TestID{Name: name}, Status: status})
	})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no gradle test results detected")
	}
	return results, nil
}
