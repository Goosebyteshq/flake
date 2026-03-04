.PHONY: test vet staticcheck lint bench bench-short build-matrix ci-local smoke-build smoke smoke-all smoke-case smoke-list

test:
	go test ./...

vet:
	go vet ./...

staticcheck:
	@command -v staticcheck >/dev/null 2>&1 || (echo "staticcheck not installed" && exit 1)
	@pkgs=$$(go list ./...); \
	test -n "$$pkgs" || (echo "no Go packages found for staticcheck" && exit 1); \
	staticcheck $$pkgs

lint: vet staticcheck

bench:
	go test ./... -run '^$' -bench . -benchmem

bench-short:
	go test ./internal/parsers ./internal/domain ./internal/engine ./internal/state -run '^$' -bench . -benchmem

build-matrix:
	@set -e; \
	mkdir -p .dist-ci; \
	for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		GOOS=$${target%/*}; GOARCH=$${target#*/}; \
		ext=""; \
		if [ "$$GOOS" = "windows" ]; then ext=".exe"; fi; \
		echo "building $$GOOS/$$GOARCH"; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build -trimpath -ldflags="-s -w -buildid=" -o ".dist-ci/flake-$$GOOS-$$GOARCH$$ext" ./cmd/flake; \
	done

ci-local: lint test build-matrix

smoke-build:
	@mkdir -p .dist-smoke
	go build -trimpath -ldflags="-s -w -buildid=" -o .dist-smoke/flake ./cmd/flake

smoke: smoke-build
	@FLAKE_BIN=$(PWD)/.dist-smoke/flake smoke/scripts/run.sh

smoke-all: smoke-build
	@SMOKE_TIERS=pr,nightly,extended FLAKE_BIN=$(PWD)/.dist-smoke/flake smoke/scripts/run.sh

smoke-case: smoke-build
	@test -n "$(CASE)" || (echo "usage: make smoke-case CASE=<case-id>" && exit 1)
	@CASE=$(CASE) FLAKE_BIN=$(PWD)/.dist-smoke/flake smoke/scripts/run.sh

smoke-list:
	@smoke/scripts/list.sh
