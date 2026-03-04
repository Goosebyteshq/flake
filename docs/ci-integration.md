# CI Integration Recipes

This guide shows practical integration patterns for `flake` in CI.

## GitHub Actions

```yaml
name: tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Run Tests and Scan Flakiness
        run: |
          go test ./... 2>&1 | tee test-output.txt
          cat test-output.txt | ./flake scan --framework auto --state .flake-state.json --events flake-events.json --fail-on broken,newfail

      - name: Upload Flake Artifacts
        uses: actions/upload-artifact@v4
        with:
          name: flake-artifacts
          path: |
            .flake-state.json
            flake-events.json

      - name: Notify (non-blocking)
        if: always()
        run: |
          ./flake notify --state .flake-state.json --config .flake-config.yaml
```

## GitLab CI

```yaml
stages:
  - test

test:
  image: golang:1.24
  script:
    - go test ./... 2>&1 | tee test-output.txt
    - cat test-output.txt | ./flake scan --framework auto --state .flake-state.json --events flake-events.json --fail-on broken,newfail
    - ./flake report --state .flake-state.json
    - ./flake notify --state .flake-state.json --config .flake-config.yaml || true
  artifacts:
    when: always
    paths:
      - .flake-state.json
      - flake-events.json
```

## Jenkins (Declarative Pipeline)

```groovy
pipeline {
  agent any
  stages {
    stage('Test + Flake Scan') {
      steps {
        sh 'go test ./... 2>&1 | tee test-output.txt'
        sh 'cat test-output.txt | ./flake scan --framework auto --state .flake-state.json --events flake-events.json --fail-on broken,newfail'
      }
    }
    stage('Report') {
      steps {
        sh './flake report --state .flake-state.json'
      }
    }
    stage('Notify') {
      steps {
        sh './flake notify --state .flake-state.json --config .flake-config.yaml || true'
      }
    }
  }
  post {
    always {
      archiveArtifacts artifacts: '.flake-state.json,flake-events.json', allowEmptyArchive: true
    }
  }
}
```

## Operational Notes

- `notify` never fails CI by design.
- `--fail-on` is the intended CI gate.
  - when tripped, stderr includes deterministic diagnostics (`class`, `failure_rate`, `test_id`, and reason codes) to speed triage
- `--events` gives a stable machine-readable artifact for downstream ingestion.
- Use `--run-meta` to attach explicit run context when environment auto-detection is insufficient.
- `publish` sends deterministic transport headers for remote dedupe/trace:
  - `X-Flake-Idempotency-Key` = SHA-256 of payload
  - `X-Flake-Run-ID` = run id from payload when available
- Docker smoke integration lane is available as manual workflow only (`.github/workflows/smoke.yml`).
