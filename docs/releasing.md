# Releasing

This project is installable as:

`go install github.com/leomorpho/flake/cmd/flake@<version>`

## Release Checklist

1. Ensure clean tree and passing tests:

```bash
git status --short
go test ./...
```

2. Create and push an annotated tag:

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin main
git push origin v0.1.0
```

Tag push triggers `.github/workflows/release.yml`, which:

- runs lint + full tests
- builds binaries for Linux/macOS/Windows targets
- packages archives per target
- publishes artifacts and `SHA256SUMS` to the GitHub Release

Tag push also triggers `.github/workflows/ci.yml` (`lint -> test -> build`).
Build jobs depend on tests, so if tests fail, no CI build artifacts are produced.

3. Verify install from a separate repo (or clean shell):

```bash
go install github.com/leomorpho/flake/cmd/flake@v0.1.0
flake
```

## Notes

- Keep semantic version tags (`vMAJOR.MINOR.PATCH`).
- `cmd/flake` is the install target; the module root is not a command.
