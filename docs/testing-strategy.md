# Testing Strategy

All core behavior is validated through deterministic unit and table tests.

## Required Coverage

- Config YAML parsing/validation/fallback behavior
- Domain classification boundaries and recovering behavior
- Sliding window append/truncate and missing-test no-op
- Transition detection and stable sort order
- Notify suppression metadata behavior
- State load/save roundtrip and schema handling
- Atomic write behavior (best-effort filesystem test)
- Parser fixtures for all supported frameworks
- Fixture convention: `testdata/<framework>/sample.*` to auto-participate in conformance/auto-detect tests
- Parser conformance harness invariants
- Parser fuzz safety/determinism checks (`FuzzRegistryParseAutoDeterministic`, `FuzzRegistryDetectAutoDeterministic`)
- RunMeta env mapping and file override behavior
- Events payload ordering and golden files
- CLI golden tests for summary/report outputs
- End-to-end determinism: same input/state/config => identical outputs and state

## Execution Rule

Run after meaningful changes:

```bash
go test ./...
```

Optional parser hardening pass:

```bash
go test ./internal/parsers -run '^$' -fuzz 'FuzzRegistry(ParseAutoDeterministic|DetectAutoDeterministic)$' -fuzztime=10s
```
