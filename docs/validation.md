# Validation

Choose validation by change type. There is no single universal validation command for every task.

## General Go Change

```sh
go fmt ./...
go test ./...
go build -o bin/dockyard$(go env GOEXE) ./cmd/dockyard
```

## CI-Style Local Gate

```sh
make verify
make build
```

## Package or Example Change

```sh
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

## Runtime Change

Runtime validation may require Docker Desktop or Docker Engine. Keep generated packages, values, rendered files, and archives outside the repository.

## Registry or OCI Change

Run focused unit tests first. Live registry validation requires network and credentials and should be reported explicitly as run or not run.

Historical detail: `.ai/onboarding/reports/validation.md`.
