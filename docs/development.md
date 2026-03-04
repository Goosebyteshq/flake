# Development Workflow

## Mandatory Loop

1. Implement change.
2. Add or update tests.
3. Update documentation in `docs/`, `README.md`, and/or `AGENTS.md` when behavior changes.
4. Run full test suite:

```bash
go test ./...
```

Run lint gates locally before PR:

```bash
make lint
```

Run local CI-equivalent gate:

```bash
make ci-local
```

Run Docker smoke cases locally (manual only):

```bash
make smoke
```

Run parser fuzz smoke (recommended for parser changes):

```bash
go test ./internal/parsers -run '^$' -fuzz 'FuzzRegistry(ParseAutoDeterministic|DetectAutoDeterministic)$' -fuzztime=10s
```

5. Commit only when tests pass.

## Scope Hygiene

- Keep `cmd/flake` thin (wiring only).
- Keep domain logic pure and table-tested.
- Keep side effects in `state` and `notify`.
- Keep output deterministic and stably sorted.
- Never print secrets in debug output (tokens/webhooks/auth headers/query secrets).
