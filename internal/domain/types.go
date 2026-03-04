package domain

import "fmt"

type TestStatus string

const (
	Pass TestStatus = "pass"
	Fail TestStatus = "fail"
	Skip TestStatus = "skip"
)

type Class string

const (
	ClassHealthy    Class = "Healthy"
	ClassNewFail    Class = "NewFail"
	ClassFlaky      Class = "Flaky"
	ClassBroken     Class = "Broken"
	ClassRecovering Class = "Recovering"
)

type DerivedState struct {
	Class       Class   `json:"class"`
	FailureRate float64 `json:"failure_rate"`
	Failures    int     `json:"failures"`
	Passes      int     `json:"passes"`
	SampleSize  int     `json:"sample_size"`
}

type Transition struct {
	TestID      string  `json:"test_id"`
	From        Class   `json:"from"`
	To          Class   `json:"to"`
	FailureRate float64 `json:"failure_rate"`
	Severity    int     `json:"severity"`
}

func (s TestStatus) Validate() error {
	switch s {
	case Pass, Fail, Skip:
		return nil
	default:
		return fmt.Errorf("invalid status %q", s)
	}
}
