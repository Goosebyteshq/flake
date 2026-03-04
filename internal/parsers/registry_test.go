package parsers

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseFrameworkAliases(t *testing.T) {
	cases := []struct {
		in   string
		want Framework
	}{
		{"auto", FrameworkAuto},
		{"go", FrameworkGo},
		{"gotest", FrameworkGo},
		{"pytest", FrameworkPytest},
		{"jest", FrameworkJest},
		{"vitest", FrameworkJest},
		{"playwright", FrameworkJest},
		{"jasmine", FrameworkJest},
		{"junit5", FrameworkJest},
		{"bun", FrameworkJest},
		{"bun-test", FrameworkJest},
		{"junitxml", FrameworkJunitXML},
		{"junit", FrameworkJunitXML},
		{"gtest", FrameworkJunitXML},
		{"ctest", FrameworkJunitXML},
		{"catch2", FrameworkJunitXML},
		{"xctest", FrameworkJunitXML},
		{"dart", FrameworkJunitXML},
		{"flutter", FrameworkJunitXML},
		{"clojure", FrameworkJunitXML},
		{"hspec", FrameworkJunitXML},
		{"webdriverio", FrameworkJunitXML},
		{"selenium", FrameworkJunitXML},
		{"testcafe", FrameworkJunitXML},
		{"cucumber", FrameworkJunitXML},
		{"robot", FrameworkJunitXML},
		{"spock", FrameworkJunitXML},
		{"appium", FrameworkJunitXML},
		{"karate", FrameworkJunitXML},
		{"behave", FrameworkJunitXML},
		{"scalatest", FrameworkJunitXML},
		{"specs2", FrameworkJunitXML},
		{"jbehave", FrameworkJunitXML},
		{"newman", FrameworkJunitXML},
		{"appium-cucumber", FrameworkJunitXML},
		{"tap", FrameworkTAP},
		{"prove", FrameworkTAP},
		{"bats", FrameworkTAP},
		{"perl", FrameworkTAP},
		{"node-test", FrameworkTAP},
		{"node-test-runner", FrameworkTAP},
		{"deno", FrameworkTAP},
		{"deno-test", FrameworkTAP},
		{"cargo", FrameworkCargo},
		{"rust", FrameworkCargo},
		{"trx", FrameworkTRX},
		{"dotnet", FrameworkTRX},
		{"xunit", FrameworkTRX},
		{"mstest", FrameworkTRX},
		{"specflow", FrameworkTRX},
		{"reqnroll", FrameworkTRX},
		{"surefire", FrameworkSurefire},
		{"testng", FrameworkSurefire},
		{"failsafe", FrameworkSurefire},
		{"maven-failsafe", FrameworkSurefire},
		{"gradle", FrameworkGradle},
		{"kotest", FrameworkGradle},
		{"sbt", FrameworkGradle},
		{"nunitxml", FrameworkNUnitXML},
		{"nunit", FrameworkNUnitXML},
		{"nunit-console", FrameworkNUnitXML},
		{"pester", FrameworkNUnitXML},
		{"mocha", FrameworkMocha},
		{"ava", FrameworkMocha},
		{"cypress", FrameworkMocha},
		{"unittest", FrameworkPytest},
		{"nose2", FrameworkPytest},
		{"tox", FrameworkPytest},
		{"rspec", FrameworkJunitXML},
		{"minitest", FrameworkJunitXML},
		{"phpunit", FrameworkJunitXML},
		{"pest", FrameworkJunitXML},
		{"exunit", FrameworkJunitXML},
		{"nextest", FrameworkCargo},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseFramework(tc.in)
			if err != nil {
				t.Fatalf("ParseFramework(%q) error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("ParseFramework(%q)=%q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestRegistrySupportedFrameworksDeterministic(t *testing.T) {
	r := NewRegistry()
	got1 := r.SupportedFrameworks()
	got2 := r.SupportedFrameworks()
	if !reflect.DeepEqual(got1, got2) {
		t.Fatalf("SupportedFrameworks is not deterministic: %v vs %v", got1, got2)
	}
	if len(got1) == 0 {
		t.Fatalf("SupportedFrameworks must not be empty")
	}
	seen := map[Framework]bool{}
	for _, f := range got1 {
		if seen[f] {
			t.Fatalf("duplicate framework in SupportedFrameworks: %q", f)
		}
		seen[f] = true
	}
	required := []Framework{FrameworkGo, FrameworkPytest, FrameworkJest, FrameworkJunitXML}
	for _, f := range required {
		if !seen[f] {
			t.Fatalf("missing required framework %q", f)
		}
	}
}

func TestRegistryAutoDetect(t *testing.T) {
	r := NewRegistry()
	for _, framework := range r.SupportedFrameworks() {
		framework := framework
		t.Run(string(framework), func(t *testing.T) {
			b := mustReadFrameworkFixture(t, framework)
			got, err := r.Detect(FrameworkAuto, strings.NewReader(string(b)))
			if err != nil {
				t.Fatalf("Detect error: %v", err)
			}
			if got != framework {
				t.Fatalf("Detect=%q, want %q", got, framework)
			}
		})
	}
}

func TestRegistryParseConformance(t *testing.T) {
	r := NewRegistry()
	for _, framework := range r.SupportedFrameworks() {
		framework := framework
		t.Run(string(framework), func(t *testing.T) {
			b := mustReadFrameworkFixture(t, framework)
			detected, results, err := r.Parse(FrameworkAuto, strings.NewReader(string(b)))
			if err != nil {
				t.Fatalf("Parse auto error: %v", err)
			}
			if detected != framework {
				t.Fatalf("detected %q, want %q", detected, framework)
			}
			if len(results) == 0 {
				t.Fatalf("expected non-empty results")
			}
			for _, tr := range results {
				if tr.ID.Canonical() == "" {
					t.Fatalf("empty canonical test id")
				}
				if err := tr.Validate(); err != nil {
					t.Fatalf("invalid test result: %v", err)
				}
			}
		})
	}
}

func TestRegistryExplicitFramework(t *testing.T) {
	r := NewRegistry()
	b := mustReadFixture(t, "go/sample.txt")
	detected, results, err := r.Parse(FrameworkGo, strings.NewReader(string(b)))
	if err != nil {
		t.Fatalf("Parse explicit framework error: %v", err)
	}
	if detected != FrameworkGo {
		t.Fatalf("detected=%q, want %q", detected, FrameworkGo)
	}
	if len(results) != 3 {
		t.Fatalf("len(results)=%d, want 3", len(results))
	}
}

func TestRegistryMixedLogMergesSuccessfulParses(t *testing.T) {
	r := NewRegistry()
	b := mustReadFixture(t, "mixed/go_pytest.txt")
	detected, results, err := r.Parse(FrameworkAuto, strings.NewReader(string(b)))
	if err != nil {
		t.Fatalf("Parse mixed error: %v", err)
	}
	if detected == "" {
		t.Fatalf("expected detected framework")
	}
	if len(results) < 4 {
		t.Fatalf("expected merged result set >=4, got %d", len(results))
	}
	gotIDs := map[string]bool{}
	for _, tr := range results {
		gotIDs[tr.ID.Canonical()] = true
	}
	for _, want := range []string{"TestGoA", "TestGoB", "tests/test_math.py::test_add", "tests/test_math.py::test_sub"} {
		if !gotIDs[want] {
			t.Fatalf("missing merged test id %q", want)
		}
	}
}

func TestExplainCandidatesConfidenceOrdering(t *testing.T) {
	r := NewRegistry()
	sample := []byte("1 failing\n✕ edge case\n")
	candidates, err := r.ExplainCandidates(FrameworkAuto, "", sample)
	if err != nil {
		t.Fatalf("ExplainCandidates error: %v", err)
	}
	if len(candidates) == 0 {
		t.Fatalf("expected candidates")
	}
	// Mocha should outrank Jest for this sample due to explicit failing-summary signal.
	mochaIdx := -1
	jestIdx := -1
	for i, c := range candidates {
		if c.Framework == FrameworkMocha {
			mochaIdx = i
		}
		if c.Framework == FrameworkJest {
			jestIdx = i
		}
	}
	if mochaIdx == -1 || jestIdx == -1 {
		t.Fatalf("expected both mocha and jest candidates, got %+v", candidates)
	}
	if mochaIdx > jestIdx {
		t.Fatalf("expected mocha candidate before jest, got %+v", candidates)
	}
}
