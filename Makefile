.PHONY: fmt test build tidy

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

test:
	go test ./...

build:
	go build -o bin/dockyard ./cmd/dockyard
