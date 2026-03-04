package parsers

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type trxRoot struct {
	XMLName     xml.Name        `xml:"TestRun"`
	ResultsList []trxUnitResult `xml:"Results>UnitTestResult"`
}

type trxUnitResult struct {
	TestName string `xml:"testName,attr"`
	Outcome  string `xml:"outcome,attr"`
}

type TRXParser struct{}

func init() {
	registerBuiltin(TRXParser{}, registerOptions{
		Priority: 20,
		Aliases:  []string{"dotnet", "vstest", "xunit", "mstest", "specflow", "reqnroll"},
	})
}

func (TRXParser) Name() Framework { return FrameworkTRX }

func (TRXParser) Detect(sample []byte) bool {
	s := strings.ToLower(string(sample))
	return strings.Contains(s, "<testrun") && strings.Contains(s, "unittestresult")
}

func (TRXParser) DetectConfidence(sample []byte) int {
	s := strings.ToLower(string(sample))
	score := 0
	if strings.Contains(s, "<testrun") {
		score += 60
	}
	if strings.Contains(s, "unittestresult") {
		score += 40
	}
	return score
}

func (TRXParser) Parse(r io.Reader) ([]TestResult, error) {
	b, _, err := readAllAndRestore(r)
	if err != nil {
		return nil, err
	}
	var root trxRoot
	if err := xml.Unmarshal(b, &root); err != nil {
		return nil, fmt.Errorf("invalid trx xml: %w", err)
	}
	results := make([]TestResult, 0, len(root.ResultsList))
	for _, ut := range root.ResultsList {
		name := strings.TrimSpace(ut.TestName)
		if name == "" {
			continue
		}
		status, ok := mapTRXOutcome(ut.Outcome)
		if !ok {
			continue
		}
		results = append(results, TestResult{ID: TestID{Name: name}, Status: status})
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no trx test results detected")
	}
	return results, nil
}
