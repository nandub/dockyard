# Testing

Dockyard testing is primarily Go unit tests plus optional package and runtime validation.

## Unit Tests

```sh
go test ./...
```

Tests are under `internal/**/*_test.go`.

## Local Verification

```sh
make verify
```

Observed Makefile behavior combines tidy check, formatting check, and Go tests.

## Package Validation

When a built `dockyard` binary is available:

```sh
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

## Smoke Tests

Docker smoke tests require Docker Desktop or Docker Engine and should keep generated files outside the repository.

## Gaps

Observed gaps in historical onboarding notes include no coverage gate, limited Docker smoke testing in CI, no live OCI registry round-trip in normal CI, and limited dedicated tests for archive/runner paths.

Historical detail: `.ai/onboarding/reports/testing.md`.
