package parsers

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type junitEnvelope struct {
	XMLName    xml.Name         `xml:""`
	Name       string           `xml:"name,attr"`
	Testcases  []junitTestcase  `xml:"testcase"`
	Suites     []junitTestsuite `xml:"testsuite"`
	Testsuites []junitTestsuite `xml:"testsuites>testsuite"`
}

type junitTestsuite struct {
	Name      string          `xml:"name,attr"`
	Testcases []junitTestcase `xml:"testcase"`
}

type junitTestcase struct {
	Name      string         `xml:"name,attr"`
	ClassName string         `xml:"classname,attr"`
	File      string         `xml:"file,attr"`
	Failure   *struct{}      `xml:"failure"`
	Error     *struct{}      `xml:"error"`
	Skipped   *junitSkipElem `xml:"skipped"`
}

type junitSkipElem struct{}

type JunitXMLParser struct{}

func init() {
	registerBuiltin(JunitXMLParser{}, registerOptions{
		Priority: 10,
		Aliases:  []string{"junit", "rspec", "minitest", "phpunit", "pest", "exunit", "gtest", "ctest", "catch2", "xctest", "dart", "flutter", "clojure", "hspec", "webdriverio", "selenium", "testcafe", "cucumber", "robot", "spock", "appium", "karate", "behave", "scalatest", "specs2", "jbehave", "newman", "appium-cucumber"},
	})
}

func (JunitXMLParser) Name() Framework { return FrameworkJunitXML }

func (JunitXMLParser) Detect(sample []byte) bool {
	s := strings.ToLower(string(sample))
	return strings.Contains(s, "<testsuite") || strings.Contains(s, "<testsuites")
}

func (JunitXMLParser) DetectConfidence(sample []byte) int {
	s := strings.ToLower(string(sample))
	score := 0
	if strings.Contains(s, "<testsuite") || strings.Contains(s, "<testsuites") {
		score += 80
	}
	if strings.Contains(s, "<testcase") {
		score += 20
	}
	return score
}

func (JunitXMLParser) Parse(r io.Reader) ([]TestResult, error) {
	b, _, err := readAllAndRestore(r)
	if err != nil {
		return nil, err
	}
	var env junitEnvelope
	if err := xml.Unmarshal(b, &env); err != nil {
		return nil, fmt.Errorf("invalid junit xml: %w", err)
	}
	suites := make([]junitTestsuite, 0, len(env.Suites)+len(env.Testsuites))
	if strings.EqualFold(env.XMLName.Local, "testsuite") && len(env.Testcases) > 0 {
		suites = append(suites, junitTestsuite{Name: env.Name, Testcases: env.Testcases})
	}
	suites = append(suites, env.Suites...)
	suites = append(suites, env.Testsuites...)

	results := make([]TestResult, 0)
	for _, suite := range suites {
		for _, tc := range suite.Testcases {
			if tc.Name == "" {
				continue
			}
			status := StatusPass
			switch {
			case tc.Skipped != nil:
				status = StatusSkip
			case tc.Failure != nil || tc.Error != nil:
				status = StatusFail
			}
			var suiteName *string
			bestSuite := strings.TrimSpace(tc.ClassName)
			if bestSuite == "" {
				bestSuite = strings.TrimSpace(suite.Name)
			}
			if bestSuite != "" {
				suiteName = &bestSuite
			}
			var file *string
			if f := strings.TrimSpace(tc.File); f != "" {
				file = &f
			}
			results = append(results, TestResult{
				ID:     TestID{Name: tc.Name, Suite: suiteName, File: file},
				Status: status,
				Meta:   map[string]string{"format": "junitxml"},
			})
		}
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no junit testcases found")
	}
	return results, nil
}
