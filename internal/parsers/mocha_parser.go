package parsers

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	mochaGlyphRE  = regexp.MustCompile(`^[\s>]*(✓|✔|✖|x|-)\s+(.+)$`)
	mochaNumberRE = regexp.MustCompile(`^\d+\)\s+(.+)$`)
)

type MochaParser struct{}

func init() {
	registerBuiltin(MochaParser{}, registerOptions{Priority: 58, Aliases: []string{"ava", "mocha-ava", "cypress"}})
}

func (MochaParser) Name() Framework { return FrameworkMocha }

func (MochaParser) Detect(sample []byte) bool {
	s := string(sample)
	return strings.Contains(s, "✖") || strings.Contains(s, "✔") || strings.Contains(strings.ToLower(s), "failing")
}

func (MochaParser) DetectConfidence(sample []byte) int {
	s := string(sample)
	ls := strings.ToLower(s)
	score := 0
	if strings.Contains(ls, "failing") || strings.Contains(ls, "passing") {
		score += 60
	}
	if strings.Contains(s, "✔") || strings.Contains(s, "✖") {
		score += 30
	}
	if strings.Contains(s, ") ") {
		score += 10
	}
	if score > 100 {
		return 100
	}
	return score
}

func (MochaParser) Parse(r io.Reader) ([]TestResult, error) {
	glyphResults := make([]TestResult, 0)
	numberResults := make([]TestResult, 0)
	suiteNames := map[string]struct{}{}
	s := bufio.NewScanner(r)
	for s.Scan() {
		raw := s.Text()
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if isMochaSuiteLine(raw, line) {
			suiteNames[line] = struct{}{}
		}
		if m := mochaGlyphRE.FindStringSubmatch(line); len(m) == 3 {
			status, ok := mapMochaGlyph(m[1])
			if ok {
				glyphResults = append(glyphResults, TestResult{ID: TestID{Name: m[2]}, Status: status})
			}
			continue
		}
		if m := mochaNumberRE.FindStringSubmatch(line); len(m) == 2 {
			name := strings.TrimSpace(m[1])
			if _, isSuiteHeader := suiteNames[name]; isSuiteHeader {
				continue
			}
			numberResults = append(numberResults, TestResult{ID: TestID{Name: name}, Status: StatusFail})
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	results := glyphResults
	if len(glyphResults) > 0 {
		seen := make(map[string]struct{}, len(glyphResults))
		for _, tr := range glyphResults {
			seen[tr.ID.Canonical()] = struct{}{}
		}
		for _, tr := range numberResults {
			if _, ok := seen[tr.ID.Canonical()]; ok {
				continue
			}
			results = append(results, tr)
		}
	} else {
		// Numeric-only fallback handles formats that do not emit glyph lines.
		results = numberResults
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no mocha/ava results detected")
	}
	return results, nil
}

func isMochaSuiteLine(raw string, trimmed string) bool {
	if !strings.HasPrefix(raw, "  ") {
		return false
	}
	if mochaGlyphRE.MatchString(trimmed) || mochaNumberRE.MatchString(trimmed) {
		return false
	}
	lt := strings.ToLower(trimmed)
	if strings.HasSuffix(lt, " passing") || strings.HasSuffix(lt, " failing") {
		return false
	}
	if strings.HasPrefix(trimmed, "at ") || strings.Contains(trimmed, "AssertionError") {
		return false
	}
	return true
}

func mapMochaGlyph(g string) (TestStatus, bool) {
	switch strings.TrimSpace(g) {
	case "✓", "✔":
		return StatusPass, true
	case "✖", "x":
		return StatusFail, true
	case "-":
		return StatusSkip, true
	default:
		return "", false
	}
}
