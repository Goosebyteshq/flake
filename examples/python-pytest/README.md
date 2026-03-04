# python-pytest

Pytest text output integration.

## Run

```bash
pytest -q 2>&1 | flake scan --framework auto --state .flake-state.json --json
```

## XML Variant

```bash
pytest --junitxml=junit.xml
flake scan --framework junitxml --input junit.xml --state .flake-state.json --json
```
