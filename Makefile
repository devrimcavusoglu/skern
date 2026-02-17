BINARY    := scribe
MODULE    := github.com/devrimcavusoglu/scribe
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE      ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS   := -s -w \
	-X '$(MODULE)/internal/cli.Version=$(VERSION)' \
	-X '$(MODULE)/internal/cli.Commit=$(COMMIT)' \
	-X '$(MODULE)/internal/cli.Date=$(DATE)'

.PHONY: build test test-v test-cover test-install test-smoke test-manual-setup test-manual-report test-manual-teardown lint fmt clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/scribe

test:
	go test ./...

test-install:
	bash scripts/install_test.sh

test-smoke: build
	bash scripts/smoke_test.sh ./scribe

test-manual-setup: build
	bash tests/manual/setup.sh

test-manual-report:
	bash tests/manual/report.sh

test-manual-teardown:
	bash tests/manual/teardown.sh

test-v:
	go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY) coverage.out coverage.html
