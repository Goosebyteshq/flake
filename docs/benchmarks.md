# Benchmarks

`flake` includes benchmark suites for parser, domain, engine, state, and end-to-end scan paths.

## Run Short Benchmarks

```bash
make bench-short
```

## Run Full Benchmarks

```bash
make bench
```

## Size Profiles

Benchmarks include these dataset sizes:

- `n=1000`
- `n=10000`
- `n=50000`

These profiles are intended to catch performance regressions in realistic CI-scale workloads.
