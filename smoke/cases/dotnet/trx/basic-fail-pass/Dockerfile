FROM mcr.microsoft.com/dotnet/sdk:8.0
WORKDIR /app
COPY . .
CMD ["sh", "-c", "dotnet test --logger \"trx;LogFileName=results.trx\" --results-directory TestResults >/tmp/test.log 2>&1 || true; cat TestResults/results.trx"]
