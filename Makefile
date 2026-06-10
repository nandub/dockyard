.PHONY: all fmt fmt-check tidy tidy-check test build dev-build verify clean release-snapshot

GOEXE := $(shell go env GOEXE)
BINARY := bin/dockyard$(GOEXE)
VERSION ?= dev

ifeq ($(OS),Windows_NT)
NULLDEV := NUL
else
NULLDEV := /dev/null
endif

COMMIT ?= $(shell git rev-parse --short HEAD 2>$(NULLDEV))
ifeq ($(strip $(COMMIT)),)
COMMIT := unknown
endif

DATE ?= $(shell git show -s --format=%%cI HEAD 2>$(NULLDEV))
ifeq ($(strip $(DATE)),)
DATE := unknown
endif

LDFLAGS := -s -w \
	-X github.com/nandub/dockyard/internal/version.Version=$(VERSION) \
	-X github.com/nandub/dockyard/internal/version.Commit=$(COMMIT) \
	-X github.com/nandub/dockyard/internal/version.Date=$(DATE)

all: build

fmt:
	go fmt ./...

tidy:
	go mod tidy

test:
	go test ./...

fmt-check:
	go run ./tools/fmtcheck ./cmd ./internal ./tools

tidy-check:
	go mod tidy -diff

dev-build: tidy fmt build

verify: tidy-check fmt-check test

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/dockyard

ifeq ($(OS),Windows_NT)
release-snapshot:
	set GOOS=windows&& set GOARCH=amd64&& go build -ldflags "$(LDFLAGS)" -o bin/dockyard-windows-amd64.exe ./cmd/dockyard
	set GOOS=linux&& set GOARCH=amd64&& go build -ldflags "$(LDFLAGS)" -o bin/dockyard-linux-amd64 ./cmd/dockyard
	set GOOS=linux&& set GOARCH=arm64&& go build -ldflags "$(LDFLAGS)" -o bin/dockyard-linux-arm64 ./cmd/dockyard
	set GOOS=darwin&& set GOARCH=amd64&& go build -ldflags "$(LDFLAGS)" -o bin/dockyard-darwin-amd64 ./cmd/dockyard
	set GOOS=darwin&& set GOARCH=arm64&& go build -ldflags "$(LDFLAGS)" -o bin/dockyard-darwin-arm64 ./cmd/dockyard
else
release-snapshot:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/dockyard-windows-amd64.exe ./cmd/dockyard
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/dockyard-linux-amd64 ./cmd/dockyard
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/dockyard-linux-arm64 ./cmd/dockyard
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/dockyard-darwin-amd64 ./cmd/dockyard
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/dockyard-darwin-arm64 ./cmd/dockyard
endif

clean:
ifeq ($(OS),Windows_NT)
	if exist bin rmdir /s /q bin
else
	rm -rf bin
endif
