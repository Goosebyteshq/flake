FROM golang:1.24
ENV PATH="/usr/local/go/bin:${PATH}"
WORKDIR /app
COPY . .
CMD ["sh", "-c", "go test -v ./... 2>&1 || true"]
