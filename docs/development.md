# Development Workflow

Observed local development uses Go, the Makefile, Docker-dependent commands for runtime paths, and registry-dependent commands for OCI paths.

## Bootstrap

Common setup:

```sh
go mod tidy
make dev-build
```

`make dev-build` may modify module and formatting state because it runs tidy and format steps.

## Daily Commands

```sh
go fmt ./...
go test ./...
go build -o bin/dockyard$(go env GOEXE) ./cmd/dockyard
```

Make targets:

```sh
make fmt
make test
make build
make verify
```

## Local Artifacts

Keep generated packages, deployment values, rendered files, and archives outside the repository, for example:

```text
../dockyard-work/
../deploy-values/
../dockyard-artifacts/
```

Historical command comparison: `.ai/onboarding/reports/development.md`.
