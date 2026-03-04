# dotnet-trx

.NET TRX integration.

## Run

```bash
dotnet test --logger "trx;LogFileName=results.trx"
flake scan --framework trx --input results.trx --state .flake-state.json --json
```
