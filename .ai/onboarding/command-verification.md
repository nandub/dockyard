# Command Verification

Commands observed in current docs, Makefile, workflows, or onboarding reports.

## Local Go Commands

```sh
go mod tidy
go fmt ./...
go test ./...
go build -o bin/dockyard$(go env GOEXE) ./cmd/dockyard
```

## Make Targets

```sh
make tidy
make tidy-check
make fmt
make fmt-check
make test
make verify
make build
make dev-build
make clean
```

## Package Validation

```sh
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

## Notes

- Docker and registry commands require the relevant local runtime or credentials.
- Historical command details are preserved in `.ai/onboarding/reports/build.md`, `.ai/onboarding/reports/ci.md`, and `.ai/onboarding/reports/validation.md`.
