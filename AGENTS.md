# AGENTS.md

This file defines how contributors and coding agents should execute work in this repository.

## Mission

Deliver a deterministic, production-grade Go CLI that detects and classifies flaky tests, maintains sliding-window history, detects transitions, integrates cleanly in CI, and emits versioned machine-readable outputs.

Scope for v1: CLI-first with production commands `scan`, `report`, `notify`, `publish`, `compact`, and `migrate`.

## Ground Rules

- Deterministic behavior only (stable sorting, injected clock, no randomness)
- Strict layer separation (no business logic in `cmd/flake`)
- Pure domain logic in `internal/domain` and fully table-tested
- Side effects isolated to `internal/state` and `internal/notify`
- Schema and policy versioning required from day one
- Never log secrets (especially Slack webhook)

## Implementation Plan

1. M1 Domain Core
2. M2 State Engine
3. M3 CLI Wiring
4. M4 Parsers
5. M5 Notify
6. M6 RunMeta + Events
7. M7 Hardening

Detailed milestone expectations are documented in [docs/roadmap.md](./docs/roadmap.md).

## Required Repository Structure

- `cmd/flake`
- `internal/app`
- `internal/domain`
- `internal/engine`
- `internal/state`
- `internal/parsers`
- `internal/report`
- `internal/notify`
- `internal/events`
- `internal/runmeta`
- `testdata/{go,pytest,jest,junitxml,tap,cargo,trx,surefire,gradle,nunitxml,mocha,mixed}`
- `docs/`

## Documentation Map

- [Architecture](./docs/architecture.md)
- [CLI Contract](./docs/cli-contract.md)
- [Configuration](./docs/configuration.md)
- [Development Workflow](./docs/development.md)
- [Framework Support](./docs/framework-support.md)
- [Extending Parsers](./docs/extending-parsers.md)
- [Data Contracts](./docs/data-contracts.md)
- [Testing Strategy](./docs/testing-strategy.md)
- [CI Integration](./docs/ci-integration.md)
- [Smoke Tests](./docs/smoke-tests.md)
- [Milestones](./docs/roadmap.md)

## Working Conventions

- Keep outputs stable and sorted.
- Prefer parser self-registration (`init` + `registerBuiltin`) to avoid central wiring edits.
- Preserve parser hint behavior: prior successful parser is attempted first for each repo.
- Add table tests for all boundary behavior.
- Always update documentation when behavior, flags, contracts, or workflows change.
- Run `go test ./...` after meaningful changes.
- Avoid coupling events payload construction to local state persistence shape.
- Keep command behavior and docs synchronized as commands graduate from reserved to implemented.
