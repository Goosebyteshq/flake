package config

import (
	"fmt"
	"os"

	"github.com/Goosebyteshq/flake/internal/domain"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Window        int           `yaml:"window"`
	PolicyVersion int           `yaml:"policy_version"`
	Policy        PolicyConfig  `yaml:"policy"`
	Notify        NotifyConfig  `yaml:"notify"`
	Slack         SlackConfig   `yaml:"slack"`
	RunMeta       RunMetaConfig `yaml:"run_meta"`
}

type PolicyConfig struct {
	NewFailSampleMax int                    `yaml:"newfail_sample_max"`
	FlakyMinFailures int                    `yaml:"flaky_min_failures"`
	FlakyMinRate     float64                `yaml:"flaky_min_rate"`
	FlakyMaxRate     float64                `yaml:"flaky_max_rate"`
	BrokenMinRate    float64                `yaml:"broken_min_rate"`
	Recovering       RecoveringPolicyConfig `yaml:"recovering"`
}

type RecoveringPolicyConfig struct {
	Enabled bool    `yaml:"enabled"`
	MinDrop float64 `yaml:"min_drop"`
}

type NotifyConfig struct {
	OnTransition           bool     `yaml:"on_transition"`
	MinFailureRate         float64  `yaml:"min_failure_rate"`
	IncludeClasses         []string `yaml:"include_classes"`
	SuppressRepeatsForRuns int      `yaml:"suppress_repeats_for_runs"`
	MinTransitionAgeRuns   int      `yaml:"min_transition_age_runs"`
	OscillationWindowRuns  int      `yaml:"oscillation_window_runs"`
	Batch                  bool     `yaml:"batch"`
	MaxItemsPerMessage     int      `yaml:"max_items_per_message"`
}

type SlackConfig struct {
	Webhook        string `yaml:"webhook"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

type RunMetaConfig struct {
	PreferEnv         bool `yaml:"prefer_env"`
	AllowFileOverride bool `yaml:"allow_file_override"`
}

func Default() Config {
	return Config{
		Window:        50,
		PolicyVersion: 1,
		Policy: func() PolicyConfig {
			p := domain.DefaultClassificationPolicy()
			return PolicyConfig{
				NewFailSampleMax: p.NewFailSampleMax,
				FlakyMinFailures: p.FlakyMinFailures,
				FlakyMinRate:     p.FlakyMinRate,
				FlakyMaxRate:     p.FlakyMaxRate,
				BrokenMinRate:    p.BrokenMinRate,
				Recovering: RecoveringPolicyConfig{
					Enabled: p.EnableRecovering,
					MinDrop: p.RecoveringMinDrop,
				},
			}
		}(),
		Notify: NotifyConfig{
			OnTransition:           true,
			MinFailureRate:         0.05,
			IncludeClasses:         []string{"Broken", "Flaky", "NewFail", "Recovering"},
			SuppressRepeatsForRuns: 5,
			MinTransitionAgeRuns:   0,
			OscillationWindowRuns:  3,
			Batch:                  true,
			MaxItemsPerMessage:     50,
		},
		Slack:   SlackConfig{TimeoutSeconds: 5},
		RunMeta: RunMetaConfig{PreferEnv: true, AllowFileOverride: true},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		if _, err := os.Stat(".flake-config.yaml"); err == nil {
			path = ".flake-config.yaml"
		} else {
			return cfg, nil
		}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("invalid config yaml: %w", err)
	}
	if cfg.Window <= 0 {
		return cfg, fmt.Errorf("window must be > 0")
	}
	if cfg.PolicyVersion <= 0 {
		return cfg, fmt.Errorf("policy_version must be > 0")
	}
	if err := cfg.ClassificationPolicy().Validate(); err != nil {
		return cfg, err
	}
	if cfg.Notify.MaxItemsPerMessage <= 0 {
		cfg.Notify.MaxItemsPerMessage = 50
	}
	if cfg.Slack.TimeoutSeconds <= 0 {
		cfg.Slack.TimeoutSeconds = 5
	}
	return cfg, nil
}

func (c Config) ClassificationPolicy() domain.ClassificationPolicy {
	return domain.ClassificationPolicy{
		NewFailSampleMax:  c.Policy.NewFailSampleMax,
		FlakyMinFailures:  c.Policy.FlakyMinFailures,
		FlakyMinRate:      c.Policy.FlakyMinRate,
		FlakyMaxRate:      c.Policy.FlakyMaxRate,
		BrokenMinRate:     c.Policy.BrokenMinRate,
		EnableRecovering:  c.Policy.Recovering.Enabled,
		RecoveringMinDrop: c.Policy.Recovering.MinDrop,
	}
}
