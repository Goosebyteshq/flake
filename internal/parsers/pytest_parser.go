package parsers

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

var pytestRE = regexp.MustCompile(`^(.+?::.+?)\s+(PASSED|FAILED|SKIPPED|XFAIL|XPASS)\b`)
var unittestRE = regexp.MustCompile(`^(.+\(.+\))\s+\.\.\.\s+(ok|FAIL|ERROR|skipped(?:\s+'.*')?)$`)

type PytestParser struct{}

func init() {
	registerBuiltin(PytestParser{}, registerOptions{
		Priority: 50,
		Aliases:  []string{"py.test", "unittest", "nose2", "tox"},
	})
}

func (PytestParser) Name() Framework { return FrameworkPytest }

func (PytestParser) Detect(sample []byte) bool {
	s := string(sample)
	return strings.Contains(s, "pytest") ||
		(strings.Contains(s, "::") && strings.Contains(s, "PASSED")) ||
		(strings.Contains(s, "Ran ") && strings.Contains(s, "tests") && strings.Contains(s, "..."))
}

func (PytestParser) DetectConfidence(sample []byte) int {
	s := string(sample)
	score := 0
	if strings.Contains(s, "test session starts") || strings.Contains(s, "collected ") {
		score += 60
	}
	if strings.Contains(s, "::") && (strings.Contains(s, "PASSED") || strings.Contains(s, "FAILED")) {
		score += 40
	}
	if strings.Contains(s, "Ran ") && strings.Contains(s, "tests") && strings.Contains(s, "...") {
		score += 40
	}
	return score
}

func (PytestParser) Parse(r io.Reader) ([]TestResult, error) {
	b, _, err := readAllAndRestore(r)
	if err != nil {
		return nil, err
	}
	results := make([]TestResult, 0)
	err = scanLines(strings.NewReader(string(b)), func(line string) {
		m := pytestRE.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) == 3 {
			status, ok := mapStatusToken(m[2])
			if !ok {
				return
			}
			results = append(results, TestResult{ID: TestID{Name: m[1]}, Status: status})
			return
		}
		m = unittestRE.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) != 3 {
			return
		}
		statusToken := strings.ToUpper(strings.TrimSpace(m[2]))
		switch {
		case statusToken == "OK":
			results = append(results, TestResult{ID: TestID{Name: m[1]}, Status: StatusPass})
		case strings.HasPrefix(statusToken, "SKIPPED"):
			results = append(results, TestResult{ID: TestID{Name: m[1]}, Status: StatusSkip})
		case statusToken == "FAIL" || statusToken == "ERROR":
			results = append(results, TestResult{ID: TestID{Name: m[1]}, Status: StatusFail})
		}
	})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no pytest results detected")
	}
	return results, nil
}
