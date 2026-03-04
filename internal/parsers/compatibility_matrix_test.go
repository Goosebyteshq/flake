package parsers

import (
	"strings"
	"testing"
)

type compatCase struct {
	Name          string
	Alias         string
	Fixture       string
	WantFramework Framework
	WantPass      int
	WantFail      int
	WantSkip      int
}

// Keep this compatibility matrix in sync with README "Compatibility Coverage"
// when adding or removing framework compatibility fixtures.
var compatMatrix = []compatCase{
	{
		Name:          "vitest",
		Alias:         "vitest",
		Fixture:       "compat/vitest/sample.txt",
		WantFramework: FrameworkJest,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "playwright",
		Alias:         "playwright",
		Fixture:       "compat/playwright/sample.txt",
		WantFramework: FrameworkJest,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "jasmine",
		Alias:         "jasmine",
		Fixture:       "compat/jasmine/sample.txt",
		WantFramework: FrameworkJest,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "cypress",
		Alias:         "cypress",
		Fixture:       "compat/cypress/sample.txt",
		WantFramework: FrameworkMocha,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      0,
	},
	{
		Name:          "unittest",
		Alias:         "unittest",
		Fixture:       "compat/unittest/sample.txt",
		WantFramework: FrameworkPytest,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "nose2",
		Alias:         "nose2",
		Fixture:       "compat/nose2/sample.txt",
		WantFramework: FrameworkPytest,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "tox",
		Alias:         "tox",
		Fixture:       "compat/tox/sample.txt",
		WantFramework: FrameworkPytest,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "junit5",
		Alias:         "junit5",
		Fixture:       "compat/junit5/sample.txt",
		WantFramework: FrameworkJest,
		WantPass:      2,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "testng",
		Alias:         "testng",
		Fixture:       "compat/testng/sample.txt",
		WantFramework: FrameworkSurefire,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "kotest",
		Alias:         "kotest",
		Fixture:       "compat/kotest/sample.txt",
		WantFramework: FrameworkGradle,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "sbt",
		Alias:         "sbt",
		Fixture:       "compat/sbt/sample.txt",
		WantFramework: FrameworkGradle,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "xunit",
		Alias:         "xunit",
		Fixture:       "compat/xunit/sample.xml",
		WantFramework: FrameworkTRX,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "nunit-console",
		Alias:         "nunit-console",
		Fixture:       "compat/nunit-console/sample.xml",
		WantFramework: FrameworkNUnitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "mstest",
		Alias:         "mstest",
		Fixture:       "compat/mstest/sample.xml",
		WantFramework: FrameworkTRX,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "rspec",
		Alias:         "rspec",
		Fixture:       "compat/rspec/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "minitest",
		Alias:         "minitest",
		Fixture:       "compat/minitest/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "phpunit",
		Alias:         "phpunit",
		Fixture:       "compat/phpunit/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "pest",
		Alias:         "pest",
		Fixture:       "compat/pest/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "nextest",
		Alias:         "nextest",
		Fixture:       "compat/nextest/sample.txt",
		WantFramework: FrameworkCargo,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "exunit",
		Alias:         "exunit",
		Fixture:       "compat/exunit/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "gtest",
		Alias:         "gtest",
		Fixture:       "compat/gtest/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "ctest",
		Alias:         "ctest",
		Fixture:       "compat/ctest/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "catch2",
		Alias:         "catch2",
		Fixture:       "compat/catch2/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "xctest",
		Alias:         "xctest",
		Fixture:       "compat/xctest/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "dart",
		Alias:         "dart",
		Fixture:       "compat/dart/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "flutter",
		Alias:         "flutter",
		Fixture:       "compat/flutter/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "clojure",
		Alias:         "clojure",
		Fixture:       "compat/clojure/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "hspec",
		Alias:         "hspec",
		Fixture:       "compat/hspec/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "prove",
		Alias:         "prove",
		Fixture:       "compat/prove/sample.txt",
		WantFramework: FrameworkTAP,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "bats",
		Alias:         "bats",
		Fixture:       "compat/bats/sample.txt",
		WantFramework: FrameworkTAP,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "webdriverio",
		Alias:         "webdriverio",
		Fixture:       "compat/webdriverio/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "selenium",
		Alias:         "selenium",
		Fixture:       "compat/selenium/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "testcafe",
		Alias:         "testcafe",
		Fixture:       "compat/testcafe/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "node-test",
		Alias:         "node-test",
		Fixture:       "compat/node-test/sample.txt",
		WantFramework: FrameworkTAP,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "bun-test",
		Alias:         "bun-test",
		Fixture:       "compat/bun-test/sample.txt",
		WantFramework: FrameworkJest,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "deno-test",
		Alias:         "deno-test",
		Fixture:       "compat/deno-test/sample.txt",
		WantFramework: FrameworkTAP,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "cucumber",
		Alias:         "cucumber",
		Fixture:       "compat/cucumber/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "robot",
		Alias:         "robot",
		Fixture:       "compat/robot/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "spock",
		Alias:         "spock",
		Fixture:       "compat/spock/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "failsafe",
		Alias:         "failsafe",
		Fixture:       "compat/failsafe/sample.txt",
		WantFramework: FrameworkSurefire,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "appium",
		Alias:         "appium",
		Fixture:       "compat/appium/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "karate",
		Alias:         "karate",
		Fixture:       "compat/karate/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "specflow",
		Alias:         "specflow",
		Fixture:       "compat/specflow/sample.xml",
		WantFramework: FrameworkTRX,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "behave",
		Alias:         "behave",
		Fixture:       "compat/behave/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "scalatest",
		Alias:         "scalatest",
		Fixture:       "compat/scalatest/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "specs2",
		Alias:         "specs2",
		Fixture:       "compat/specs2/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "jbehave",
		Alias:         "jbehave",
		Fixture:       "compat/jbehave/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "pester",
		Alias:         "pester",
		Fixture:       "compat/pester/sample.xml",
		WantFramework: FrameworkNUnitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "newman",
		Alias:         "newman",
		Fixture:       "compat/newman/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
	{
		Name:          "appium-cucumber",
		Alias:         "appium-cucumber",
		Fixture:       "compat/appium-cucumber/sample.xml",
		WantFramework: FrameworkJunitXML,
		WantPass:      1,
		WantFail:      1,
		WantSkip:      1,
	},
}

func TestCompatibilityAliasesResolve(t *testing.T) {
	for _, tc := range compatMatrix {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := ParseFramework(tc.Alias)
			if err != nil {
				t.Fatalf("ParseFramework(%q) error: %v", tc.Alias, err)
			}
			if got != tc.WantFramework {
				t.Fatalf("ParseFramework(%q)=%q, want %q", tc.Alias, got, tc.WantFramework)
			}
		})
	}
}

func TestCompatibilityFixturesParse(t *testing.T) {
	r := NewRegistry()
	for _, tc := range compatMatrix {
		t.Run(tc.Name, func(t *testing.T) {
			in := mustReadFixture(t, tc.Fixture)
			detected, results, err := r.Parse(FrameworkAuto, strings.NewReader(string(in)))
			if err != nil {
				t.Fatalf("Parse auto error: %v", err)
			}
			if detected != tc.WantFramework {
				t.Fatalf("detected=%q, want %q", detected, tc.WantFramework)
			}

			var pass, fail, skip int
			for _, tr := range results {
				switch tr.Status {
				case StatusPass:
					pass++
				case StatusFail:
					fail++
				case StatusSkip:
					skip++
				}
			}
			if pass != tc.WantPass || fail != tc.WantFail || skip != tc.WantSkip {
				t.Fatalf("status counts pass/fail/skip=%d/%d/%d, want %d/%d/%d",
					pass, fail, skip, tc.WantPass, tc.WantFail, tc.WantSkip)
			}
		})
	}
}
