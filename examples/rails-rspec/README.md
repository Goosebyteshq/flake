# rails-rspec

Minimal RSpec-style output integration (Rails-oriented).

## Run

```bash
bundle exec rspec --format progress 2>&1 | flake scan --framework auto --state .flake-state.json --json
```

## XML Variant (recommended in larger CI setups)

```bash
bundle exec rspec --format RspecJunitFormatter --out junit.xml
flake scan --framework junitxml --input junit.xml --state .flake-state.json --json
```
