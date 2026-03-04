# Extending Parsers

Goal: adding a new parser should be a low-friction contribution.

## Minimal Steps

1. Add one new file under `internal/parsers` implementing:
   - `Name() Framework`
   - `Detect(sample []byte) bool`
   - `Parse(io.Reader) ([]TestResult, error)`
2. In the same file, self-register in `init()` using `registerBuiltin(...)` with:
   - aliases (`--framework` synonyms)
   - detect priority (lower number = tried earlier)
3. Add a parser fixture at `testdata/<framework>/sample.*` (`.txt` or `.xml`).
4. Run `go test ./...`.

Parser conformance/auto-detect tests are fixture-driven and will pick up new frameworks automatically from the registry + `sample.*` fixture convention.

## Determinism Rules

- `Detect` must be deterministic.
- `Parse` must only emit normalized statuses: `pass`, `fail`, `skip`.
- Canonical IDs must be non-empty.
- Output ordering must be stable (registry enforces this).

## Parser Hinting

`flake` remembers successful parser selection per repo in `.flake-parser-hints.json`.

- future runs try the remembered parser first
- fallback only happens when the remembered parser fails

This improves reliability and avoids unnecessary parser probing once a repo is known.
