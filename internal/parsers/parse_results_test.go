package parsers

import (
	"strings"
	"testing"
)

func TestParsersStatusesAndCanonicalOrder(t *testing.T) {
	r := NewRegistry()
	cases := []struct {
		framework Framework
		input     string
		wantOrder []string
	}{
		{
			framework: FrameworkGo,
			input: `--- FAIL: TestB (0.00s)
--- PASS: TestA (0.00s)
--- SKIP: TestC (0.00s)
`,
			wantOrder: []string{"TestA", "TestB", "TestC"},
		},
		{
			framework: FrameworkTAP,
			input: `TAP version 13
ok 2 - b
not ok 1 - a
`,
			wantOrder: []string{"a", "b"},
		},
		{
			framework: FrameworkMocha,
			input: `
  sample
    ✔ pass case
    1) fail case


  1 passing (5ms)
  1 failing

  1) sample
       fail case:
`,
			wantOrder: []string{"fail case", "pass case"},
		},
		{
			framework: FrameworkSurefire,
			input: `[INFO] Running com.example.SampleTest
[ERROR] com.example.SampleTest.testFail -- Time elapsed: 0.002 s <<< FAILURE!
[INFO] com.example.SampleTest.testPass -- Time elapsed: 0.001 s <<< SUCCESS!
`,
			wantOrder: []string{"com.example.SampleTest.testFail", "com.example.SampleTest.testPass"},
		},
	}

	for _, tc := range cases {
		t.Run(string(tc.framework), func(t *testing.T) {
			_, got, err := r.Parse(tc.framework, strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if len(got) != len(tc.wantOrder) {
				t.Fatalf("len(got)=%d, want %d", len(got), len(tc.wantOrder))
			}
			for i := range got {
				if got[i].ID.Canonical() != tc.wantOrder[i] {
					t.Fatalf("order[%d]=%q, want %q", i, got[i].ID.Canonical(), tc.wantOrder[i])
				}
			}
		})
	}
}
