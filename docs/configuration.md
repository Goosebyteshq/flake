# Configuration

`flake` reads configuration from YAML.

Default path behavior:

- If `--config <path>` is provided, that file is used.
- If `--config` is omitted and `.flake-config.yaml` exists, it is auto-loaded.
- If no config file is present, built-in defaults are used.

## File Format

```yaml
window: 50
policy_version: 1

policy:
  newfail_sample_max: 3
  flaky_min_failures: 3
  flaky_min_rate: 0.05
  flaky_max_rate: 0.50
  broken_min_rate: 0.80
  recovering:
    enabled: true
    min_drop: 0.20

notify:
  on_transition: true
  min_failure_rate: 0.05
  include_classes:
    - Broken
    - Flaky
    - NewFail
    - Recovering
  suppress_repeats_for_runs: 5
  min_transition_age_runs: 0
  oscillation_window_runs: 3
  batch: true
  max_items_per_message: 50

slack:
  webhook: ""
  timeout_seconds: 5

run_meta:
  prefer_env: true
  allow_file_override: true
```

## Validation Rules

- `window` must be greater than `0`
- `policy_version` must be greater than `0`
- `policy.flaky_min_rate <= policy.flaky_max_rate`
- policy rates and drop values must be in `[0,1]`
- `notify.max_items_per_message` falls back to `50` if `<= 0`
- `slack.timeout_seconds` falls back to `5` if `<= 0`

## Notify Noise Controls

- `notify.suppress_repeats_for_runs`: suppress repeated notifications of the same class for N runs.
- `notify.min_transition_age_runs`: suppress notifications for very new tests/transitions.
- `notify.oscillation_window_runs`: suppress rapid class ping-pong back to a recent class.

## Report Views

`flake report` supports focused views:

- `default`: all tests by severity/failure-rate
- `unstable`: classes `Broken`, `Flaky`, `NewFail`
- `recovered`: class `Recovering`
- `long-failing`: tests with trailing fail streaks, sorted by streak then rate
  - default minimum trailing streak is `3`
  - override per command with `flake report --view long-failing --min-fail-streak <n>`

## Security

- Never commit real webhook values.
- The webhook value is never printed in logs.
