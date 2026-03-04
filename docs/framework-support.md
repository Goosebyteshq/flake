# Framework Support

`flake` supports direct parsing plus broad compatibility via shared report formats.

## Direct v1 Parsers

| Language | Framework | Parser Key |
| --- | --- | --- |
| Go | go | `go` |
| Python | pytest | `pytest` |
| JavaScript/TypeScript | jest | `jest` |
| Java/Kotlin/Scala (XML emit) | junitxml | `junitxml` |
| Multi-language (TAP emit) | tap | `tap` |
| Rust | cargo | `cargo` |
| .NET | trx | `trx` |
| Java (Maven Surefire text) | surefire | `surefire` |
| Java (Gradle text) | gradle | `gradle` |
| .NET (NUnit XML emit) | nunitxml | `nunitxml` |
| JavaScript/TypeScript | mocha | `mocha` |

Operational integration guidance is in [README Getting Started](../README.md#getting-started).

## High-Coverage Strategy

Most ecosystems can emit either JUnit XML, TAP, or TRX. This gives immediate coverage beyond native parsers, including common CI integrations for Java, Kotlin, Scala, JavaScript, Python, Ruby, PHP, and C#.

## Compatibility Fixtures (Unit-Tested)

The following frameworks are validated via compatibility fixtures and routed to existing parsers:

- Vitest -> `jest`
- Playwright Test -> `jest`
- Jasmine -> `jest`
- Cypress -> `mocha`
- Python unittest -> `pytest`
- Python nose2 -> `pytest`
- tox (pytest output) -> `pytest`
- JUnit 5 console -> `jest`
- TestNG console -> `surefire`
- Kotest (Gradle output) -> `gradle`
- sbt test output -> `gradle`
- xUnit (.trx emit) -> `trx`
- NUnit console/XML emit -> `nunitxml`
- MSTest (.trx emit) -> `trx`
- RSpec (JUnit XML emit) -> `junitxml`
- Minitest (JUnit XML emit) -> `junitxml`
- PHPUnit (JUnit XML emit) -> `junitxml`
- Pest (JUnit XML emit) -> `junitxml`
- cargo-nextest output -> `cargo`
- ExUnit (JUnit XML emit) -> `junitxml`
- GoogleTest (JUnit XML emit) -> `junitxml`
- CTest (JUnit XML emit) -> `junitxml`
- Catch2 (JUnit XML emit) -> `junitxml`
- XCTest (JUnit XML emit) -> `junitxml`
- Dart test (JUnit XML emit) -> `junitxml`
- Flutter test (JUnit XML emit) -> `junitxml`
- clojure.test (JUnit XML emit) -> `junitxml`
- hspec (JUnit XML emit) -> `junitxml`
- Perl prove (TAP) -> `tap`
- bats (TAP) -> `tap`
- WebdriverIO (JUnit XML emit) -> `junitxml`
- Selenium (JUnit XML emit) -> `junitxml`
- TestCafe (JUnit XML emit) -> `junitxml`
- Node test runner (TAP) -> `tap`
- bun test -> `jest`
- deno test (TAP mode) -> `tap`
- Cucumber (JUnit XML emit) -> `junitxml`
- Robot Framework (JUnit XML emit) -> `junitxml`
- Spock (JUnit XML emit) -> `junitxml`
- Maven Failsafe text -> `surefire`
- Appium (JUnit XML emit) -> `junitxml`
- Karate (JUnit XML emit) -> `junitxml`
- SpecFlow/Reqnroll (.trx emit) -> `trx`
- Behave (JUnit XML emit) -> `junitxml`
- ScalaTest (JUnit XML emit) -> `junitxml`
- specs2 (JUnit XML emit) -> `junitxml`
- JBehave (JUnit XML emit) -> `junitxml`
- Pester (NUnit XML emit) -> `nunitxml`
- Newman (JUnit XML emit) -> `junitxml`
- Appium+Cucumber (JUnit XML emit) -> `junitxml`

See grouped integration guidance in [README](../README.md#getting-started).

## Auto-Detect Behavior

Detection is deterministic and priority-ordered.

1. Try previously successful parser hint for this repo.
2. If that parser fails, rank matching parsers by deterministic confidence score and try highest confidence first.
3. In auto mode, if multiple parsers succeed, merge successful normalized results deterministically.
4. If no parser succeeds, return parse error.

Parser hints are persisted locally in `.flake-parser-hints.json`.

## Current Gaps

Some frameworks still require either:

- native parser implementation, or
- exporting results as JUnit XML / TAP / TRX

See [Extending Parsers](./extending-parsers.md) to add support with minimal changes.
