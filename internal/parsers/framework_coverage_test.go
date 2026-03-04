package parsers

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type coverageEntry struct {
	Language  string
	Framework Framework
}

// Keep this matrix in sync with README's "Framework Keys" list.
// If you add/remove parser support, update BOTH this list and README.md.
var frameworkCoverageMatrix = []coverageEntry{
	{Language: "Go", Framework: FrameworkGo},
	{Language: "Python", Framework: FrameworkPytest},
	{Language: "JavaScript/TypeScript", Framework: FrameworkJest},
	{Language: "Java/Kotlin/Scala (XML emit)", Framework: FrameworkJunitXML},
	{Language: "Multi-language (TAP emit)", Framework: FrameworkTAP},
	{Language: "Rust", Framework: FrameworkCargo},
	{Language: ".NET", Framework: FrameworkTRX},
	{Language: "Java (Maven Surefire text)", Framework: FrameworkSurefire},
	{Language: "Java (Gradle text)", Framework: FrameworkGradle},
	{Language: ".NET (NUnit XML emit)", Framework: FrameworkNUnitXML},
	{Language: "JavaScript/TypeScript", Framework: FrameworkMocha},
}

func TestFrameworkCoverageMatrixMatchesRegisteredFrameworks(t *testing.T) {
	r := NewRegistry()
	got := slices.Clone(r.SupportedFrameworks())
	slices.Sort(got)

	want := make([]Framework, 0, len(frameworkCoverageMatrix))
	for _, row := range frameworkCoverageMatrix {
		want = append(want, row.Framework)
	}
	slices.Sort(want)

	if len(got) != len(want) {
		t.Fatalf("framework count mismatch: got=%d want=%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("framework mismatch at %d: got=%q want=%q", i, got[i], want[i])
		}
	}
}

func TestReadmeFrameworkKeysListIncludesMatrixRows(t *testing.T) {
	readmePath := filepath.Join("..", "..", "README.md")
	b, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	s := string(b)
	for _, row := range frameworkCoverageMatrix {
		expectedToken := "`" + string(row.Framework) + "`"
		if !strings.Contains(s, expectedToken) {
			t.Fatalf("README missing framework key token %q", expectedToken)
		}
	}
}
