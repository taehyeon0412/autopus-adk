BINARY := auto
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X github.com/anthropics/autopus-adk/pkg/version.version=$(VERSION) -X github.com/anthropics/autopus-adk/pkg/version.commit=$(COMMIT) -X github.com/anthropics/autopus-adk/pkg/version.date=$(DATE)"

.PHONY: build test lint clean install

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/auto

test:
	go test -race -count=1 ./...

lint:
	go vet ./...

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

clean:
	rm -rf bin/ coverage.out

install: build
	cp bin/$(BINARY) $(GOPATH)/bin/$(BINARY)
