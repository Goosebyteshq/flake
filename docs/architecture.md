# Architecture

## Principle

The system is split into deterministic core logic and isolated side-effect layers.

## Layers

- `cmd/flake`: command wiring and flag parsing only
- `internal/app`: orchestration of scan/report/notify flows
- `internal/domain`: pure classification, transitions, severity logic
- `internal/engine`: deterministic in-memory state mutation
- `internal/state`: load/save, schema handling, atomic file writes
- `internal/parsers`: framework detection and parse normalization
- `internal/report`: grouped/sorted human and JSON view models
- `internal/notify`: Slack adapter and message batching/filtering
- `internal/events`: publish-ready, versioned events payload builder
- `internal/runmeta`: deterministic run context normalization

## Determinism Requirements

- Stable sorting for all user-visible and persisted lists
- Injected clock for reproducible tests and golden outputs
- No randomness
- No nondeterministic map iteration in JSON/state output

## Extension Points

- Parser registry (`internal/parsers`) supports adding frameworks
- Parser self-registration (`init` + priority + aliases) minimizes central edits
- Parser hint store remembers successful framework per repo to reduce probe churn
- Classification policy is config-driven and versioned (`policy_version` + `policy` block)
- Event builder (`internal/events`) supports future `publish`
- Policy versioning enables future threshold/config evolution
