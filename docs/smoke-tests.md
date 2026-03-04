# Smoke Tests (Docker, Local-First)

Smoke tests validate parser auto-detection and classification against real framework outputs in Docker containers.

## Goals

- No host runtime installs for Python/Node/Go frameworks
- Deterministic assertions on `flake scan --framework auto --json`
- Easy expansion across languages/frameworks

## Layout

- `smoke/manifest.json`: master inventory and expected outputs
- `smoke/cases/<language>/<framework>/<case-id>/`: one Dockerized case
- `smoke/scripts/run.sh`: run one or many cases
- `smoke/scripts/run-case.sh`: execute one case + assertions

Current seeded cases:

- `go/go-test/basic-fail-pass`
- `python/pytest/basic-fail-pass`
- `javascript/jest/basic-fail-pass`
- `java/surefire/basic-fail-pass`
- `java/gradle/basic-fail-pass`
- `dotnet/trx/basic-fail-pass`
- `javascript/mocha/basic-fail-pass`
- `javascript/tap/basic-fail-pass`
- `python/unittest-junitxml/basic-fail-pass`

## Commands

List registered smoke cases:

```bash
make smoke-list
```

Run active `pr` tier cases (default):

```bash
make smoke
```

Run all active tiers:

```bash
make smoke-all
```

Run one case:

```bash
make smoke-case CASE=python-pytest-basic-fail-pass
```

Keep temp artifacts for debugging:

```bash
SMOKE_KEEP=1 make smoke-case CASE=go-go-test-basic-fail-pass
```

## Manifest Contract

Each case in `smoke/manifest.json` defines:

- `id`
- `language`
- `framework`
- `case`
- `tier` (`pr|nightly|extended`)
- `status` (`active|wip|broken|disabled`)
- `path`
- `expected.framework`
- `expected.class_counts`

## CI Policy

Smoke is not run on push or PR.

Manual run only via GitHub Actions:

- workflow: `.github/workflows/smoke.yml`
- trigger: `workflow_dispatch`
- optional inputs: `case_id`, `tiers`
