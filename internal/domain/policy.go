package domain

import "fmt"

// ClassificationPolicy controls deterministic class boundaries.
type ClassificationPolicy struct {
	NewFailSampleMax int     `json:"newfail_sample_max"`
	FlakyMinFailures int     `json:"flaky_min_failures"`
	FlakyMinRate     float64 `json:"flaky_min_rate"`
	FlakyMaxRate     float64 `json:"flaky_max_rate"`
	BrokenMinRate    float64 `json:"broken_min_rate"`

	EnableRecovering  bool    `json:"enable_recovering"`
	RecoveringMinDrop float64 `json:"recovering_min_drop"`
}

func DefaultClassificationPolicy() ClassificationPolicy {
	return ClassificationPolicy{
		NewFailSampleMax:  3,
		FlakyMinFailures:  3,
		FlakyMinRate:      0.05,
		FlakyMaxRate:      0.50,
		BrokenMinRate:     0.80,
		EnableRecovering:  true,
		RecoveringMinDrop: 0.20,
	}
}

func (p ClassificationPolicy) Validate() error {
	if p.NewFailSampleMax <= 0 {
		return fmt.Errorf("policy.newfail_sample_max must be > 0")
	}
	if p.FlakyMinFailures <= 0 {
		return fmt.Errorf("policy.flaky_min_failures must be > 0")
	}
	if p.FlakyMinRate < 0 || p.FlakyMinRate > 1 {
		return fmt.Errorf("policy.flaky_min_rate must be in [0,1]")
	}
	if p.FlakyMaxRate < 0 || p.FlakyMaxRate > 1 {
		return fmt.Errorf("policy.flaky_max_rate must be in [0,1]")
	}
	if p.FlakyMinRate > p.FlakyMaxRate {
		return fmt.Errorf("policy.flaky_min_rate must be <= policy.flaky_max_rate")
	}
	if p.BrokenMinRate < 0 || p.BrokenMinRate > 1 {
		return fmt.Errorf("policy.broken_min_rate must be in [0,1]")
	}
	if p.RecoveringMinDrop < 0 || p.RecoveringMinDrop > 1 {
		return fmt.Errorf("policy.recovering_min_drop must be in [0,1]")
	}
	return nil
}
