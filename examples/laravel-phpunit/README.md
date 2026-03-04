# laravel-phpunit

Laravel/PHPUnit integration using JUnit XML.

## Run

```bash
php artisan test --log-junit junit.xml
flake scan --framework junitxml --input junit.xml --state .flake-state.json --json
```
