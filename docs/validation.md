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

For a public catalog/package smoke test, keep downloaded archives outside the repository:

```powershell
New-Item -ItemType Directory -Force ..\dockyard-artifacts | Out-Null
dockyard catalog list
dockyard pull redis -o ..\dockyard-artifacts\redis-0.2.0.dockyard.tgz
dockyard verify ..\dockyard-artifacts\redis-0.2.0.dockyard.tgz
dockyard package test redis --strict
```

This verifies catalog retrieval, catalog shorthand resolution, OCI package pull, archive verification, package quality checks, render, policy checks, and `docker compose config`.

Historical detail: `.ai/onboarding/reports/validation.md`.
