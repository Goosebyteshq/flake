FROM gradle:8.10.2-jdk17
WORKDIR /app
COPY . .
CMD ["sh", "-c", "gradle test --console=plain --no-daemon 2>&1 || true"]
