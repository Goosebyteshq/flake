package parsers

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type nunitRoot struct {
	XMLName  xml.Name    `xml:"test-run"`
	TestCase []nunitCase `xml:"test-suite>results>test-case"`
	AllCases []nunitCase `xml:"test-case"`
}

type nunitCase struct {
	FullName string `xml:"fullname,attr"`
	Result   string `xml:"result,attr"`
}

type NUnitXMLParser struct{}

func init() {
	registerBuiltin(NUnitXMLParser{}, registerOptions{Priority: 22, Aliases: []string{"nunit", "nunit-console", "pester"}})
}

func (NUnitXMLParser) Name() Framework { return FrameworkNUnitXML }

func (NUnitXMLParser) Detect(sample []byte) bool {
	s := strings.ToLower(string(sample))
	return strings.Contains(s, "<test-run") && strings.Contains(s, "test-case")
}

func (NUnitXMLParser) DetectConfidence(sample []byte) int {
	s := strings.ToLower(string(sample))
	score := 0
	if strings.Contains(s, "<test-run") {
		score += 60
	}
	if strings.Contains(s, "test-case") {
		score += 40
	}
	return score
}

func (NUnitXMLParser) Parse(r io.Reader) ([]TestResult, error) {
	b, _, err := readAllAndRestore(r)
	if err != nil {
		return nil, err
	}
	var root nunitRoot
	if err := xml.Unmarshal(b, &root); err != nil {
		return nil, fmt.Errorf("invalid nunit xml: %w", err)
	}
	cases := root.TestCase
	if len(cases) == 0 {
		cases = root.AllCases
	}
	results := make([]TestResult, 0, len(cases))
	for _, c := range cases {
		name := strings.TrimSpace(c.FullName)
		if name == "" {
			continue
		}
		status, ok := mapNUnitResult(c.Result)
		if !ok {
			continue
		}
		results = append(results, TestResult{ID: TestID{Name: name}, Status: status})
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no nunit test results detected")
	}
	return results, nil
}

func mapNUnitResult(v string) (TestStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "passed", "success":
		return StatusPass, true
	case "failed", "error":
		return StatusFail, true
	case "skipped", "inconclusive", "ignored":
		return StatusSkip, true
	default:
		return "", false
	}
}
