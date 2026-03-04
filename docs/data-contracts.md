# Data Contracts (v1)

## Test Identity

Canonical identity is stable and framework-aware:

- Go/Pytest/Jest: `test_id = Name`
- JUnit XML: `test_id = Suite + "::" + Name` (when suite exists)

## Status Domain

Only normalized statuses are valid:

- `pass`
- `fail`
- `skip`

## State File

Local file: `.flake-state.json`

Top-level contract:

- `schema_version` (int)
- `policy_version` (int)
- `window` (int)
- `updated_at` (RFC3339)
- `tests` map keyed by canonical `test_id`
- `last_run` containing `run_id`, `timestamp`, `framework`, `run_meta`, `transitions`

Per-test contract includes:

- `history` sliding window (`P`/`F` string)
- `first_seen`, `last_seen`, `last_failed_at`, `last_passed_at`
- `last_state` (`class`, `failure_rate`)
- `last_notified_run_id` (optional)

## Events Payload

Emitted by `flake scan --events`:

- Versioned and stable ordered JSON payload
- Contains run metadata, derived per-test state, and transitions
- Designed so future `flake publish` can upload directly without schema redesign

## Scan JSON Output

Emitted by `flake scan --json`:

- `policy`: effective classification policy snapshot used for that run
- `classifications`: deterministic per-test details sorted by `test_id`
  - `test_id`, `history`
  - `derived`: `class`, `failure_rate`, `failures`, `passes`, `sample_size`
  - `explanation`: `base_class`, `recovering_applied`, `reasons`
  - optional previous fields when available: `previous_class`, `previous_failure_rate`
