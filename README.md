# flake

`flake` is a deterministic, production-grade Go CLI for detecting and tracking flaky tests across frameworks.

It is designed for CI first and emits stable machine-readable outputs for local workflows and remote ingestion.

## Install

Install latest published version:

```bash
go install github.com/leomorpho/flake/cmd/flake@latest
```

Or install a specific version:

```bash
go install github.com/leomorpho/flake/cmd/flake@v0.1.0
```

Verify:

```bash
flake
```

## 60-Second Quickstart

Run tests and scan:

```bash
<your test command> 2>&1 | flake scan --framework auto --state .flake-state.json --json
```

See current status:

```bash
flake report --state .flake-state.json
```

Optional CI gate:

```bash
<your test command> 2>&1 | flake scan --framework auto --state .flake-state.json --fail-on broken,newfail
```

## Copy-Paste by CI

Ready-to-copy files live in [`ci/snippets`](./ci/snippets):

- [GitHub Actions snippet](./ci/snippets/github-actions.yml)
- [GitLab CI snippet](./ci/snippets/gitlab-ci.yml)
- [Jenkins snippet](./ci/snippets/Jenkinsfile.snippet)
- [CircleCI snippet](./ci/snippets/circleci-config.yml)

### GitHub Actions

```yaml
- name: Run tests + flake scan
  run: |
    <your test command> 2>&1 | tee test-output.txt
    flake scan --framework auto --input test-output.txt --state .flake-state.json --events flake-events.json --fail-on broken,newfail
```

### GitLab CI

```yaml
script:
  - <your test command> 2>&1 | tee test-output.txt
  - flake scan --framework auto --input test-output.txt --state .flake-state.json --events flake-events.json --fail-on broken,newfail
```

### Jenkins

```groovy
sh '<your test command> 2>&1 | tee test-output.txt'
sh 'flake scan --framework auto --input test-output.txt --state .flake-state.json --events flake-events.json --fail-on broken,newfail'
```

### CircleCI

```yaml
steps:
  - run: <your test command> 2>&1 | tee test-output.txt
  - run: flake scan --framework auto --input test-output.txt --state .flake-state.json --events flake-events.json --fail-on broken,newfail
```

## Getting Started

Default usage in CI:

```bash
<your test command> 2>&1 | flake scan --framework auto --state .flake-state.json --json
```

`flake` auto-detects parser format and remembers the successful parser for future runs in the same repo.

### Framework Keys

These are the core parser keys supported by `--framework`:

- `go`
- `pytest`
- `jest`
- `junitxml`
- `tap`
- `cargo`
- `trx`
- `surefire`
- `gradle`
- `nunitxml`
- `mocha`

### Integration Methods

Use one of these methods based on how your CI emits test results.

#### 1. Native Text Output (No Extra Reporter)

Use `--framework auto` and pipe stdout/stderr directly.

```bash
<test command> 2>&1 | flake scan --framework auto --state .flake-state.json --json
```

Covers:

- Go (`go test`)
- Pytest text output
- Jest/Vitest/Playwright/Jasmine/bun test text output
- Mocha/Cypress text output
- Cargo/nextest text output
- Maven Surefire/Failsafe/TestNG text output
- Gradle/Kotest/sbt text output

#### 2. JUnit XML Output

If your framework emits JUnit XML, point `flake` at that file.

```bash
<test command that writes junit.xml>
flake scan --framework junitxml --input junit.xml --state .flake-state.json --json
```

Common CI flags/examples:

- Pytest: `pytest --junitxml=junit.xml`
- PHPUnit/Pest: `--log-junit junit.xml`
- Newman: `newman run ... -r junit --reporter-junit-export junit.xml`
- Cucumber: enable JUnit formatter/output file
- GoogleTest: `--gtest_output=xml:junit.xml`
- CTest: `ctest --output-junit junit.xml`
- TestCafe: `-r junit:junit.xml`

Frameworks covered via `junitxml` include:

- WebdriverIO, Selenium, TestCafe
- Cucumber, Robot Framework, Appium, Appium+Cucumber, Behave, Karate, JBehave
- RSpec, Minitest, ExUnit
- ScalaTest, specs2, Spock
- GoogleTest, CTest, Catch2, XCTest
- Dart test, Flutter test, clojure.test, hspec
- PHPUnit, Pest, Newman

#### 3. TRX Output (.NET)

```bash
dotnet test --logger \"trx;LogFileName=results.trx\"
flake scan --framework trx --input results.trx --state .flake-state.json --json
```

Covers:

- `dotnet test`
- xUnit (.trx)
- MSTest (.trx)
- SpecFlow / Reqnroll (.trx)

#### 4. NUnit XML Output

```bash
<test command that writes nunit.xml>
flake scan --framework nunitxml --input nunit.xml --state .flake-state.json --json
```

Covers:

- NUnit console / NUnit XML
- Pester in NUnit XML mode

#### 5. TAP Output

```bash
<test command that writes tap.out>
flake scan --framework tap --input tap.out --state .flake-state.json --json
```

Covers:

- Node test runner (TAP reporter)
- Deno test (TAP mode)
- Perl `prove`
- `bats`
- node-tap style output

Compatibility fixture coverage for all mapped ecosystems lives in `internal/parsers/compatibility_matrix_test.go`.

## Common Stack Recipes

### Ruby on Rails (RSpec)

```bash
bundle exec rspec --format progress --format RspecJunitFormatter --out junit.xml
flake scan --framework junitxml --input junit.xml --state .flake-state.json --json
```

### Ruby on Rails (Minitest with JUnit XML reporter)

```bash
<run minitest with junit xml output to junit.xml>
flake scan --framework junitxml --input junit.xml --state .flake-state.json --json
```

### Laravel (PHPUnit / Pest)

```bash
php artisan test --log-junit junit.xml
flake scan --framework junitxml --input junit.xml --state .flake-state.json --json
```

### Node + Jest / Vitest

```bash
npm test 2>&1 | flake scan --framework auto --state .flake-state.json --json
```

### Python + Pytest

```bash
pytest -q 2>&1 | flake scan --framework auto --state .flake-state.json --json
```

### .NET (TRX)

```bash
dotnet test --logger "trx;LogFileName=results.trx"
flake scan --framework trx --input results.trx --state .flake-state.json --json
```

## What It Does

- Scans test output from multiple frameworks
- Auto-detects parser format and remembers successful parser per repo
- Maintains per-test sliding window history
- Classifies test health and detects state transitions
- Produces stable JSON scan/report outputs and publish-ready events payloads
- Sends filtered Slack notifications without breaking CI

## Core Commands

- `flake scan` with `--json`, `--events`, `--run-meta`, `--debug`
- `flake scan --fail-on <broken|flaky|newfail>` CI gating
- `flake report` with `--json`, `--view`, `--limit`, `--min-fail-streak`
- `flake notify` with deterministic filtering, batching, and suppression

## Docs

- [CI Integration Recipes](./docs/ci-integration.md)
- [Examples](./examples/README.md)
- [Framework Support](./docs/framework-support.md)
- [Configuration](./docs/configuration.md)
- [CLI Contract](./docs/cli-contract.md)
- [Architecture](./docs/architecture.md)
- [Extending Parsers](./docs/extending-parsers.md)
- [Releasing](./docs/releasing.md)
