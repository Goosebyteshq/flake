# Milestones

## M1 Domain Core

- Pure classification logic
- Sliding window mutation rules
- Transition detection + stable ordering
- Table tests for boundaries and recovering override

## M2 State Engine

- Schema-defined state model
- Load/save with validation
- Atomic write (temp + rename)
- Roundtrip and schema compatibility tests

## M3 CLI Wiring

- `scan` + `report` command flows
- Stable, concise output format
- JSON output mode
- Exit code conformance tests

## M4 Parsers

- Go, Pytest, Jest, JUnit XML parsers
- Deterministic parser registry
- Conformance harness + framework fixtures

## M5 Notify

- Slack notifier abstraction + adapter
- Transition filtering and batching
- Suppression using `last_notified_run_id`
- Slack failures non-fatal to CI

## M6 RunMeta + Events

- Deterministic run metadata normalization
- Versioned events builder and output
- Golden tests for payload stability

## M7 Hardening

- 10k+ test performance checks
- Memory sanity checks
- Error message quality pass
- Cross-platform CLI build verification
