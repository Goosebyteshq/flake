# CLI Contract (v1)

## Commands

### `flake scan`

- Input: `stdin` or `--input`
- Detects framework unless overridden with `--framework`
- Updates state and computes transitions
- Outputs human summary
- Optional `--json`
  - includes policy snapshot and per-test classification explanations
- Optional `--events <path|->` for publish-ready payload
- Optional `--debug` for deterministic parser diagnostics (hinted parser, detected parser, ranked parser candidates)

### `flake report`

- Reads local state
- Outputs grouped, sorted report
- Optional `--json`
- Optional `--view <default|unstable|recovered|long-failing>`
- Optional `--limit <n>`
- Optional `--min-fail-streak <n>` (used by `long-failing`, default `3`)

### `flake notify`

- Reads transitions from last run
- Applies notify filters
- Sends Slack webhook
- Slack failures never fail CI
- Applies suppression (`suppress_repeats_for_runs`) and deterministic batching

### `flake publish`

- Reads events payload from `--events <path|->`
- Publishes JSON payload to remote endpoint via HTTP POST
- Supports auth token via `--token` or `--token-env`
- Deterministic retry policy via `--retries`
- Optional `--json`
- Sends deterministic `X-Flake-Idempotency-Key` (SHA-256 of payload)
- Sends `X-Flake-Run-ID` when present in payload
- Optional `--debug` for deterministic publish attempt diagnostics

### `flake compact`

- Reads local state from `--state`
- Optional `--drop-untouched-days <n>` to prune stale tests
- Optional `--max-tests <n>` to keep most recently seen tests
- Optional `--dry-run` to show effect without writing
- Optional `--json`
- Optional `--debug` to print prune settings and removed-id traces

## Reserved Commands (Not Implemented in v1)

- none

Implemented:

- `flake publish` (remote HTTP publish of events payload)
- `flake migrate` (local state schema migration)

## Common Flags

- `--state <path>` default `.flake-state.json`
- `--config <path>` default `.flake-config.yaml` when present
- `--window <n>` default `50`
- `--framework <auto|go|pytest|jest|junitxml|tap|cargo|trx|surefire|gradle|nunitxml|mocha>`
- `--json`
- `--events <path|->` (`scan` only)
- `--run-meta <path>`
- `--fail-on <broken|flaky|newfail>` (scan gate; default off)
  - when triggered, `scan` exits `1` and prints deterministic diagnostics to stderr (triggered classes + top tests and reason codes)

## Parser Selection Memory

`scan` persists successful parser selection per repository and reuses it on the next run.

- Hint file: `.flake-parser-hints.json`
- Behavior: try hinted parser first, fall back to other parsers only if hinted parser fails
- Auto mode uses deterministic confidence scoring to rank parser attempts

## Exit Codes

- Success: `0`
- Parse error: non-zero
- State error: non-zero
- Slack failure: `0`
