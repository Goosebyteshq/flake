package parsers

import (
	"fmt"
	"sort"
)

// TestStatus is the normalized test outcome used across all parser outputs.
type TestStatus string

const (
	StatusPass TestStatus = "pass"
	StatusFail TestStatus = "fail"
	StatusSkip TestStatus = "skip"
)

// TestID holds parser-normalized identity fields.
type TestID struct {
	Name  string
	File  *string
	Suite *string
}

// Canonical returns the v1 canonical test id string.
func (id TestID) Canonical() string {
	if id.Suite != nil && *id.Suite != "" {
		return *id.Suite + "::" + id.Name
	}
	return id.Name
}

func (id TestID) Validate() error {
	if id.Name == "" {
		return fmt.Errorf("test name is required")
	}
	return nil
}

// TestResult is a normalized parsed result from any framework.
type TestResult struct {
	ID     TestID
	Status TestStatus
	Meta   map[string]string
}

func (r TestResult) Validate() error {
	if err := r.ID.Validate(); err != nil {
		return err
	}
	switch r.Status {
	case StatusPass, StatusFail, StatusSkip:
		return nil
	default:
		return fmt.Errorf("invalid status %q", r.Status)
	}
}

// StableSortResults applies deterministic ordering.
func StableSortResults(in []TestResult) {
	sort.SliceStable(in, func(i, j int) bool {
		a := in[i]
		b := in[j]
		if a.ID.Canonical() != b.ID.Canonical() {
			return a.ID.Canonical() < b.ID.Canonical()
		}
		return a.Status < b.Status
	})
}
