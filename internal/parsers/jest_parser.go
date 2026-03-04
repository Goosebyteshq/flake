package parsers

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	jestGlyphRE = regexp.MustCompile(`^[\s>]*(✓|✔|✕|✘|○)\s+(.+)$`)
	jestWordRE  = regexp.MustCompile(`^[\s>]*(PASS|FAIL|SKIP)\s+(.+)$`)
)

type JestParser struct{}

func init() {
	registerBuiltin(JestParser{}, registerOptions{
		Priority: 60,
		Aliases:  []string{"vitest", "playwright", "jasmine", "junit5", "bun", "bun-test"},
	})
}

func (JestParser) Name() Framework { return FrameworkJest }

func (JestParser) Detect(sample []byte) bool {
	s := string(sample)
	return strings.Contains(s, "Jest") || strings.Contains(s, "Vitest") || strings.Contains(s, "Playwright") || strings.Contains(s, "Jasmine") || strings.Contains(s, "✓") || strings.Contains(s, "✕") || strings.Contains(s, "✘")
}

func (JestParser) DetectConfidence(sample []byte) int {
	s := string(sample)
	score := 0
	if strings.Contains(s, "Jest") || strings.Contains(s, "Vitest") || strings.Contains(s, "Playwright") || strings.Contains(s, "Jasmine") {
		score += 70
	}
	if strings.Contains(s, "\nPASS ") || strings.Contains(s, "\nFAIL ") || strings.HasPrefix(s, "PASS ") || strings.HasPrefix(s, "FAIL ") {
		score += 20
	}
	if strings.Contains(s, "○ ") {
		score += 10
	}
	if strings.Contains(s, "✓") || strings.Contains(s, "✕") {
		score += 10
	}
	if score > 100 {
		return 100
	}
	return score
}

func (JestParser) Parse(r io.Reader) ([]TestResult, error) {
	results := make([]TestResult, 0)
	err := scanLines(r, func(line string) {
		line = strings.TrimSpace(line)
		if line == "" {
			return
		}
		if m := jestGlyphRE.FindStringSubmatch(line); len(m) == 3 {
			status, ok := mapStatusToken(m[1])
			if ok {
				results = append(results, TestResult{ID: TestID{Name: m[2]}, Status: status})
			}
			return
		}
		if m := jestWordRE.FindStringSubmatch(line); len(m) == 3 {
			status, ok := mapStatusToken(m[1])
			if ok {
				results = append(results, TestResult{ID: TestID{Name: m[2]}, Status: status})
			}
		}
	})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no jest/vitest results detected")
	}
	return results, nil
}
